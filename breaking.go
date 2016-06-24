// Package breaking reports breaking changes across two versions of a package.
//
// This is the exhaustive list of changes that are considered breaking:
//  - Removing an exposed name (constant, type, variable, function).
//  - Adding or removing a method in an interface.
//  - Adding or removing a parameter in a function or interface.
//  - Changing the type of a parameter or result in a function or interface.
//  - Adding or removing a result in a function or interface.
//  - Changing the type of an exported struct field.
//  - Removing an exported field from a struct.
//  - Adding an unexported field to a struct containing only exported fields.
//  - Adding an exported field before the last field of a struct
//    containing only exported fields.
//  - Repositioning a field in a struct containing only exported fields.
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

	"github.com/sprt/breaking/typecmp"
)

// An Object describes a named language entity such as
// a constant, type, variable, function.
type Object struct {
	obj  types.Object
	fset *token.FileSet
}

// Name returns the name of the object.
func (o *Object) Name() string {
	return o.obj.Name()
}

// Fpos returns the position of the object within a file.
func (o *Object) Fpos() token.Position {
	return o.fset.Position(o.obj.Pos())
}

// An ObjectDiff represents a breaking change in the representation of two objects
// that share the same name across two packages.
type ObjectDiff struct {
	a, b *Object
}

// Name returns the name of the objects.
func (d *ObjectDiff) Name() string {
	return d.a.obj.Name()
}

// Old returns the object before the change.
func (d *ObjectDiff) Old() *Object {
	return d.a
}

// New returns the object after the change, or nil if it was deleted.
func (d *ObjectDiff) New() *Object {
	if d.b.obj == nil {
		return nil
	}
	return d.b
}

// ComparePackages returns the breaking changes introduced by package b
// relative to package a.
//
// A package can be passed as either a string or a map of string -> io.Reader.
// If a string, it is the path to the package.
// If a map, it maps filenames to source code.
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
		if typecmp.Compatible(x, y) {
			continue
		}
		objx := &Object{x, afset}
		objy := &Object{y, bfset}
		diffs = append(diffs, &ObjectDiff{objx, objy})
	}

	return diffs, nil
}


	}


func parseAndCheckPackage(f interface{}, fset *token.FileSet) (*types.Package, error) {
	var path string
	var pkg *ast.Package

	switch ff := f.(type) {
	case string:
		path = ff
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
