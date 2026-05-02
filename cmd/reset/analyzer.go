package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"
	"strings"
	"unicode"
)

// AnalyzedPackage is a result of analysis of [LoadedPackage]; it contains a list of structs
// that need generation of `Reset()` method.
type AnalyzedPackage struct {
	Name    string
	Dir     string
	Structs []Struct
}

// Analyzer performs analysis of AST and types of a [LoadedPackage] in order to extract
// structs that have magic comment (see [magicComment] constant).
type Analyzer struct{}

type packageAnalyzer struct {
	syntax []*ast.File
	types  *types.Info
	result []Struct
}

// NewAnalyzer creates an instance of [Analyzer].
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Init is called by [pipeline.Step] stage runner on stage startup.
// For [Analyzer], it does nothing.
func (a *Analyzer) Init(_ context.Context) error {
	return nil
}

// Process is called by workers of [pipeline.Step] stage, in a loop, **concurrently**.
// It is responsible for analyzing incoming packages to extract structs that have [magicComment].
func (a *Analyzer) Process(ctx context.Context, pkg *LoadedPackage) (*AnalyzedPackage, error) {
	if pkg == nil {
		return nil, nil
	}

	pkgAnalyzer := &packageAnalyzer{
		syntax: pkg.Syntax,
		types:  pkg.TypesInfo,
		result: make([]Struct, 0),
	}

	err := pkgAnalyzer.Analyze(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot analyze package %s: %w", pkg.Name, err)
	}

	result := &AnalyzedPackage{
		Name:    pkg.Name,
		Dir:     pkg.Dir,
		Structs: pkgAnalyzer.result,
	}

	return result, nil
}

// Close is called by [pipeline.Step] stage runner on stage shutdown.
// For [Analyzer], it does nothing.
func (a *Analyzer) Close(_ context.Context) error {
	return nil
}

// Analyze traverses AST trees for a single package and extracts all structs with [magicComment].
func (pa *packageAnalyzer) Analyze(ctx context.Context) error {
	for _, file := range pa.syntax {
		err := pa.analyzeFile(ctx, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pa *packageAnalyzer) analyzeFile(ctx context.Context, file *ast.File) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("analyzer aborted: %w", ctx.Err())
	default:
	}

	for _, decl := range file.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			pa.analyzeGenDecl(v)
		}
	}

	return nil
}

func (pa *packageAnalyzer) analyzeGenDecl(decl *ast.GenDecl) {
	if decl.Tok != token.TYPE {
		return
	}

	for _, spec := range decl.Specs {
		ts := spec.(*ast.TypeSpec) // safe because we ensure `decl.Tok == token.TYPE` above.

		pa.analyzeTypeSpec(ts, decl.Doc)
	}
}

func (pa *packageAnalyzer) analyzeTypeSpec(spec *ast.TypeSpec, declDoc *ast.CommentGroup) {
	stype, ok := spec.Type.(*ast.StructType)
	if !ok {
		return
	}

	// When a single type is declared (e.g. `type Something struct`), the comment above it
	// belongs to [ast.GenDecl], not to [ast.TypeSpec].
	// However, if types are declared in a group (e.g. `type ( multiple declarations )`), a comment above
	// a single type inside the group belongs to [ast.TypeSpec], but if the type does not have a
	// comment above it, then the comment from the group ([ast.GenDecl]) should be applicable.
	if !containsMagicComment(spec.Doc) && !containsMagicComment(declDoc) {
		return
	}

	sobj := Struct{
		Receiver:   extractStructReceiver(spec.Name),
		Name:       spec.Name.Name,
		TypeParams: extractSpecTypeParams(spec),
		Fields:     make([]Field, 0),
	}

	pa.populateStructFields(&sobj, stype)

	if len(sobj.Fields) > 0 { // makes no sense to "reset" empty structs
		pa.result = append(pa.result, sobj)
	}
}

func (pa *packageAnalyzer) populateStructFields(sobj *Struct, astStruct *ast.StructType) {
	stype, ok := pa.types.TypeOf(astStruct).(*types.Struct)
	if !ok {
		slog.Debug("package analyzer: cannot find type for AST struct", slog.Any("struct", astStruct))

		return
	}

	tc := &typeCreator{
		gotypes: pa.types,
	}

	for field := range stype.Fields() {
		ftype, ok := tc.FromType(field.Type())
		if !ok {
			continue
		}

		sobj.Fields = append(sobj.Fields, Field{Name: field.Name(), Type: ftype})
	}
}

func containsMagicComment(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}

	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimSpace(text)

		if text == magicComment {
			return true
		}
	}

	return false
}

func extractSpecTypeParams(spec *ast.TypeSpec) []string {
	if spec.TypeParams == nil {
		return nil
	}

	extracted := make([]string, 0)

	for _, field := range spec.TypeParams.List {
		for _, name := range field.Names {
			extracted = append(extracted, name.Name)
		}
	}

	return extracted
}

func extractStructReceiver(ident *ast.Ident) string {
	if ident == nil {
		return ""
	}

	name := []rune(ident.Name)
	if len(name) == 0 { // unlikely, but better to prevent index error below
		return ""
	}

	rcv := []rune{unicode.ToLower(name[0])}

	for _, ch := range name[1:] {
		if unicode.IsUpper(ch) {
			rcv = append(rcv, unicode.ToLower(ch))
		}
	}

	reserved := map[string]bool{ // only shorter than 4 symbols
		"for": true,
		"go":  true,
		"if":  true,
		"map": true,
		"var": true,
	}

	if len(rcv) > 3 || reserved[string(rcv)] {
		rcv = append(rcv, '_') // ensure we have no collisions with Go reserved keywords
	}

	return string(rcv)
}
