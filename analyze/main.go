package main

import (
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
	packageName := "code.toastmobile.com/logiraptor/catchup-server/routers"
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
	log.Printf("Processing package: %s", pkg.Name)

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

	// Find all the methods
	var methods []*ast.FuncDecl
	ast.Inspect(pkg, func(node ast.Node) bool {
		if d, ok := node.(*ast.FuncDecl); ok && d.Recv != nil {
			methods = append(methods, d)
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

	// log.Println(info)

	var resources []Resource

	for _, node := range nodes {
		spec := node.Specs[0].(*ast.TypeSpec)

		resource := new(Resource)
		resource.Name = spec.Name.String()
		// Get the receiver type
		t, _, err := types.EvalNode(fSet, spec.Name, typePackage, typePackage.Scope())
		if err != nil {
			log.Fatalln(err.Error())
		}

		targetType := t.(*types.Named)

		for _, method := range methods {
			t, _, err = types.EvalNode(fSet, method.Recv.List[0].Type, typePackage, typePackage.Scope())
			if err != nil {
				log.Fatalln(err.Error())
			}
			if !types.Identical(t, targetType) {
				continue
			}

			ast.Inspect(method.Body, func(node ast.Node) bool {
				s, ok := node.(*ast.ReturnStmt)
				if !ok {
					return true
				}

				ast.Inspect(s, func(node ast.Node) bool {
					c, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}

					resp := analyzeNapCall(typePackage, info, c)
					if resp == nil {
						return true
					}

					switch method.Name.String() {
					case "Get":
						resource.Get = append(resource.Get, resp)
					case "Post":
						resource.Post = append(resource.Post, resp)
					case "Put":
						resource.Put = append(resource.Put, resp)
					case "Delete":
						resource.Delete = append(resource.Delete, resp)
					}

					return false

				})

				return false
			})
		}

		resources = append(resources, *resource)
	}

	for _, res := range resources {
		fmt.Println("------", res.Name, "------")
		for _, resp := range res.Get {
			fmt.Println(resp.Describe())
		}

		for _, resp := range res.Post {
			fmt.Println(resp.Describe())
		}

		for _, resp := range res.Put {
			fmt.Println(resp.Describe())
		}

		for _, resp := range res.Delete {
			fmt.Println(resp.Describe())
		}
	}

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
			for typ != typ.Underlying() {
				typ = typ.Underlying()
			}
			ptr, ok := typ.(*types.Pointer)
			for ok {
				typ = ptr.Elem()
				ptr, ok = typ.(*types.Pointer)
			}

			return jsonResponse{typeSpec: typ}
		}
	case "JSONError", "JSONErrorf":
		return errorResponse{}
	}

	return nil
}
