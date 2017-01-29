package controlflow

import (
	"go/ast"
	"go/token"
)

func not(x ast.Expr) ast.Expr {
	return &ast.UnaryExpr{Op: token.NOT, X: x}
}

var (
	astTrue  = ast.NewIdent("true")
	astFalse = ast.NewIdent("false")
)

func makeIf(cond ast.Expr, body []ast.Stmt) ast.Stmt {
	return &ast.IfStmt{Cond: cond, Body: &ast.BlockStmt{List: body}}
}

func makeLoop(body []ast.Stmt) ast.Stmt {
	return &ast.ForStmt{Body: &ast.BlockStmt{List: body}}
}

func define(ident *ast.Ident, value ast.Expr) ast.Stmt {
	return &ast.AssignStmt{Lhs: []ast.Expr{ident}, Tok: token.DEFINE, Rhs: []ast.Expr{value}}
}

func assign(ident *ast.Ident, value ast.Expr) ast.Stmt {
	return &ast.AssignStmt{Lhs: []ast.Expr{ident}, Tok: token.ASSIGN, Rhs: []ast.Expr{value}}
}

func makeBreak() ast.Stmt {
	return &ast.BranchStmt{Tok: token.BREAK}
}

func labeled(label *ast.Ident, stmt ast.Stmt) ast.Stmt {
	return &ast.LabeledStmt{Label: label, Stmt: stmt}
}

func replaceIfBody(stmt *ast.IfStmt, body []ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Init: stmt.Init,
		Cond: stmt.Cond,
		Body: &ast.BlockStmt{List: body},
		Else: stmt.Else,
	}
}

func replaceForBody(stmt *ast.ForStmt, body []ast.Stmt) *ast.ForStmt {
	return &ast.ForStmt{
		Init: stmt.Init,
		Cond: stmt.Cond,
		Post: stmt.Post,
		Body: &ast.BlockStmt{List: body},
	}
}
