package main

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"

	_ "code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/types"
)

func main() {
	packageName := "github.com/Logiraptor/nap/example"
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

	typePackage, err := types.Check(path, fSet, files)
	if err != nil {
		log.Fatalln(err.Error())
	}

	info := new(types.Info)
	checker := types.NewChecker(&types.Config{}, fSet, typePackage, info)
	log.Println(checker.Info)

	docPackage := doc.New(pkg, path, doc.AllDecls)

	allTypes := docPackage.Types
	for _, typ := range allTypes {
		log.Println(typ.Name)
		log.Println(typ.Doc)
	}

}
