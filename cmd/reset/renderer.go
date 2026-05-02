package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"go/format"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

const (
	templateMain           = "main.gotmpl"
	templateFieldBasic     = "field_basic.gotmpl"
	templateFieldSlice     = "field_slice.gotmpl"
	templateFieldMap       = "field_map.gotmpl"
	templateFieldPtr       = "field_ptr.gotmpl"
	templateFieldStruct    = "field_struct.gotmpl"
	templateFieldInterface = "field_interface.gotmpl"
)

//go:embed templates/*.gotmpl
var templateFS embed.FS

// Renderer adds methods for structs from [AnalyzedPackage] using a collection of [template.Template].
type Renderer struct {
	stderr  io.Writer
	tmpl    *template.Template
	funcs   template.FuncMap
	bufpool sync.Pool
}

type renderedPackage struct {
	Name    string
	Structs []Struct
}

type renderedField struct {
	RefName          string
	Type             Renderable
	IndirectionLevel int
}

// NewRenderer creates an instance of [Renderer].
func NewRenderer(stderr io.Writer) *Renderer {
	r := &Renderer{
		stderr: stderr,
		bufpool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
	}

	r.funcs = template.FuncMap{
		"join":             strings.Join,
		"add":              func(a, b int) int { return a + b },
		"renderTypeParams": r.renderTypeParams,
		"renderRefName":    r.renderRefName,
		"renderField":      r.renderField,
	}

	return r
}

// Init is called by [pipeline.Sink] stage runner on stage startup.
// For [Renderer], it parses templates embedded into the binary (see [templateFS]).
func (r *Renderer) Init(ctx context.Context) error {
	var err error

	r.tmpl, err = template.New(outputFilename).Funcs(r.funcs).ParseFS(templateFS, "templates/*.gotmpl")
	if err != nil {
		return fmt.Errorf("cannot parse templates: %w", err)
	}

	return nil
}

// Consume is called by workers of [pipeline.Sink] stage, in a loop, **concurrently**.
// It renders loaded templates for given [AnalyzedPackage], formats resulting source code, and
// writes formatted source code into [outputFilename] located inside package's directory.
func (r *Renderer) Consume(ctx context.Context, pkg *AnalyzedPackage) error {
	if pkg == nil || len(pkg.Structs) == 0 { // no structs -> nothing to generate
		return nil
	}

	slog.DebugContext(ctx, "rendering package", slog.String("name", pkg.Name), slog.String("dir", pkg.Dir))

	buf := r.getBuffer()
	defer r.bufpool.Put(buf)

	err := r.tmpl.ExecuteTemplate(buf, templateMain, renderedPackage{
		Name:    pkg.Name,
		Structs: pkg.Structs,
	})
	if err != nil {
		return fmt.Errorf("cannot render template for package %s: %w", pkg.Name, err)
	}

	content, err := format.Source(buf.Bytes())
	if err != nil {
		r.dumpInvalidCodeToStderr(pkg, buf.Bytes())

		return fmt.Errorf("cannot format generated code for package %s: %w", pkg.Name, err)
	}

	outPath := filepath.Join(pkg.Dir, outputFilename)

	err = os.WriteFile(outPath, content, 0644)
	if err != nil {
		return fmt.Errorf("cannot write generated code to %s: %w", outPath, err)
	}

	return nil
}

// Close is called by [pipeline.Sink] stage runner on stage shutdown.
// For [Renderer], it currently does nothing.
func (r *Renderer) Close(ctx context.Context) error {
	return nil
}

func (r *Renderer) getBuffer() *bytes.Buffer {
	buf := r.bufpool.Get().(*bytes.Buffer)

	buf.Reset()

	return buf
}

func (r *Renderer) renderTypeParams(typeParams []string) string {
	if len(typeParams) == 0 {
		return ""
	}

	return "[" + strings.Join(typeParams, ",") + "]"
}

func (r *Renderer) renderRefName(refName string, indirectionLevel int) string {
	if indirectionLevel <= 0 {
		return refName
	}

	return strings.Repeat("*", indirectionLevel) + refName
}

func (r *Renderer) renderField(refName string, ftype Renderable, indirectionLevel int) (string, error) {
	buf := r.getBuffer()
	defer r.bufpool.Put(buf)

	err := r.tmpl.ExecuteTemplate(buf, ftype.TemplateName(), renderedField{
		RefName:          refName,
		Type:             ftype,
		IndirectionLevel: indirectionLevel,
	})
	if err != nil {
		return "", err
	}

	trimmed := strings.TrimSpace(buf.String())

	return trimmed, nil
}

func (r *Renderer) dumpInvalidCodeToStderr(pkg *AnalyzedPackage, code []byte) {
	fmt.Fprintf(r.stderr, "INVALID CODE: pkg=%s (%s)\n", pkg.Name, pkg.Dir)
	fmt.Fprint(r.stderr, "<rendered>\n")

	num := 1
	for line := range bytes.Lines(code) {
		fmt.Fprintf(r.stderr, "%d: %s", num, line)
		num++
	}

	fmt.Fprint(r.stderr, "</rendered>\n")
}
