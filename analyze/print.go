package main

import (
	"fmt"

	"code.google.com/p/go.tools/go/types"
)

func printTypeWrapper(i interface{}) string {
	typ, ok := i.(types.Type)
	if !ok {
		return "non-type"
	}

	return printType(typ, map[types.Type]struct{}{})
}

func printType(t types.Type, seen map[types.Type]struct{}) string {
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
			resp += field.Name() + " " + printType(typ, seen) + "\n"
		}
		if resp == "" {
			resp += "{}"
		} else {
			resp = "struct {\n" + resp + "}"
		}
	case *types.Named:
		resp += typ.Obj().Pkg().Name() + "." + typ.Obj().Name()
	case *types.Slice:
		resp += "[]" + printType(typ.Elem(), seen) + "\n"
	case *types.Pointer:
		resp += printType(typ.Elem(), seen)
	case *types.Basic:
		resp += typ.String()
	default:
		resp += fmt.Sprintf("%T", typ)
	}
	return resp
}

func getImports(t types.Type, seen map[types.Type]struct{}) []string {
	if _, ok := seen[t]; ok {
		if _, ok := t.(*types.Pointer); ok {
			return nil
		}
	}
	seen[t] = struct{}{}
	resp := []string{}
	switch typ := t.(type) {
	case *types.Struct:
		numField := typ.NumFields()
		for i := 0; i < numField; i++ {
			field := typ.Field(i)
			if !field.Exported() {
				continue
			}
			typ := field.Type()
			resp = append(resp, getImports(typ, seen)...)
		}
	case *types.Named:
		resp = append(resp, typ.Obj().Pkg().Path())
	case *types.Slice:
		resp = append(resp, getImports(typ.Elem(), seen)...)
	case *types.Pointer:
		resp = append(resp, getImports(typ.Elem(), seen)...)
	case *types.Basic:
	default:
	}
	return resp
}
