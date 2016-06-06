package main

// deleted

func VarToFunc() {}

func FuncParameterAdded(foo int) {}

func FuncParamTypeChanged(foo float64) {}

func FuncResAdded() int { return 1 }

func FuncRetTypeChanged() float64 {
	return 1
}

type InterfaceMethodAdded interface {
	Foo()
	Bar()
}

type InterfaceMethodParameterAdded interface {
	Foo(foo int)
}

type InterfaceMethParamTypeChanged interface {
	Foo(foo float64)
}

type InterfaceMethRetTypeChanged interface {
	Foo() float64
}

type StructExportedAddedExported struct {
	Foo, Bar int
}

type StructExportedAddedUnexported struct {
	Foo, foo int
}

type StructExportedRemoved struct {
}

type StructExportedRepositioned struct {
	Bar string
	Foo int
}

type StructFieldRenamed struct {
	Bar int
}

type StructMixedExportedRemoved struct {
	foo int
}

var VarTypeChanged float64

var TypeStructToVar struct{}

// not deleted

func FuncParamRenamed(bar int) {}

func FuncResRenamed() (bar int) { return }

type InterfaceMethodDeleted interface {
}

type StructEmptyAddedExported struct {
	Foo int
}

type StructEmptyAddedUnexported struct {
	foo int
}

type StructUnexportedAddedExported struct {
	foo, Foo int
}

type StructUnexportedAddedUnexported struct {
	foo, bar int
}

type StructUnexportedRepositioned struct {
	bar string
	foo int
}

type StructMixedAddedExported struct {
	foo, Foo, Bar int
}

type StructMixedRepositionedExported struct {
	foo int
	Bar string
	Foo int
}

type StructMixedRepositionedUnexported struct {
	foo, bar, Foo int
}

type StructUnexportedRemoved struct {
}