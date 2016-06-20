package breaking

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"path/filepath"
	"reflect"

	"github.com/sprt/breaking/typecmp"
)

type Report struct {
	Deleted []types.Object
}

func CompareFiles(a, b interface{}) (*Report, error) {
	oldPkg, err := parseAndCheckPackage(a)
	if err != nil {
		return nil, err
	}

	newPkg, err := parseAndCheckPackage(b)
	if err != nil {
		return nil, err
	}

	analyzer := &analyzer{a: oldPkg.Scope(), b: newPkg.Scope()}
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
	// We don't care about a previously unexported name:
	// both adding or removing an unexported name and
	// adding an exported name are okay.
	if !a.Exported() {
		return false
	}

	// Exported name removed
	if b == nil {
		return true
	}

	// Different kinds
	if reflect.TypeOf(a.Type()) != reflect.TypeOf(b.Type()) {
		return true
	}

	if typecmp.AssignableTo(a.Type(), b.Type()) {
		return false
	}

	oldStruct, _ := a.Type().Underlying().(*types.Struct)
	newStruct, _ := b.Type().Underlying().(*types.Struct)
	if oldStruct != nil && newStruct != nil {
		if oldStruct.NumFields() == 0 {
			return false
		}

		oldExported := []*types.Var{}
		oldUnexportedNum := 0
		for i := 0; i < oldStruct.NumFields(); i++ {
			if oldStruct.Field(i).Exported() {
				oldExported = append(oldExported, oldStruct.Field(i))
			} else {
				oldUnexportedNum++
			}
		}
		if len(oldExported) == 0 {
			return false
		}

		newExported := []*types.Var{}
		for i := 0; i < newStruct.NumFields(); i++ {
			if newStruct.Field(i).Exported() {
				newExported = append(newExported, newStruct.Field(i))
			} else if oldUnexportedNum == 0 { // oldExportedNum > 0
				// The old struct does not have unexported fields
				// but the new struct does
				return true
			}
		}

		for i, oldf := range oldExported {
			// If the old struct does not have unexported fields,
			// the order of the exported fields must be preserved.
			// In any case, exported fields must not be removed.
			if oldUnexportedNum == 0 {
				if i >= len(newExported) {
					return true
				}
				newf := newExported[i]
				if newf.Name() != oldf.Name() || !typecmp.AssignableTo(oldf.Type(), newf.Type()) {
					return true
				}
			} else {
				for _, f := range newExported {
					if f.Name() == oldf.Name() {
						return !typecmp.AssignableTo(oldf.Type(), f.Type())
					}
				}
				return true // no matching field in newExported
			}
		}

		return false
	}

	return true
}

func parseAndCheckPackage(f interface{}) (*types.Package, error) {
	var fset = token.NewFileSet()
	var path string
	var pkg *ast.Package

	switch ff := f.(type) {
	case string:
		path = filepath.Dir(ff)
		pkgs, err := parser.ParseDir(fset, path, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, p := range pkgs {
			pkg = p
			continue
		}
		if pkg == nil {
			return nil, errors.New("no package found")
		}

	case map[string]io.Reader:
		pkg = &ast.Package{Files: make(map[string]*ast.File)}
		for filename, reader := range ff {
			path = filepath.Dir(filename)
			if src, err := parser.ParseFile(fset, filename, reader, 0); err == nil {
				name := src.Name.Name
				pkg.Name = name
				pkg.Files[filename] = src
			} else {
				return nil, err
			}
		}

	default:
		panic(f)
	}

	conf := &types.Config{
		Error: func(err error) {
			fmt.Println("type checker:", err)
		},
		IgnoreFuncBodies: true,
		Importer:         importer.Default(),
	}

	files := make([]*ast.File, 0, len(pkg.Files))
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	return conf.Check(path, fset, files, nil)
}
