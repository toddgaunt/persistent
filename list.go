// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package persistent

import "fmt"

type List[T any] struct {
	count int
	first T
	rest  *List[T]
}

// NewList creates a new persistent list constructed using vals with
// the first element of vals being the head of the list, and the last
// element of vals being the end of the list. As an example, NewList(1,
// 2, 3, 4) results in a (1, 2, 3, 4) and not (4, 3, 2, 1) as what
// would be constructed if done manually using Cons for each value.
func NewList[T any](vals ...T) List[T] {
	var l List[T]

	for i := len(vals) - 1; i >= 0; i-- {
		l = l.Cons(vals[i])
	}

	return l
}

// Empty returns true if the list is empty, false otherwise
func (l List[T]) Empty() bool {
	return l.count == 0
}

// Count returns the number of items in the list.
func (l List[T]) Count() int {
	return l.count
}

// First returns the value contained within the head of the list.
func (l List[T]) First() T {
	return l.first
}

// Rest returns a list of items containing all but the first item of the
// original list.
func (l List[T]) Rest() List[T] {
	return *l.rest
}

// Cons returns a new list where val is the new head, and the original list is
// the rest.
func (l List[T]) Cons(val T) List[T] {
	return List[T]{
		count: l.count + 1,
		first: val,
		rest:  &l,
	}
}

// Conj adds a value to the beginning of the list. Functionally the same as Cons,
// but used to satisfy the Collection interface.
func (l List[T]) Conj(val T) List[T] {
	return l.Cons(val)
}

// String returns a representation of a list similar to standard Go types:
//     With no elements: ()
//     With one element: (1)
//     With more than one element: (1, 2, 3)
func (l List[T]) String() string {
	s := "("
	s += fmt.Sprintf("%v", l.first)
	for walk := l.rest; walk != nil; walk = walk.rest {
		s += fmt.Sprintf(", %v", walk.first)
	}
	s += ")"

	return s
}
