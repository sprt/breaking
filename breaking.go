package breaking

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"golang.org/x/tools/go/types/typeutil"
)

type Report struct {
	Deleted []types.Object
}

func CompareFiles(a, b string) (*Report, error) {
	oldPkg, err := parseAndCheckFromFilename(a)
	if err != nil {
		return nil, err
	}

	newPkg, err := parseAndCheckFromFilename(b)
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

	if types.AssignableTo(a.Type(), b.Type()) {
		return false
	}

	oldStruct, _ := a.Type().Underlying().(*types.Struct)
	newStruct, _ := b.Type().Underlying().(*types.Struct)
	if oldStruct != nil && newStruct != nil {
		if oldStruct.NumFields() == 0 {
			return false
		}

		oldExportedNum := 0
		oldUnexportedNum := 0
		for i := 0; i < oldStruct.NumFields(); i++ {
			if oldStruct.Field(i).Exported() {
				oldExportedNum++
			} else {
				oldUnexportedNum++
			}
		}
		if oldExportedNum == 0 {
			return false
		}

		fields := make(map[*fieldKey]int)
		for i := 0; i < oldStruct.NumFields(); i++ {
			f := oldStruct.Field(i)
			if f.Exported() {
				fields[fieldToKey(f)] |= 1
			}
		}
		for i := 0; i < newStruct.NumFields(); i++ {
			f := newStruct.Field(i)
			if f.Exported() {
				fields[fieldToKey(f)] |= 2
			}
		}

		for _, v := range fields {
			if v == 2 && oldUnexportedNum > 0 {
				return false
			}
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

func parseAndCheckFromFilename(path string) (*types.Package, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	fset := token.NewFileSet()
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	return parseAndCheck(fset, dir, map[string]io.Reader{base: fd})
}

func parseAndCheck(fset *token.FileSet, path string, files map[string]io.Reader) (*types.Package, error) {
	pkg, err := parseFiles(fset, files)
	if err != nil {
		return nil, err
	}
	return checkPackage(fset, path, pkg)
}

func parseFiles(fset *token.FileSet, files map[string]io.Reader) (*ast.Package, error) {
	pkg := &ast.Package{Files: make(map[string]*ast.File)}
	for filename, reader := range files {
		if src, err := parser.ParseFile(fset, filename, reader, 0); err == nil {
			name := src.Name.Name
			pkg.Name = name
			pkg.Files[filename] = src
		} else {
			return nil, err
		}
	}
	return pkg, nil
}

func checkPackage(fset *token.FileSet, path string, pkg *ast.Package) (*types.Package, error) {
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
