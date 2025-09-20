// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package list provides a persistent List datastructure similar to the one
// found in the Clojure programming language. The actual implementation is
// a very simple linked list, with an API designed carefully to integrate
// well with other persistent data structures.
package list

import (
	"fmt"
)

// List is a persistent data structure that can be treated as a value
// (similarly to an int or another built-in Go type). This means even when
// modifying a List provided by this package, the previous version
// of that List can used without any any of the modifications apparent.
type List[T any] struct {
	count int
	first T
	rest  *List[T]
}

// New creates a new persistent list constructed using vals with
// the first of vals being the head of the list, and the last
// of vals being the end of the list. As an example, New(1,
// 2, 3, 4) results in (1, 2, 3, 4). Note that this is the reverse of
// (4, 3, 2, 1) which is what would be constructed with Conj.
func New[T any](vals ...T) List[T] {
	var l List[T]

	for i := len(vals) - 1; i >= 0; i-- {
		l = l.Conj(vals[i])
	}

	return l
}

// Len returns the number of values in the list.
func (l List[T]) Len() int {
	return l.count
}

// First returns the value contained within the head of the list.
func (l List[T]) First() T {
	return l.first
}

// Rest returns a list of items containing all but the first item.
func (l List[T]) Rest() List[T] {
	return *l.rest
}

// Conj returns a new list where val is the new head, and the original list is
// the rest.
func (l List[T]) Conj(val T) List[T] {
	return List[T]{
		count: l.count + 1,
		first: val,
		rest:  &l,
	}
}

// String returns a representation of a list similar to standard Go types
// when using the "%v" formatting verb as in the standard fmt package:
//     With no items: ()
//     With one item: (1)
//     With more than one item: (1 2 3)
func (l List[T]) String() string {
	if l.count == 0 {
		return "()"
	}

	s := "("
	s += fmt.Sprintf("%v", l.first)
	for walk := l.rest; walk.count > 0; walk = walk.rest {
		s += fmt.Sprintf(" %v", walk.first)
	}
	s += ")"

	return s
}

// IsEmpty returns true if the list is empty, false otherwise
func IsEmpty[T any](l List[T]) bool {
	return l.count == 0
}

// Equal compares two lists to see if the contain the same items, analogous
// to bytes.Equal from the standard Go bytes package.
func Equal[T comparable](a, b List[T]) bool {
	if a.Len() != b.Len() {
		return false
	}

	for aw, bw := &a, &b; aw.count > 0 && bw.count > 0; aw, bw = aw.rest, bw.rest {
		if aw.first != bw.first {
			return false
		}
	}
	return true
}
