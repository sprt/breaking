package breaking

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/types/typeutil"
)

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
		a := anal.a.Lookup(name)
		b := anal.b.Lookup(name)
		if isDeleted(a, b) {
			deleted = append(deleted, a)
		}
	}
	return deleted
}

func isDeleted(a, b types.Object) bool {
	if !a.Exported() {
		return false
	}

	if b == nil {
		return true
	}

	if types.AssignableTo(a.Type(), b.Type()) {
		return false
	}

	structa, _ := a.Type().Underlying().(*types.Struct)
	structb, _ := b.Type().Underlying().(*types.Struct)
	if structa != nil && structb != nil {
		if structa.NumFields() == 0 {
			return false
		}

		oldExportedNum := 0
		oldUnexportedNum := 0
		for i := 0; i < structa.NumFields(); i++ {
			if structa.Field(i).Exported() {
				oldExportedNum++
			} else {
				oldUnexportedNum++
			}
		}

		if oldExportedNum == 0 {
			return false
		}

		fields := make(map[*fieldKey]int)
		for i := 0; i < structa.NumFields(); i++ {
			f := structa.Field(i)
			if f.Exported() {
				k := fieldToKey(f)
				fields[k] |= 1
			}
		}
		for i := 0; i < structb.NumFields(); i++ {
			f := structb.Field(i)
			if f.Exported() {
				k := fieldToKey(f)
				fields[k] |= 2
			}
		}

		addedExported := false
		for _, v := range fields {
			if v == 2 {
				addedExported = true
			}
		}

		if oldUnexportedNum > 0 && addedExported {
			return false
		}

		return true
	}

	return true
}

type fieldKey struct {
	name string
	typ  uint32
}

func fieldToKey(f *types.Var) *fieldKey {
	h := typeutil.MakeHasher()
	return &fieldKey{name: f.Name(), typ: h.Hash(f.Type())}
}

func parseFile(filename string) (*types.Scope, error) {
	config := &types.Config{
		Error: func(err error) {
			fmt.Println(err)
		},
		Importer: importer.Default(),
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

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

		pkg, err := config.Check(filename, fset, files, info)
		if err != nil {
			return nil, err
		}

		return pkg.Scope(), nil
	}

	// TODO: do something here
	return nil, nil
}
