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

func FuncParameterAdded() {}
