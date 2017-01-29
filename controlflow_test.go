package controlflow

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "./testdata", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg := pkgs["testdata"]
	tests := map[string]struct {
		input    []ast.Stmt
		expected []ast.Stmt
	}{}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				name := decl.Name.Name
				if strings.HasSuffix(name, "Input") {
					k := strings.TrimSuffix(name, "Input")
					test := tests[k]
					test.input = decl.Body.List
					tests[k] = test
				} else if strings.HasSuffix(name, "Expected") {
					k := strings.TrimSuffix(name, "Expected")
					test := tests[k]
					test.expected = decl.Body.List
					tests[k] = test
				}
			}
		}
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotAST := ElimGotos(test.input)
			want := astString(test.expected, token.NewFileSet())
			got := astString(gotAST, token.NewFileSet())
			if want != got {
				t.Error("\n" + sideBySide("want:", "got:", lines(want), lines(got)))
			}
		})
	}
}

func sideBySide(titleX, titleY string, xs, ys []string) string {
	n := len(xs)
	if len(ys) > n {
		n = len(ys)
	}
	var w bytes.Buffer
	fmt.Fprintf(&w, "%-40s %-40s\n", titleX, titleY)
	for i := 0; i < n; i++ {
		x := ""
		y := ""
		if i < len(xs) {
			x = strings.TrimRight(xs[i], " \r\n\t")
		}
		if i < len(ys) {
			y = strings.TrimRight(ys[i], " \r\n\t")
		}
		fmt.Fprintf(&w, "| %-40s | %-40s\n", x, y)
	}
	return w.String()
}

func astString(node interface{}, fset *token.FileSet) string {
	var buf bytes.Buffer
	// Tabs screw up the side-by-side alignment
	config := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}
	config.Fprint(&buf, fset, node)
	return buf.String()
}

func lines(s string) []string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t") + "\n"
	}
	return lines
}
