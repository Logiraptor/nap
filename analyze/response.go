package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"text/template"

	"code.google.com/p/go.tools/go/types"
)

// Resource describes an exposed rest resource
type Resource struct {
	Name   string
	URL    string
	Get    []Response
	Put    []Response
	Post   []Response
	Delete []Response
}

// Response describes a potential rest response
type Response interface {
	Describe() string
}

type errorResponse struct{}

func (e errorResponse) Describe() string {
	return "{\n\tError: string\n}"
}

type jsonResponse struct {
	desc     string
	typeSpec types.Type
}

func (j jsonResponse) Describe() string {
	if j.desc != "" {
		return j.desc
	}

	var imports = map[string]struct{}{}
	fillImports(imports, j.typeSpec)

	return write(j.typeSpec, map[types.Type]struct{}{}, 0)
}

func fillImports(imp map[string]struct{}, t types.Type) {
	switch typ := t.(type) {
	case *types.Struct:
		numField := typ.NumFields()
		for i := 0; i < numField; i++ {
			field := typ.Field(i)
			typ := field.Type()
			fillImports(imp, typ)
		}
	case *types.Named:
		imp[typ.Obj().Pkg().Path()] = struct{}{}
	case *types.Slice:
		fillImports(imp, typ.Elem())
	case *types.Basic:
	default:
		imp[fmt.Sprintf("%T", t)] = struct{}{}
	}
}

func write(t types.Type, seen map[types.Type]struct{}, depth int) string {
	if _, ok := seen[t]; ok {
		if _, ok := t.(*types.Pointer); ok {
			return "<recursive>"
		}
	}
	seen[t] = struct{}{}
	resp := ""
	switch typ := t.(type) {
	case *types.Struct:
		numField := typ.NumFields()
		for i := 0; i < numField; i++ {
			field := typ.Field(i)
			if !field.Exported() {
				continue
			}
			typ := field.Type()
			resp += strings.Repeat("\t", depth+1) + field.Name() + ": " + write(typ, seen, depth+1) + "\n"
		}
		if resp == "" {
			resp += "{}"
		} else {
			resp = "{\n" + resp + strings.Repeat("\t", depth) + "}"
		}
	case *types.Named:
		resp += write(typ.Obj().Type().Underlying(), seen, depth)
	case *types.Slice:
		resp += "[" + write(typ.Elem(), seen, depth) + "]\n"
	case *types.Pointer:
		resp += write(typ.Elem(), seen, depth)
	case *types.Basic:
		resp += typ.String()
	default:
		resp += fmt.Sprintf("%T", typ)
	}
	return resp
}

type namedResponse struct {
	name *types.Named
	desc string
}

func (n namedResponse) Describe() string {
	if n.desc != "" {
		return n.desc
	}

	return "not yet implemented"
}

func getOutput(src string, data interface{}) string {
	fileName := "nap_temp_.go"
	tmpl := template.New("temp")
	tmpl, err := tmpl.Parse(src)
	if err != nil {
		return err.Error()
	}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "temp", data)
	if err != nil {
		return err.Error()
	}
	err = ioutil.WriteFile(fileName, buf.Bytes(), 0666)
	if err != nil {
		return err.Error()
	}

	cmd := exec.Command("go", "run", fileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}

	return string(output)
}
