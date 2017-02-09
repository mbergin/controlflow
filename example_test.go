package controlflow

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func ExampleElimGotos() {
	code := `
package branchy
func isGoodNumber(x int) bool {

	if x%3 == 0 {
		goto fail
	} else if x%5 == 0 {
		goto fail
	}
    return true
fail:
	return false	
}
`

	// Parse the code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "code.go", code, 0)
	if err != nil {
		panic(err)
	}

	// List of statements in the function's body
	body := file.Decls[0].(*ast.FuncDecl).Body.List

	// Rewrite the function body without goto statements
	newBody := ElimGotos(body)

	// Print the result
	config := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}
	config.Fprint(os.Stdout, token.NewFileSet(), newBody)
	// Output:
	// gotofail := false
	// if x%3 == 0 {
	//     gotofail = true
	// } else {
	//     gotofail = x%5 == 0
	// }
	// if !gotofail {
	//     return true
	// }
	// return false

	if err != nil {
		panic(err)
	}

}
