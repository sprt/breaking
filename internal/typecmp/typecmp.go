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

import (
	"go/types"
	"reflect"
)

func Compatible(x, y types.Object) bool {
	if y == nil {
		return false
	}
	if reflect.TypeOf(x.Type()) != reflect.TypeOf(y.Type()) {
		return false
	}
	return ident(x.Type(), y.Type(), nil, nil)
}

// identical reports whether x and y are identical.
func identical(x, y types.Type) bool {
	return ident(x, y, nil, nil)
}

// An ifacePair is a node in a stack of interface type pairs compared for identity.
type ifacePair struct {
	x, y *types.Interface
	prev *ifacePair
}

func (p *ifacePair) identical(q *ifacePair) bool {
	return p.x == q.x && p.y == q.y || p.x == q.y && p.y == q.x
}

type structPair struct {
	x, y *types.Struct
	prev *structPair
}

func (p *structPair) identical(q *structPair) bool {
	return p.x == q.x && p.y == q.y || p.x == q.y && p.y == q.x
}

func ident(x, y types.Type, pi *ifacePair, ps *structPair) bool {
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
			return x.Len() == y.Len() && ident(x.Elem(), y.Elem(), pi, ps)
		}

	case *types.Slice:
		// Two slice types are identical if they have identical element types.
		if y, ok := y.(*types.Slice); ok {
			return ident(x.Elem(), y.Elem(), pi, ps)
		}

	case *types.Struct:
		// Two struct types are identical if they have the same sequence of fields,
		// and if corresponding fields have the same names, and identical types,
		// and identical tags. Two anonymous fields are considered to have the same
		// name. Lower-case field names from different packages are always different.
		if y, ok := y.(*types.Struct); ok {
			if x.NumFields() == y.NumFields() {
				qs := &structPair{x, y, ps}
				for ps != nil {
					if ps.identical(qs) {
						return true // same pair was compared before
					}
					ps = ps.prev
				}
				for i := 0; i < x.NumFields(); i++ {
					f := x.Field(i)
					g := y.Field(i)
					if f.Anonymous() != g.Anonymous() ||
						x.Tag(i) != y.Tag(i) ||
						f.Name() != g.Name() ||
						!ident(f.Type(), g.Type(), pi, qs) {
						goto compat
					}
				}
				return true
			}
		compat:
			return structcompat(x, y)
		}

	case *types.Pointer:
		// Two pointer types are identical if they have identical base types.
		if y, ok := y.(*types.Pointer); ok {
			return ident(x.Elem(), y.Elem(), pi, ps)
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
						if !ident(v.Type(), w.Type(), pi, ps) {
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
				ident(x.Params(), y.Params(), pi, ps) &&
				ident(x.Results(), y.Results(), pi, ps)
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
				qi := &ifacePair{x, y, pi}
				for pi != nil {
					if pi.identical(qi) {
						return true // same pair was compared before
					}
					pi = pi.prev
				}
				for i := 0; i < x.NumMethods(); i++ {
					f := x.Method(i)
					g := y.Method(i)
					if f.Name() != g.Name() || !ident(f.Type(), g.Type(), qi, ps) {
						return false
					}
				}
				return true
			}
		}

	case *types.Map:
		// Two map types are identical if they have identical key and value types.
		if y, ok := y.(*types.Map); ok {
			return ident(x.Key(), y.Key(), pi, ps) && ident(x.Elem(), y.Elem(), pi, ps)
		}

	case *types.Chan:
		// Two channel types are identical if they have identical value types
		// and the same direction.
		if y, ok := y.(*types.Chan); ok {
			return x.Dir() == y.Dir() && ident(x.Elem(), y.Elem(), pi, ps)
		}

	case *types.Named:
		return ident(x.Underlying(), y.Underlying(), pi, ps)

	default:
		panic("unreachable")
	}

	return false
}

func structcompat(x, y *types.Struct) bool {
	if x.NumFields() == 0 {
		return true
	}

	var oldExported []*types.Var
	var oldUnexportedNum int
	for i := 0; i < x.NumFields(); i++ {
		if x.Field(i).Exported() {
			oldExported = append(oldExported, x.Field(i))
		} else {
			oldUnexportedNum++
		}
	}
	if len(oldExported) == 0 {
		return true
	}

	var newExported []*types.Var
	for i := 0; i < y.NumFields(); i++ {
		if y.Field(i).Exported() {
			newExported = append(newExported, y.Field(i))
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
			if newf.Name() != oldf.Name() || !identical(oldf.Type(), newf.Type()) {
				return false
			}
		} else {
			for _, f := range newExported {
				if f.Name() == oldf.Name() {
					return identical(oldf.Type(), f.Type())
				}
			}
			// No matching field in newExported
			return false
		}
	}

	return true
}
