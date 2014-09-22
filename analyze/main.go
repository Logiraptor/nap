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

type outputResponse struct {
	Schema  string
	Example string
}

type outputResource struct {
	Name string
	Doc  string
	Get  struct {
		Responses []outputResponse
		Doc       string
	}
	Put struct {
		Responses []outputResponse
		Doc       string
	}
	Post struct {
		Responses []outputResponse
		Doc       string
	}
	Delete struct {
		Responses []outputResponse
		Doc       string
	}
}

type outputFormat struct {
	Resources []outputResource
}

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

	var resources []Resource
	for _, node := range nodes {
		spec := node.Specs[0].(*ast.TypeSpec)

		resource := new(Resource)
		resource.Name = spec.Name.String()
		resource.Doc = node.Doc.Text()
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

				methodSwitch:
					switch method.Name.String() {
					case "Get":
						resource.GetDoc = method.Doc.Text()
						for _, r := range resource.Get {
							if r == resp {
								break methodSwitch
							}
						}
						resource.Get = append(resource.Get, resp)
					case "Post":
						resource.PostDoc = method.Doc.Text()
						for _, r := range resource.Post {
							if r == resp {
								break methodSwitch
							}
						}
						resource.Post = append(resource.Post, resp)
					case "Put":
						resource.PutDoc = method.Doc.Text()
						for _, r := range resource.Put {
							if r == resp {
								break methodSwitch
							}
						}
						resource.Put = append(resource.Put, resp)
					case "Delete":
						resource.DeleteDoc = method.Doc.Text()
						for _, r := range resource.Delete {
							if r == resp {
								break methodSwitch
							}
						}
						resource.Delete = append(resource.Delete, resp)
					}

					return false

				})

				return false
			})
		}

		resources = append(resources, *resource)
	}

	var output outputFormat

	for _, res := range resources {
		var outRes outputResource
		outRes.Name = res.Name
		outRes.Doc = res.Doc
		outRes.Get.Doc = res.GetDoc
		outRes.Post.Doc = res.PostDoc
		outRes.Put.Doc = res.PutDoc
		outRes.Delete.Doc = res.DeleteDoc

		for _, resp := range res.Get {
			outRes.Get.Responses = append(outRes.Get.Responses, outputResponse{
				Schema:  resp.Describe(),
				Example: resp.Example(),
			})
		}

		for _, resp := range res.Post {
			outRes.Post.Responses = append(outRes.Post.Responses, outputResponse{
				Schema:  resp.Describe(),
				Example: resp.Example(),
			})
		}

		for _, resp := range res.Put {
			outRes.Put.Responses = append(outRes.Put.Responses, outputResponse{
				Schema:  resp.Describe(),
				Example: resp.Example(),
			})
		}

		for _, resp := range res.Delete {
			outRes.Delete.Responses = append(outRes.Delete.Responses, outputResponse{
				Schema:  resp.Describe(),
				Example: resp.Example(),
			})
		}
		output.Resources = append(output.Resources, outRes)
	}

	out := os.Stdout
	if len(os.Args) == 3 {
		out, err = os.Create(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	json.NewEncoder(out).Encode(output)
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
