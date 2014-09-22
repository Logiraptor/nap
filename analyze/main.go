package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"

	_ "code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/types"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: analyze [package name]")
		return
	}

	packageName := os.Args[1]
	goPath := os.Getenv("GOPATH")

	absPath := path.Join(goPath, "src", packageName)

	fSet := token.NewFileSet()
	packages, err := parser.ParseDir(fSet, absPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, v := range packages {
		processPackage(packageName, fSet, v)
	}
}

func processPackage(path string, fSet *token.FileSet, pkg *ast.Package) {
	fmt.Printf("Processing package: %s\n", pkg.Name)

	var files []*ast.File
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	// Find all the types
	var nodes []*ast.GenDecl
	ast.Inspect(pkg, func(node ast.Node) bool {
		if d, ok := node.(*ast.GenDecl); ok && d.Tok == token.TYPE {
			nodes = append(nodes, d)
			return false
		}
		return true
	})

	cfg := &types.Config{}
	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
	}

	typePackage, err := cfg.Check(path, fSet, files, info)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var resources []Resource
	for _, node := range nodes {
		spec := node.Specs[0].(*ast.TypeSpec)

		resource := new(Resource)
		resource.Name = spec.Name.String()
		resource.Doc = node.Doc.Text()
		resource.Methods = make(map[string]*Method)
		// Get the receiver type
		t, _, err := types.EvalNode(fSet, spec.Name, typePackage, typePackage.Scope())
		if err != nil {
			log.Fatalln(err.Error())
		}

		targetType := t.(*types.Named)

		// Find all the methods for targetType
		var methods []*ast.FuncDecl
		ast.Inspect(pkg, func(node ast.Node) bool {
			if d, ok := node.(*ast.FuncDecl); ok {
				if d.Recv == nil {
					return true
				}

				t, _, err := types.EvalNode(fSet, d.Recv.List[0].Type, typePackage, typePackage.Scope())
				if err != nil {
					log.Fatalln(err.Error())
				}

				if types.Identical(t, targetType) {
					methods = append(methods, d)
					return false
				}
				return true
			}
			return true
		})
		for _, method := range methods {
			analyzeMethod(fSet, typePackage, info, method, resource)
		}

		resources = append(resources, *resource)
	}

	out := os.Stdout
	if len(os.Args) == 3 {
		out, err = os.Create(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	json.NewEncoder(out).Encode(map[string]interface{}{
		"Resources": resources,
	})
}

func analyzeNapCall(pkg *types.Package, info *types.Info, c *ast.CallExpr) Response {
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	napName, ok := sel.X.(*ast.Ident)
	if !ok || napName.Name != "nap" {
		return nil
	}

	switch sel.Sel.Name {
	case "JSON":
		tv, ok := info.Types[c.Args[0]]
		if ok {
			typ := tv.Type
			return jsonResponse{typeSpec: typ}
		}
	case "JSONError", "JSONErrorf":
		return errorResponse{}
	case "JSONSuccess":
		return successResponse{}
	}

	return nil
}
