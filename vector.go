// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package persistent

import "fmt"

// These constants determine the maximum width of vector nodes
const vecNodeBits = 2
const vecNodeWidth = 1 << vecNodeBits
const vecNodeMask = vecNodeWidth - 1

type TVec[T any] struct {
	invalid bool
	count int         // Number of elements in this vector
	depth int         // Depth of the tree under root
	root  *vecNode[T] // Root of the tree; Contains either child nodes or elements
	tail  []T         // Quickly access elements at the end of the vector
}

type Vec[T any] struct {
	count int         // Number of elements in this vector
	depth int         // Depth of the tree under root
	root  *vecNode[T] // Root of the tree; Contains either child nodes or elements
	tail  []T         // Quickly access elements at the end of the vector
}

type vecNode[T any] struct {
	children []*vecNode[T]
	values   []T
}

// indexAt extracts the bits from i that are needed to index a node at a given
// level in the tree.
func indexAt(level, i int) int {
	return (i >> (level * vecNodeBits)) & vecNodeMask
}

// tailOffset returns the total number of elements within a Vec minus the tail.
func (v *Vec[T]) tailOffset() int {
	return v.count - len(v.tail)
}

// isDeepEnoughToAppend evaluates a depth for a Vec to be deep enough for a
// given count. Returns true if a Vec of depth can be appended to without
// creating a new root, otherwise returns false.
func isDeepEnoughToAppend(depth, count int) bool {
	return (count >> vecNodeBits) <= (1 << depth)
}

// findValues returns the slice of values within the vector which contains the
// value i is associated with.
func (v Vec[T]) findValues(i int) []T {
	if i < 0 || i >= v.count {
		panic("index out of bounds")
	}

	if i >= v.tailOffset() {
		return v.tail
	}

	// The index is not associated with the tail, so do a slow lookup for the
	// node it is associated with.
	var walk = v.root
	for level := v.depth; level > 0; level -= 1 {
		walk = walk.children[indexAt(level, i)]
	}

	return walk.values
}

// NewVec creates a new persistent vector constructed using vals.
func NewVec[T any](vals ...T) Vec[T] {
	var v TVec[T]

	for i := 0; i < len(vals); i++ {
		v = v.Conj(vals[i])
	}

	return v.Persistent()
}

func (v Vec[T]) Count() int {
	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v Vec[T]) Nth(i int) T {
	return v.findValues(i)[indexAt(0, i)]
}

// Peek returns the last value from a vector.
func (v Vec[T]) Peek() T {
	return v.Nth(v.count - 1)
}

// Assoc creates a new vector that contains val at the location indexed by key.
// The key must be greater than zero and less than v.Count().
func (v Vec[T]) Assoc(key int, val T) Vec[T] {
	if key < 0 || key >= v.count {
		panic("index out of bounds")
	}

	var newRoot = v.root
	var newTail = v.tail

	// Either the tail is being updated, or a node in the tree is.
	var leaf []T
	if key >= v.tailOffset() {
		// The value to update is in the tail, so make a copy of the tail
		newTail = make([]T, len(v.tail))
		copy(newTail, v.tail)
		leaf = newTail
	} else {
		// The value to update is in the tree, so create a new path of nodes

		// Clone the root node first so the changes to the path don't effect
		// the old vector
		newRoot = &vecNode[T]{}
		newRoot.children = append([]*vecNode[T]{}, v.root.children...)
		newRoot.values = append([]T{}, v.root.values...)

		var walk = newRoot
		for level := v.depth; level > 0; level -= 1 {
			var oldNode = walk.children[indexAt(level, key)]

			walk.children[indexAt(level, key)] = &vecNode[T]{}
			walk.children = append([]*vecNode[T]{}, oldNode.children...)
			walk.values = append([]T{}, oldNode.values...)

			walk = walk.children[indexAt(level, key)]
		}
		leaf = walk.values
	}

	// Update the value
	leaf[indexAt(0, key)] = val

	return Vec[T]{
		depth: v.depth,
		count: v.count,
		root:  newRoot,
		tail:  newTail,
	}
}

// Conj creates a new vector with a value appended to the end.
func (v Vec[T]) Conj(val T) Vec[T] {
	var newDepth = v.depth
	var newRoot = v.root
	var newTail []T

	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < vecNodeWidth {
		// The tail can still be grown, so make a copy to add the new value to.
		newTail = make([]T, len(v.tail)+1)
		copy(newTail, v.tail)
	} else {
		// There is no room in the tail, so move the tail into the tree.
		if !isDeepEnoughToAppend(v.depth, v.count + 1) {
			// No space left in the current tree, so deepen the tree one level
			// with a new node containing the old root.
			newDepth = v.depth + 1
			newRoot = &vecNode[T]{}
			// TODO(todd): Make this more elegant. Essentially the problem is
			// that go arrays and slices need to be initialized here for child
			// nodes to insert.
			newRoot.children = (&[vecNodeWidth]*vecNode[T]{v.root})[:]
		}

		// Walk through the tree with an indirect pointer to find location the
		// tail will end up being moved to, making copies of nodes along the path.
		var indirect = &newRoot
		for level := newDepth; level > 0; level -= 1 {
			if *indirect == nil {
				*indirect = &vecNode[T]{children: make([]*vecNode[T], vecNodeWidth)}
			} else {
				var newNode = &vecNode[T]{
					children: make([]*vecNode[T], 0, vecNodeWidth),
					values: make([]T, 0, vecNodeWidth),
				}
				newNode.children = append([]*vecNode[T]{}, (*indirect).children...)
				newNode.values = append([]T{}, (*indirect).values...)
				newNode.children = newNode.children[:cap(newNode.children)]
				newNode.values = newNode.values[:cap(newNode.values)]
				*indirect = newNode
			}
			indirect = &(*indirect).children[indexAt(level, v.count-1)]
		}
		// Move the old tail into the trie
		*indirect = &vecNode[T]{values: v.tail}

		// Create a new tail for conjugating the new value to.
		newTail = make([]T, 1)
	}
	newTail[indexAt(0, v.count)] = val

	return Vec[T]{
		depth: newDepth,
		count: v.count + 1,
		root:  newRoot,
		tail:  newTail,
	}
}

// Conj returns a transient vector with a value appended to the end, invalidating
// the value of the transient vector previously passed in.
func (v TVec[T]) Conj(val T) TVec[T] {
	if v.invalid {
		panic("attempt at operating on invalidated transient vector")
	}

	// Invalidate this transient vector since it will be mutated.
	v.invalid = true

	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < vecNodeWidth {
		// The tail still has space, so just append to it.

		if v.tail == nil {
			v.tail = make([]T, 0, vecNodeWidth)
		}

		return TVec[T]{
			invalid: false,
			depth: v.depth,
			count: v.count + 1,
			root:  v.root,
			tail:  append(v.tail, val),
		}
	} else {
		// There is no room in the tail, so move the tail into the tree.

		var depth = v.depth
		var root = v.root
		var tail = v.tail

		if !isDeepEnoughToAppend(v.depth, v.count) {
			// No space left in the current tree, so deepen the tree one level
			// with a new root node to contain the old root.
			depth = v.depth + 1
			root = &vecNode[T]{}
			root.children = make([]*vecNode[T], vecNodeWidth)
			root.children[0] = v.root
		}

		// Walk through the tree with an indirect pointer to find the location
		// the tail will end up being moved to, making copies of nodes along
		// the path.
		var indirect = &root
		for level := depth; level > 0; level -= 1 {
			if *indirect == nil {
				*indirect = &vecNode[T]{children: make([]*vecNode[T], vecNodeWidth)}
			} else {
				/*
				if (false) {
					var newNode = &vecNode[T]{
						children: make([]*vecNode[T], 0, vecNodeWidth),
						values: make([]T, 0, vecNodeWidth),
					}
					newNode.children = append([]*vecNode[T]{}, (*indirect).children...)
					newNode.values = append([]T{}, (*indirect).values...)
					newNode.children = newNode.children[:vecNodeWidth]
					newNode.values = newNode.values[:vecNodeWidth]
					*indirect = newNode
				} else {
					// Do nothing.
				}
				*/
			}
			indirect = &(*indirect).children[indexAt(level, v.count-1)]
		}
		// Move the old tail into the trie
		*indirect = &vecNode[T]{values: v.tail}

		// Create a new tail for conjugating the new value to.
		tail = make([]T, 0, vecNodeWidth)

		tail = append(tail, val)

		return TVec[T]{
			invalid: false,
			depth: depth,
			count: v.count + 1,
			root:  root,
			tail:  tail,
		}
	}
}

func (v TVec[T]) Persistent() Vec[T] {
	return Vec[T]{
		depth: v.depth,
		count: v.count,
		root: v.root,
		tail: v.tail,
	}
}

func printNode[T any](node *vecNode[T]) {
	fmt.Printf("%#v\n", node)
	if node == nil {
		return
	}
	if node.children != nil {
		fmt.Printf("children {\n")
		for _, child := range node.children {
			printNode(child)
		}
		fmt.Printf("}\n")
	}
	if node.values != nil {
		fmt.Printf("values [")
		for _, val := range node.values {
			fmt.Printf("%v, ", val)
		}
		fmt.Printf("]\n")
	}
}

func (v Vec[T]) Printd() {
	fmt.Printf("%#v\n", v)
	printNode(v.root)
}

// String returns a representation of a vector in the same form as a Go slice
// when using the "%v" formatting verb as in the standard fmt package:
//     With no elements: []
//     With one element: [1]
//     With more than one element: [1 2 3]
func (v Vec[T]) String() string {
	var s = "["
	for i := 0; i < v.count; i += 1 {
		if i == 0 {
			s += fmt.Sprintf("%v", v.Nth(i))
		} else {
			s += fmt.Sprintf(" %v", v.Nth(i))
		}
	}
	s += "]"

	return s
}
