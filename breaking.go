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

	"github.com/sprt/breaking/typecmp"
)

// An Object describes a named language entity such as
// a constant, type, variable, function.
type Object struct {
	obj  types.Object
	fset *token.FileSet
	decl ast.Node
}

// Name returns the name of the object.
func (o *Object) Name() string {
	return o.obj.Name()
}

// Fpos returns the position of the object within a file.
func (o *Object) Fpos() token.Position {
	return o.fset.Position(o.obj.Pos())
}

// String returns the object as it appears in code.
func (o *Object) String() string {
	decl := o.decl
	if funcDecl, ok := decl.(*ast.FuncDecl); ok {
		// Strip the body (forward declaration)
		decl = &ast.FuncDecl{
			Doc:  funcDecl.Doc,
			Recv: funcDecl.Recv,
			Name: funcDecl.Name,
			Type: funcDecl.Type,
			Body: nil,
		}
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, o.fset, decl)
	return buf.String()
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
	pkga, err := parseAndCheckPackage(a)
	if err != nil {
		return nil, err
	}

	pkgb, err := parseAndCheckPackage(b)
	if err != nil {
		return nil, err
	}

	var diffs []*ObjectDiff
	for _, name := range pkga.scope.Names() {
		x := pkga.scope.Lookup(name)
		if !x.Exported() {
			continue
		}
		y := pkgb.scope.Lookup(name)
		if typecmp.Compatible(x, y) {
			continue
		}
		objx := &Object{x, pkga.fset, pkga.decls[name]}
		objy := &Object{y, pkgb.fset, pkgb.decls[name]}
		diffs = append(diffs, &ObjectDiff{objx, objy})
	}

	return diffs, nil
}

type pkg struct {
	decls map[string]ast.Node
	fset  *token.FileSet
	scope *types.Scope
}

func parseAndCheckPackage(f interface{}) (*pkg, error) {
	pkg := &pkg{
		fset:  token.NewFileSet(),
		decls: make(map[string]ast.Node),
	}

	var path string
	var parsed *ast.Package

	switch ff := f.(type) {
	case string:
		path = ff
		pkgs, err := parser.ParseDir(pkg.fset, path, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, p := range pkgs {
			parsed = p
			continue
		}
		if parsed == nil {
			return nil, errors.New("no package found")
		}

	case map[string]io.Reader:
		parsed = &ast.Package{Files: make(map[string]*ast.File)}
		for filename, reader := range ff {
			path = filepath.Dir(filename)
			if src, err := parser.ParseFile(pkg.fset, filename, reader, 0); err == nil {
				name := src.Name.Name
				parsed.Name = name
				parsed.Files[filename] = src
			} else {
				return nil, err
			}
		}

	default:
		panic(f)
	}

	for _, f := range parsed.Files {
		for name, obj := range f.Scope.Objects {
			pkg.decls[name] = obj.Decl.(ast.Node)
		}
	}

	conf := &types.Config{
		Error: func(err error) {
			fmt.Println("type checker:", err)
		},
		IgnoreFuncBodies: true,
		Importer:         importer.Default(),
	}

	files := make([]*ast.File, 0, len(parsed.Files))
	for _, f := range parsed.Files {
		files = append(files, f)
	}

	checked, err := conf.Check(path, pkg.fset, files, nil)
	if err != nil {
		return nil, err
	}

	pkg.scope = checked.Scope()
	return pkg, nil
}
