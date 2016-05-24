package main

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

func VarToFunc() {}

func FuncParameterAdded(foo int) {}
