package main

import (
	"go/ast"
	"go/types"
	"log/slog"
)

// Struct represents a `struct` object extracted from Go source code.
// It accumulates information necessary for generating new Go code.
type Struct struct {
	Receiver   string
	Name       string
	TypeParams []string
	Fields     []Field
}

// Field describe a field of a [Struct] with given name and type.
type Field struct {
	Name string
	Type Renderable
}

// Renderable describes a renderable object associated with a template.
type Renderable interface {
	TemplateName() string
}

type (
	// TypeBasic represents basic Go types, e.g. `int`, `string`, `bool`, etc.
	// It provides the type's zero value as a string.
	TypeBasic struct {
		ZeroValue string
	}

	// TypeSlice represents a Go slice with elements of any type.
	TypeSlice struct{}

	// TypeMap represents a Go map with keys/values of any type.
	TypeMap struct{}

	// TypePtr represents a pointer to a field;
	// a pointer's target can be either a pointer or a regular field.
	TypePtr struct {
		Target Renderable
	}

	// TypeStruct represents field of [types.Struct] type.
	TypeStruct struct{}

	// TypeInterface represents field of [types.Interface] type.
	TypeInterface struct{}
)

type typeCreator struct {
	gotypes *types.Info
}

// TemplateName returns template name for rendering [TypeBasic] field.
func (t TypeBasic) TemplateName() string {
	return templateFieldBasic
}

// TemplateName returns template name for rendering [TypeSlice] field.
func (t TypeSlice) TemplateName() string {
	return templateFieldSlice
}

// TemplateName returns template name for rendering [TypeMap] field.
func (t TypeMap) TemplateName() string {
	return templateFieldMap
}

// TemplateName returns template name for rendering [TypePtr] field.
func (t TypePtr) TemplateName() string {
	return templateFieldPtr
}

// TemplateName returns template name for rendering [TypeStruct] field.
func (t TypeStruct) TemplateName() string {
	return templateFieldStruct
}

// TemplateName returns template name for rendering [TypeInterface] field.
func (t TypeInterface) TemplateName() string {
	return templateFieldInterface
}

// FromExpr analyzes AST expression and creates appropriate [Renderable] object.
func (tc *typeCreator) FromExpr(expr ast.Expr) (Renderable, bool) {
	switch value := expr.(type) {
	case *ast.Ident:
		return tc.FromIdent(value)
	case *ast.SelectorExpr:
		return tc.FromIdent(value.Sel)
	case *ast.ArrayType:
		return TypeSlice{}, true
	case *ast.MapType:
		return TypeMap{}, true
	case *ast.StarExpr:
		target, ok := tc.FromExpr(value.X)
		if ok {
			return TypePtr{Target: target}, true
		}
	default:
		slog.Debug("type creator: unknown AST expression", slog.Any("expr", value))
	}

	return nil, false
}

// FromIdent analyzes AST identifier and creates appropriate [Renderable] object.
func (tc *typeCreator) FromIdent(ident *ast.Ident) (Renderable, bool) {
	itype := tc.gotypes.TypeOf(ident)
	if itype == nil {
		slog.Debug("type creator: ident type missing", slog.Any("ident", ident))

		return nil, false
	}

	return tc.FromType(itype)
}

// FromType analyzes AST identifier and creates appropriate [Renderable] object.
func (tc *typeCreator) FromType(ttype types.Type) (Renderable, bool) {
	switch value := ttype.(type) {
	case *types.Basic:
		return tc.createBasic(value)
	case *types.Slice:
		return TypeSlice{}, true
	case *types.Map:
		return TypeMap{}, true
	case *types.Named:
		return tc.FromType(value.Underlying())
	case *types.Struct:
		return TypeStruct{}, true
	case *types.Interface:
		return TypeInterface{}, true
	case *types.Alias:
		return tc.FromType(value.Rhs())
	case *types.TypeParam:
		return tc.FromType(value.Constraint())
	case *types.Pointer:
		target, ok := tc.FromType(value.Elem())
		if ok {
			return TypePtr{Target: target}, true
		}
	default:
		slog.Debug("type creator: unknown type", slog.Any("type", value))
	}

	return nil, false
}

func (tc *typeCreator) createBasic(value *types.Basic) (TypeBasic, bool) {
	var result TypeBasic

	switch value.Kind() {
	case types.Bool:
		result.ZeroValue = "false"
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		result.ZeroValue = "0"
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		result.ZeroValue = "0"
	case types.Float32, types.Float64:
		result.ZeroValue = "0.0"
	case types.String:
		result.ZeroValue = `""`
	default:
		slog.Debug("type creator: unknown basic type", slog.Any("value", value))

		return result, false
	}

	return result, true
}
