package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.google.com/p/go.tools/go/types"
)

type Method struct {
	Doc       string
	Params    map[string]struct{}
	Responses []Response
}

func (m *Method) MarshalJSON() ([]byte, error) {
	type resp struct {
		Schema  string
		Example string
	}
	var resps []resp

	for _, r := range m.Responses {
		resps = append(resps, resp{
			Schema:  r.Describe(),
			Example: r.Example(),
		})
	}

	var params = []string{}
	for k := range m.Params {
		params = append(params, k)
	}

	return json.Marshal(map[string]interface{}{
		"Doc":       m.Doc,
		"Params":    params,
		"Responses": resps,
	})
}

// Resource describes an exposed rest resource
type Resource struct {
	Name    string
	Doc     string
	URL     string
	Methods map[string]*Method
}

// Response describes a potential rest response
type Response interface {
	Describe() string
	Example() string
}

type successResponse struct{}

func (s successResponse) Describe() string {
	return "{\n\tStatus: string\n}"
}

func (s successResponse) Example() string {
	return "{\n\t\"Status\": \"Success\"\n}"
}

type errorResponse struct{}

func (e errorResponse) Describe() string {
	return "{\n\tError: string\n\tCode: int\n}"
}

func (e errorResponse) Example() string {
	return "{\n\t\"Error\": \"You must log in to do that\"\n\t\"Code\": 405\n}"
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

func (j jsonResponse) Example() string {
	return getOutput(getJSON, map[string]interface{}{
		"Type":    j.typeSpec,
		"Imports": getImports(j.typeSpec, map[types.Type]struct{}{}),
	})
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
