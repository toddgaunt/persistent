package lists

import "fmt"

type List[T any] struct {
	count int
	first T
	rest  *List[T]
}

// New creates a new persistent list constructed using vals with
// the first element of vals being the head of the list, and the last
// element of vals being the end of the list. As an example, New(1,
// 2, 3, 4) results in a (1, 2, 3, 4) and not (4, 3, 2, 1) as what
// would be constructed if done manually using Cons for each value.
func New[T any](vals ...T) List[T] {
	var l List[T]

	for i := len(vals) - 1; i >= 0; i-- {
		l = l.Conj(vals[i])
	}

	return l
}

// Len returns the number of items in the list. Note that the name Len was
// chosen rather than Len(), as it is called in Clojure, to fit into the Go
// ecosystem better.
func (l List[T]) Len() int {
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
//     With no elements: ()
//     With one element: (1)
//     With more than one element: (1 2 3)
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

// Equal compares two lists to see if the contain the same elements, analogous
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
