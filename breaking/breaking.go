package breaking

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
)

var typeConfig = &types.Config{
	Error: func(err error) {
		fmt.Println(err)
	},
	Importer: importer.Default(),
}

var typeInfo = &types.Info{
	Types: make(map[ast.Expr]types.TypeAndValue),
	Defs:  make(map[*ast.Ident]types.Object),
	Uses:  make(map[*ast.Ident]types.Object),
}

type Report struct {
	Deleted []types.Object
}

func CompareFiles(a, b string) (*Report, error) {
	scopea, err := parseFile(a)
	if err != nil {
		return nil, err
	}

	scopeb, err := parseFile(b)
	if err != nil {
		return nil, err
	}

	analyzer := &analyzer{a: scopea, b: scopeb}

	report := &Report{}
	report.Deleted = analyzer.deleted()

	return report, nil
}

type analyzer struct {
	a, b *types.Scope
}

func (anal *analyzer) deleted() []types.Object {
	deleted := []types.Object{}

	for _, name := range anal.a.Names() {
		obja := anal.a.Lookup(name)
		if !obja.Exported() {
			continue
		}

		objb := anal.b.Lookup(name)
		if objb == nil {
			deleted = append(deleted, obja)
		}
	}

	return deleted
}

func parseFile(filename string) (*types.Scope, error) {
	fset := token.NewFileSet()
	dirname := filepath.Dir(filename)

	pkgs, err := parser.ParseDir(fset, dirname, nil, 0)
	if err != nil {
		return nil, err
	}

	for _, p := range pkgs {
		files := []*ast.File{}
		for _, f := range p.Files {
			files = append(files, f)
		}

		pkg, err := typeConfig.Check(filename, fset, files, typeInfo)
		if err != nil {
			return nil, err
		}

		return pkg.Scope(), nil
	}

	// TODO: do something here
	return nil, nil
}
