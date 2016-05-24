package main

var VarDeleted int
var VarToFunc int
var VarTypeChanged int

type InterfaceMethodAdded interface {
	Foo()
}

type InterfaceMethodDeleted interface {
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

func FuncParameterAdded() {}

func FuncParamTypeChanged(foo int) {}

func FuncRetTypeChanged() int {
	return 1
}
