// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package typecmp

import "go/types"

// AssignableTo reports whether a value of type V is assignable to a variable of type T.
func AssignableTo(V, T types.Type) bool {
	if isNil(T) {
		return true // avoid spurious errors
	}

	if Identical(V, T) {
		return true
	}

	Vu := V.Underlying()
	Tu := T.Underlying()

	// T is an interface type and x implements T
	// (Do this check first as it might succeed early.)
	if Ti, ok := Tu.(*types.Interface); ok {
		if types.Implements(V, Ti) {
			return true
		}
	}

	// x's type V and T have identical underlying types
	// and at least one of V or T is not a named type
	if Identical(Vu, Tu) && (!isNamed(V) || !isNamed(T)) {
		return true
	}

	// x is a bidirectional channel value, T is a channel
	// type, x's type V and T have identical element types,
	// and at least one of V or T is not a named type
	if Vc, ok := Vu.(*types.Chan); ok && Vc.Dir() == types.SendRecv {
		if Tc, ok := Tu.(*types.Chan); ok && Identical(Vc.Elem(), Tc.Elem()) {
			return !isNamed(V) || !isNamed(T)
		}
	}

	// x is the predeclared identifier nil and T is a pointer,
	// function, slice, map, channel, or interface type
	if isNil(V) {
		switch t := Tu.(type) {
		case *types.Basic:
			if t.Kind() == types.UnsafePointer {
				return true
			}
		case *types.Pointer, *types.Signature, *types.Slice, *types.Map, *types.Chan, *types.Interface:
			return true
		}
		return false
	}

	// x is an untyped constant representable by a value of type T
	// TODO(gri) This is borrowing from checker.convertUntyped and
	//           checker.representable. Need to clean up.
	if isUntyped(Vu) {
		switch t := Tu.(type) {
		case *types.Basic:
			// The result of a comparison is an untyped boolean,
			// but may not be a constant.
			if Vb, _ := Vu.(*types.Basic); Vb != nil {
				return Vb.Kind() == types.UntypedBool && isBoolean(Tu)
			}
		case *types.Interface:
			return isNil(V) || t.Empty()
		case *types.Pointer, *types.Signature, *types.Slice, *types.Map, *types.Chan:
			return isNil(V)
		}
	}

	return false
}

// Identical reports whether x and y are identical.
func Identical(x, y types.Type) bool {
	return identical(x, y, nil)
}

// An ifacePair is a node in a stack of interface type pairs compared for identity.
type ifacePair struct {
	x, y *types.Interface
	prev *ifacePair
}

func (p *ifacePair) identical(q *ifacePair) bool {
	return p.x == q.x && p.y == q.y || p.x == q.y && p.y == q.x
}

func identical(x, y types.Type, p *ifacePair) bool {
	if x == y {
		return true
	}

	switch x := x.(type) {
	case *types.Basic:
		// Basic types are singletons except for the rune and byte
		// aliases, thus we cannot solely rely on the x == y check
		// above.
		if y, ok := y.(*types.Basic); ok {
			return x.Kind() == y.Kind()
		}

	case *types.Array:
		// Two array types are identical if they have identical element types
		// and the same array length.
		if y, ok := y.(*types.Array); ok {
			return x.Len() == y.Len() && identical(x.Elem(), y.Elem(), p)
		}

	case *types.Slice:
		// Two slice types are identical if they have identical element types.
		if y, ok := y.(*types.Slice); ok {
			return identical(x.Elem(), y.Elem(), p)
		}

	case *types.Struct:
		// Two struct types are identical if they have the same sequence of fields,
		// and if corresponding fields have the same names, and identical types,
		// and identical tags. Two anonymous fields are considered to have the same
		// name. Lower-case field names from different packages are always different.
		if y, ok := y.(*types.Struct); ok {
			if x.NumFields() == y.NumFields() {
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					g := y.Field(i)
					if f.Anonymous() != g.Anonymous() ||
						x.Tag(i) != y.Tag(i) ||
						f.Name() != g.Name() ||
						!identical(f.Type(), g.Type(), p) {
						return false
					}
				}
				return true
			}
		}

	case *types.Pointer:
		// Two pointer types are identical if they have identical base types.
		if y, ok := y.(*types.Pointer); ok {
			return identical(x.Elem(), y.Elem(), p)
		}

	case *types.Tuple:
		// Two tuples types are identical if they have the same number of elements
		// and corresponding elements have identical types.
		if y, ok := y.(*types.Tuple); ok {
			if x.Len() == y.Len() {
				if x != nil {
					for i := 0; i < x.Len(); i++ {
						v := x.At(i)
						w := y.At(i)
						if !identical(v.Type(), w.Type(), p) {
							return false
						}
					}
				}
				return true
			}
		}

	case *types.Signature:
		// Two function types are identical if they have the same number of parameters
		// and result values, corresponding parameter and result types are identical,
		// and either both functions are variadic or neither is. Parameter and result
		// names are not required to match.
		if y, ok := y.(*types.Signature); ok {
			return x.Variadic() == y.Variadic() &&
				identical(x.Params(), y.Params(), p) &&
				identical(x.Results(), y.Results(), p)
		}

	case *types.Interface:
		// Two interface types are identical if they have the same set of methods with
		// the same names and identical function types. Lower-case method names from
		// different packages are always different. The order of the methods is irrelevant.
		if y, ok := y.(*types.Interface); ok {
			if x.NumMethods() == y.NumMethods() {
				// Interface types are the only types where cycles can occur
				// that are not "terminated" via named types; and such cycles
				// can only be created via method parameter types that are
				// anonymous interfaces (directly or indirectly) embedding
				// the current interface. Example:
				//
				//    type T interface {
				//        m() interface{T}
				//    }
				//
				// If two such (differently named) interfaces are compared,
				// endless recursion occurs if the cycle is not detected.
				//
				// If x and y were compared before, they must be equal
				// (if they were not, the recursion would have stopped);
				// search the ifacePair stack for the same pair.
				//
				// This is a quadratic algorithm, but in practice these stacks
				// are extremely short (bounded by the nesting depth of interface
				// type declarations that recur via parameter types, an extremely
				// rare occurrence). An alternative implementation might use a
				// "visited" map, but that is probably less efficient overall.
				q := &ifacePair{x, y, p}
				for p != nil {
					if p.identical(q) {
						return true // same pair was compared before
					}
					p = p.prev
				}
				for i := 0; i < x.NumMethods(); i++ {
					f := x.Method(i)
					g := y.Method(i)
					if f.Name() != g.Name() || !identical(f.Type(), g.Type(), q) {
						return false
					}
				}
				return true
			}
		}

	case *types.Map:
		// Two map types are identical if they have identical key and value types.
		if y, ok := y.(*types.Map); ok {
			return identical(x.Key(), y.Key(), p) && identical(x.Elem(), y.Elem(), p)
		}

	case *types.Chan:
		// Two channel types are identical if they have identical value types
		// and the same direction.
		if y, ok := y.(*types.Chan); ok {
			return x.Dir() == y.Dir() && identical(x.Elem(), y.Elem(), p)
		}

	case *types.Named:
		return identical(x.Underlying(), y.Underlying(), p)

	default:
		panic("unreachable")
	}

	return false
}

func isBoolean(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsBoolean != 0
}

func isNamed(typ types.Type) bool {
	if _, ok := typ.(*types.Basic); ok {
		return ok
	}
	_, ok := typ.(*types.Named)
	return ok
}

func isNil(typ types.Type) bool {
	return typ == types.Typ[types.UntypedNil]
}

func isUntyped(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsUntyped != 0
}
