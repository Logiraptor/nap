package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"code.google.com/p/go.tools/go/exact"

	"code.google.com/p/go.tools/go/types"
)

func analyzeMethod(fSet *token.FileSet, typePackage *types.Package, info *types.Info, method *ast.FuncDecl, resource *Resource) {
	ast.Inspect(method.Body, func(node ast.Node) bool {
		// fmt.Printf("%[1]T: %[1]v\n", node)

		c, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		m, ok := resource.Methods[method.Name.String()]
		if !ok {
			m = new(Method)
			m.Params = make(map[string]struct{})
		}
		defer func() { resource.Methods[method.Name.String()] = m }()

		param, ok := analyzeNapParam(typePackage, info, c)
		if ok {
			m.Params[param] = struct{}{}
			return true
		}

		resp := analyzeNapCall(typePackage, info, c)
		if resp == nil {
			return true
		}

		m.Doc = method.Doc.Text()
		for _, r := range m.Responses {
			if r == resp {
				return true
			}
		}
		m.Responses = append(m.Responses, resp)

		return true

	})
}

func analyzeNapParam(pkg *types.Package, info *types.Info, c *ast.CallExpr) (string, bool) {
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	xType := info.Types[sel.X]
	if xType.Type == nil {
		return "", false
	}

	if xType.Type.String() != "github.com/Logiraptor/nap.Request" {
		return "", false
	}
	// printNode(c)

	nameTV := info.Types[c.Args[0]]

	return exact.StringVal(nameTV.Value), true
}

func printNode(n ast.Node) {
	ast.Inspect(n, func(n ast.Node) bool {
		fmt.Printf("%T: %v\n", n, n)

		return true
	})
}
