package osexit

// This is essentially a copy of `golang.org/x/tools/go/analysis/passes/buildssa` (from v0.36.0)
// that does not rely on `golang.org/x/tools/go/analysis/passes/ctrlflow` to detect "noreturn"
// functions.
// This behavior was introduced somewhere between v0.36.0 and v0.40.0 versions, and, while it makes
// total sense for normal checkers, it totally breaks our `osexit` checker.
// The breakage happens because SSA callgraph from the new (v0.40.0) version does not contain all the
// functions that SSA used to detect in v0.36.0; this is due to pruning of "noreturn" functions
// detected by `ctrlflow` analyzer.
//
// Although in the long term `osexit` checker functionality might need some rethinking,
// we need to restore it to the operational state in the short term.
// As we cannot downgrade `golang.org/x/tools` version due to other dependencies requiring at least
// v0.40.0, the easiest way to fix things is to copy SSA building logic into `osexit` checker.

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"
)

// SSA provides SSA-form intermediate representation for all the
// source functions in the current package.
type SSA struct {
	Pkg      *ssa.Package
	SrcFuncs []*ssa.Function
}

func buildSSA(pass *analysis.Pass) (*SSA, error) {
	mode := ssa.BuilderMode(0)

	prog := ssa.NewProgram(pass.Fset, mode)

	// Create SSA packages for direct imports.
	for _, p := range pass.Pkg.Imports() {
		prog.CreatePackage(p, nil, nil, true)
	}

	// Create and build the primary package.
	ssapkg := prog.CreatePackage(pass.Pkg, pass.Files, pass.TypesInfo, false)
	ssapkg.Build()

	// Compute list of source functions, including literals,
	// in source order.
	var funcs []*ssa.Function
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			if fdecl, ok := decl.(*ast.FuncDecl); ok {
				// (init functions have distinct Func
				// objects named "init" and distinct
				// ssa.Functions named "init#1", ...)

				fn := pass.TypesInfo.Defs[fdecl.Name].(*types.Func)
				if fn == nil {
					panic(fn)
				}

				f := ssapkg.Prog.FuncValue(fn)
				if f == nil {
					panic(fn)
				}

				var addAnons func(f *ssa.Function)
				addAnons = func(f *ssa.Function) {
					funcs = append(funcs, f)
					for _, anon := range f.AnonFuncs {
						addAnons(anon)
					}
				}
				addAnons(f)
			}
		}
	}

	return &SSA{Pkg: ssapkg, SrcFuncs: funcs}, nil
}
