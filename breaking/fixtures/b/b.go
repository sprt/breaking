package main

// deleted

var VarTypeChanged float64

type InterfaceMethodAdded interface {
	Foo()
	Bar()
}

type InterfaceMethodDeleted interface {
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

func VarToFunc() {}

func FuncParameterAdded(foo int) {}

func FuncParamTypeChanged(foo float64) {}

func FuncRetTypeChanged() float64 {
	return 1
}

// deleted

func FuncParamRenamed(bar int) {}
