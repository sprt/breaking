package breaking

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"path/filepath"
	"reflect"

	"github.com/sprt/breaking/typecmp"
)

// An Object describes a named language entity such as
// a constant, type, variable, function, or label.
type Object struct {
	types.Object
	fset *token.FileSet
}

// Fpos returns the position of the object within a file.
func (o *Object) Fpos() token.Position {
	return o.fset.Position(o.Pos())
}

// String returns the representation of the object as reported by
// the go/printer package.
func (o *Object) String() string {
	var buf bytes.Buffer
	printer.Fprint(&buf, o.fset, o)
	return buf.String()
}

// An ObjectDiff represents a breaking change in the representation of two objects
// that share the same name across two packages.
type ObjectDiff struct {
	a, b *Object
}

// Name returns the name of the objects.
func (d *ObjectDiff) Name() string {
	return d.a.Name()
}

// Old returns the object before the change.
func (d *ObjectDiff) Old() *Object {
	return d.a
}

// New returns the object after the change, or nil if it was deleted.
func (d *ObjectDiff) New() *Object {
	if d.b.Object == nil {
		return nil
	}
	return d.b
}

func ComparePackages(a, b interface{}) ([]*ObjectDiff, error) {
	afset := token.NewFileSet()
	apkg, err := parseAndCheckPackage(a, afset)
	if err != nil {
		return nil, err
	}
	ascope := apkg.Scope()

	bfset := token.NewFileSet()
	bpkg, err := parseAndCheckPackage(b, bfset)
	if err != nil {
		return nil, err
	}
	bscope := bpkg.Scope()

	var diffs []*ObjectDiff
	for _, name := range ascope.Names() {
		x := ascope.Lookup(name)
		if !x.Exported() {
			continue
		}
		y := bscope.Lookup(name)
		if y != nil && compatible(x, y) {
			continue
		}
		objx := &Object{x, afset}
		objy := &Object{y, bfset}
		diffs = append(diffs, &ObjectDiff{objx, objy})
	}

	return diffs, nil
}

func compatible(a, b types.Object) bool {
	if reflect.TypeOf(a.Type()) != reflect.TypeOf(b.Type()) {
		// Different kinds
		return false
	}

	if typecmp.AssignableTo(a.Type(), b.Type()) {
		return true
	}

	oldStruct, _ := a.Type().Underlying().(*types.Struct)
	newStruct, _ := b.Type().Underlying().(*types.Struct)
	if oldStruct != nil && newStruct != nil {
		if oldStruct.NumFields() == 0 {
			return true
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
			return true
		}

		newExported := []*types.Var{}
		for i := 0; i < newStruct.NumFields(); i++ {
			if newStruct.Field(i).Exported() {
				newExported = append(newExported, newStruct.Field(i))
			} else if oldUnexportedNum == 0 { // oldExportedNum > 0
				// The old struct does not have unexported fields
				// but the new struct does
				return false
			}
		}

		for i, oldf := range oldExported {
			// If the old struct does not have unexported fields,
			// the order of the exported fields must be preserved.
			// In any case, exported fields must not be removed.
			if oldUnexportedNum == 0 {
				if i >= len(newExported) {
					return false
				}
				newf := newExported[i]
				if newf.Name() != oldf.Name() || !typecmp.AssignableTo(oldf.Type(), newf.Type()) {
					return false
				}
			} else {
				for _, f := range newExported {
					if f.Name() == oldf.Name() {
						return typecmp.AssignableTo(oldf.Type(), f.Type())
					}
				}
				// No matching field in newExported
				return false
			}
		}

		return true
	}

	return false
}

func parseAndCheckPackage(f interface{}, fset *token.FileSet) (*types.Package, error) {
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
