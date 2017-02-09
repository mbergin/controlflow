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
func elimOneGoto(stmts []ast.Stmt) []ast.Stmt {
	for i := len(stmts) - 1; i >= 0; i-- {
		stmt := stmts[i]
		var newStmts []ast.Stmt
		if lbl := conditionalGotoLabel(stmt); lbl != nil {
			gotoOffset := i
			condition := stmt.(*ast.IfStmt).Cond
			lblOffset := findLabelOffset(stmts, lbl)
			if lblOffset > 0 {
				labelStmt := stmts[lblOffset].(*ast.LabeledStmt)
				if gotoOffset < lblOffset {
					// goto before label, eliminate with conditional
					newStmts = append(newStmts, stmts[:i]...)
					newStmts = append(newStmts, makeIf(not(condition), stmts[i+1:lblOffset]))
					newStmts = append(newStmts, stmts[lblOffset:]...)
				} else {
					// goto after label, eliminate with loop
					newStmts = append(newStmts, stmts[:lblOffset]...)
					var loopBody []ast.Stmt
					loopBody = append(loopBody, labelStmt.Stmt)
					loopBody = append(loopBody, stmts[lblOffset+1:i]...)
					loopBody = append(loopBody, makeIf(not(condition), []ast.Stmt{makeBreak()}))
					newStmts = append(newStmts, labeled(labelStmt.Label, makeLoop(loopBody)))
					newStmts = append(newStmts, stmts[i+1:]...)
				}
				return newStmts
			}
		}
	}
	return nil
}

func (fs *funcScope) liftGotoIf(stmts []ast.Stmt, postStmts map[*ast.Object]ast.Stmt) []ast.Stmt {
	for {
		var newStmts []ast.Stmt
		for i := len(stmts) - 1; i >= 0; i-- {
			stmt := stmts[i]
			if lbl := conditionalGotoLabel(stmt); lbl != nil {
				condition := stmt.(*ast.IfStmt).Cond
				gotoIdent := fs.tempVar(lbl)
				// remove goto
				gotoStmt := stmt.(*ast.IfStmt).Body.List[0]
				newStmts = append(newStmts, stmts[:i]...)
				if gotoIdent != condition {
					newStmts = append(newStmts, assign(gotoIdent, condition))
				}
				if len(stmts[i+1:]) > 0 {
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
		Body: &ast.BlockStmt{List: fs.liftGotoIf(ifStmt.Body.List, postStmts)},
	}
	if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
		newIfStmt.Else = makeElseBlock(fs.liftGotoIf(elseBlock.List, postStmts))
	}
	stmts := []ast.Stmt{newIfStmt}
	for _, post := range postStmts {
		stmts = append(stmts, post)
	}
	return stmts
}

func (fs *funcScope) elimGotos(stmt ast.Stmt) []ast.Stmt {
	var stmts []ast.Stmt
	switch stmt := stmt.(type) {
	case *ast.IfStmt:
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
	case *ast.BlockStmt:
		var newStmts []ast.Stmt
		for _, bodyStmt := range stmt.List {
			newStmts = append(newStmts, fs.elimGotos(bodyStmt)...)
		}
		stmts = newStmts

		// Eliminate conditional gotos from this block
		for {
			if newStmts2 := elimOneGoto(stmts); newStmts2 != nil {
				stmts = newStmts2
			} else {
				break
			}
		}
		stmts = removeLabels(stmts)
	case *ast.BranchStmt:
		// Unconditional goto. Wrap in "if true { goto L }"
		if stmt.Tok == token.GOTO {
			stmts = fs.elimGotos(&ast.IfStmt{Cond: astTrue, Body: &ast.BlockStmt{List: []ast.Stmt{stmt}}})
		}
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
