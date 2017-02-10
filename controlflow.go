package controlflow

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

func findLabelOffset(stmts []ast.Stmt, target *ast.Object) int {
	for i, stmt := range stmts {
		if label, ok := stmt.(*ast.LabeledStmt); ok && label.Label.Obj == target {
			return i
		}
	}
	return -1
}

func conditionalGotoLabel(stmt ast.Stmt) *ast.Object {
	if node, ok := stmt.(*ast.IfStmt); ok {
		if node.Else == nil && node.Init == nil && len(node.Body.List) == 1 {
			if branchStmt, ok := node.Body.List[0].(*ast.BranchStmt); ok {
				if branchStmt.Tok == token.GOTO {
					if branchStmt.Label.Obj == nil {
						panic("BranchStmt with no matching Object")
					}
					return branchStmt.Label.Obj
				}
			}
		}
	}
	return nil
}

type funcScope struct {
	labelToTempVar map[*ast.Object]*ast.Ident
}

func newFuncScope() *funcScope {
	return &funcScope{labelToTempVar: map[*ast.Object]*ast.Ident{}}
}

func (fs *funcScope) tempVar(lbl *ast.Object) *ast.Ident {
	if ident, ok := fs.labelToTempVar[lbl]; ok {
		return ident

	}
	ident := ast.NewIdent("goto" + lbl.Name)
	fs.labelToTempVar[lbl] = ident
	return ident
}

// pre: stmts only contains conditional gotos at the top level
// post: stmts only contains conditional gotos for labels defined in an outer scope
func (fs *funcScope) elimSiblings(stmts []ast.Stmt) []ast.Stmt {
	for {
		// Find a goto pointing with a label inside the current block
		gotoOffset := -1
		lblOffset := -1
		for i := range stmts {

			// Is this a conditional goto?
			if lbl := conditionalGotoLabel(stmts[i]); lbl != nil {

				// Find the matching label
				lblOffset = findLabelOffset(stmts, lbl)

				// If the label isn't in this block, it must be in an outer block,
				// so do nothing, but continue searching.
				if lblOffset < 0 {
					continue
				}

				// Finished: this is a goto with a label in this block
				gotoOffset = i
				break
			}

		}

		// Finished if no gotos left
		if gotoOffset == -1 {
			break
		}

		// Eliminate this goto, but leave the label in case another goto points to it.
		var newStmts []ast.Stmt
		condition := stmts[gotoOffset].(*ast.IfStmt).Cond
		labelStmt := stmts[lblOffset].(*ast.LabeledStmt)
		if gotoOffset < lblOffset {
			// goto before label, eliminate with conditional
			newStmts = append(newStmts, stmts[:gotoOffset]...)
			newStmts = append(newStmts, fs.elimGotos(makeIf(not(condition), stmts[gotoOffset+1:lblOffset]))...)
			newStmts = append(newStmts, stmts[lblOffset:]...)
		} else {
			// goto after label, eliminate with loop
			newStmts = append(newStmts, stmts[:lblOffset]...)
			var loopBody []ast.Stmt
			loopBody = append(loopBody, labelStmt.Stmt)
			loopBody = append(loopBody, stmts[lblOffset+1:gotoOffset]...)
			loopBody = append(loopBody, makeIf(not(condition), []ast.Stmt{makeBreak()}))
			newStmts = append(newStmts, labeled(labelStmt.Label, &ast.EmptyStmt{}))
			newStmts = append(newStmts, fs.elimGotos(makeLoop(loopBody))...)
			newStmts = append(newStmts, stmts[gotoOffset+1:]...)
		}
		stmts = newStmts
	}
	return stmts
}

func (fs *funcScope) liftGoto(stmts []ast.Stmt, postStmts map[*ast.Object]ast.Stmt, useBreak bool) []ast.Stmt {
	for {
		var newStmts []ast.Stmt
		for i, stmt := range stmts {
			if lbl := conditionalGotoLabel(stmt); lbl != nil {
				condition := stmt.(*ast.IfStmt).Cond
				gotoIdent := fs.tempVar(lbl)
				// remove goto
				gotoStmt := stmt.(*ast.IfStmt).Body.List[0]
				newStmts = append(newStmts, stmts[:i]...)
				if gotoIdent != condition {
					newStmts = append(newStmts, assign(gotoIdent, condition))
				}
				if useBreak {
					// if gotoLabel { break }
					newStmts = append(newStmts, makeIf(gotoIdent, []ast.Stmt{makeBreak()}))
					// copy statements after the goto
					newStmts = append(newStmts, stmts[i+1:]...)
				} else if len(stmts[i+1:]) > 0 {
					newStmts = append(newStmts, makeIf(not(gotoIdent), stmts[i+1:]))
				}

				postStmts[lbl] = makeIf(gotoIdent, []ast.Stmt{gotoStmt})
				break
			}
		}
		if newStmts == nil {
			break
		}
		stmts = newStmts
	}
	return stmts
}

func makeElseBlock(stmts []ast.Stmt) ast.Stmt {
	if len(stmts) == 1 {
		if elseIf, ok := stmts[0].(*ast.IfStmt); ok {
			return elseIf
		}
	}
	return &ast.BlockStmt{List: stmts}
}

// pre: if body contains only gotos that refer to an outer label
// post: if body contains no gotos
func (fs *funcScope) moveGotosOutOfIf(ifStmt *ast.IfStmt) []ast.Stmt {
	postStmts := map[*ast.Object]ast.Stmt{}
	newIfStmt := &ast.IfStmt{
		Init: ifStmt.Init,
		Cond: ifStmt.Cond,
		Body: &ast.BlockStmt{List: fs.liftGoto(ifStmt.Body.List, postStmts, false)},
	}
	if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
		newIfStmt.Else = makeElseBlock(fs.liftGoto(elseBlock.List, postStmts, false))
	}
	stmts := []ast.Stmt{newIfStmt}
	for _, post := range postStmts {
		stmts = append(stmts, post)
	}
	return stmts
}

func (fs *funcScope) moveGotosOutOfFor(forStmt *ast.ForStmt) []ast.Stmt {
	postStmts := map[*ast.Object]ast.Stmt{}
	newForStmt := &ast.ForStmt{
		Init: forStmt.Init,
		Cond: forStmt.Cond,
		Body: &ast.BlockStmt{List: fs.liftGoto(forStmt.Body.List, postStmts, true)},
		Post: forStmt.Post,
	}
	stmts := []ast.Stmt{newForStmt}
	for _, post := range postStmts {
		stmts = append(stmts, post)
	}
	return stmts
}

func (fs *funcScope) moveGotosOutOfRange(rangeStmt *ast.RangeStmt) []ast.Stmt {
	postStmts := map[*ast.Object]ast.Stmt{}
	newRangeStmt := replaceRangeBody(rangeStmt, fs.liftGoto(rangeStmt.Body.List, postStmts, true))
	stmts := []ast.Stmt{newRangeStmt}
	for _, post := range postStmts {
		stmts = append(stmts, post)
	}
	return stmts
}

func (fs *funcScope) elimGotos(stmt ast.Stmt) []ast.Stmt {
	var stmts []ast.Stmt
	switch stmt := stmt.(type) {
	case *ast.IfStmt:
		// Only recurse if this is not a simple conditional goto
		if conditionalGotoLabel(stmt) == nil {
			stmt = replaceIfBody(stmt, fs.elimGotos(stmt.Body))
			elseBlock := stmt.Else
			stmt = &ast.IfStmt{
				Init: stmt.Init,
				Cond: stmt.Cond,
				Body: &ast.BlockStmt{List: fs.elimGotos(stmt.Body)},
			}
			if elseBlock != nil {
				stmt.Else = &ast.BlockStmt{List: fs.elimGotos(elseBlock)}
			}
			// Now body only contains gotos where the label is in an outer scope
			// Move these gotos out one level
			stmts = fs.moveGotosOutOfIf(stmt)
		} else {
			stmts = []ast.Stmt{stmt}
		}
	case *ast.ForStmt:
		stmt = replaceForBody(stmt, fs.elimGotos(stmt.Body))
		stmts = fs.moveGotosOutOfFor(stmt)
	case *ast.RangeStmt:
		stmt = replaceRangeBody(stmt, fs.elimGotos(stmt.Body))
		stmts = fs.moveGotosOutOfRange(stmt)
	case *ast.SwitchStmt:
		newSwitch := &ast.SwitchStmt{
			Init: stmt.Init,
			Tag:  stmt.Tag,
			Body: &ast.BlockStmt{List: make([]ast.Stmt, len(stmt.Body.List))},
		}
		for i, s := range stmt.Body.List {
			cc := s.(*ast.CaseClause)
			tempBlock := &ast.BlockStmt{List: cc.Body}
			newSwitch.Body.List[i] = replaceCaseClauseBody(cc, fs.elimGotos(tempBlock))
		}
		stmts = []ast.Stmt{newSwitch}
	case *ast.BlockStmt:
		var newStmts []ast.Stmt
		for _, bodyStmt := range stmt.List {
			newStmts = append(newStmts, fs.elimGotos(bodyStmt)...)
		}
		stmts = newStmts

		// Eliminate conditional gotos from this block
		stmts = fs.elimSiblings(stmts)

		// Remove the (now unused) labels
		stmts = removeLabels(stmts)
	case *ast.BranchStmt:
		// Unconditional goto. Wrap in "if true { goto L }"
		if stmt.Tok == token.GOTO {
			stmts = fs.elimGotos(&ast.IfStmt{Cond: astTrue, Body: &ast.BlockStmt{List: []ast.Stmt{stmt}}})
		} else {
			stmts = []ast.Stmt{stmt}
		}
	case *ast.LabeledStmt:
		stmts = []ast.Stmt{&ast.LabeledStmt{Label: stmt.Label, Stmt: &ast.EmptyStmt{}}}
		stmts = append(stmts, fs.elimGotos(stmt.Stmt)...)
	default:
		stmts = []ast.Stmt{stmt}
	}

	return stmts
}

// dump numbered statements (for debugging)
func p(stmts []ast.Stmt) {
	for i, stmt := range stmts {
		fmt.Printf("/*%d*/ ", i)
		printer.Fprint(os.Stdout, token.NewFileSet(), stmt)
		fmt.Println()
	}
	fmt.Println()
	fmt.Println()
}

func removeLabels(stmts []ast.Stmt) []ast.Stmt {
	var noLabels []ast.Stmt
	for _, stmt := range stmts {
		for {
			ls, ok := stmt.(*ast.LabeledStmt)
			if !ok {
				break
			}
			stmt = ls.Stmt
		}
		noLabels = append(noLabels, stmt)
	}
	return noLabels
}

// ElimGotos removes any goto statements from stmts by rewriting them as conditionals and loops.
// A transformed syntax tree is returned; stmts is not modified.
func ElimGotos(stmts []ast.Stmt) []ast.Stmt {
	fs := newFuncScope()
	elim := fs.elimGotos(&ast.BlockStmt{List: stmts})
	var newStmts []ast.Stmt
	for _, tempVar := range fs.labelToTempVar {
		newStmts = append(newStmts, define(tempVar, astFalse))
	}
	newStmts = append(newStmts, elim...)
	return newStmts
}
