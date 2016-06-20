package main

// deleted

func FuncParameterAdded() {}

func FuncParamTypeChanged(foo int) {}

func FuncResAdded() {}

func FuncRetTypeChanged() int {
	return 1
}

type InterfaceMethodAdded interface {
	Foo()
}

type InterfaceMethodParameterAdded interface {
	Foo()
}

type InterfaceMethParamTypeChanged interface {
	Foo(foo int)
}

type InterfaceMethRetTypeChanged interface {
	Foo() int
}

type StructExportedAddedUnexported struct {
	Foo int
}

type StructExportedPrependedExported struct {
	Foo int
}

type StructExportedRepositioned struct {
	Foo int
	Bar string
}

type StructExportedRemoved struct {
	Foo int
}

type StructExportedTypeChanged struct {
	Foo int
}

type StructFieldRenamed struct {
	Foo int
}

type StructMixedExportedRemoved struct {
	Foo, foo int
}

type TypeStructToVar struct{}

var VarDeleted int
var VarToFunc int
var VarTypeChanged int

// not deleted

func FuncParamRenamed(foo int) {}

func FuncResRenamed() (foo int) { return }

type InterfaceMethodDeleted interface {
	Foo()
}

type NamedType int

type StructExportedAppendedExported struct {
	Foo int
}

type StructEmptyAddedExported struct {
}

type StructEmptyAddedUnexported struct {
}

type StructMixedAddedExported struct {
	foo, Foo int
}

type StructMixedRepositionedExported struct {
	foo int
	Foo int
	Bar string
}

type StructMixedRepositionedUnexported struct {
	foo, Foo, bar int
}

type StructMixedExportedNamedType struct {
	foo int
	Foo NamedType
}

type namedType int
type StructNamedType struct {
	Foo namedType
}

type StructUnexportedAddedExported struct {
	foo int
}

type StructUnexportedAddedUnexported struct {
	foo int
}

type StructUnexportedRemoved struct {
	foo int
}

type StructUnexportedRepositioned struct {
	foo int
	bar string
}
