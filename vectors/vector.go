// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package vectors

import "fmt"

// These constants determine the maximum width of vector nodes
const nodeBits = 2
const nodeWidth = 1 << nodeBits
const nodeMask = nodeWidth - 1

// indexAt extracts the bits from i that are needed to index a node at a given
// level in the tree.
func indexAt(level, i int) int {
	return (i >> (level * nodeBits)) & nodeMask
}

// tailOffset returns the total number of elements within a Vec minus the tail.
func tailOffset[T any](count int, tail []T) int {
	return count - len(tail)
}

// isDeepEnoughToAppend evaluates a depth for a Vec to be deep enough for a
// given count. Returns true if a Vec of depth can be appended to without
// creating a new root, otherwise returns false.
func isDeepEnoughToAppend(depth, count int) bool {
	return (count >> nodeBits) <= (1 << depth)
}

// findValues returns the slice of values within the vector which contains the
// value i is associated with.
func findValues[T any](count, depth int, root *node[T], tail []T, index int) []T {
	if index < 0 || index >= count {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, count))
	}

	if index >= tailOffset(count, tail) {
		return tail
	}

	// The index is not associated with the tail, so do a slow lookup for the
	// node it is associated with.
	var walk = root
	for level := depth; level > 0; level -= 1 {
		walk = walk.nodes[indexAt(level, index)]
	}

	return walk.values
}

type node[T any] struct {
	children any
	nodes    []*node[T]
	values   []T
}

func cloneNode[T any](original *node[T]) *node[T] {
	if original == nil {
		return nil
	}

	var clone = &node[T]{
		nodes:  make([]*node[T], len(original.nodes)),
		values: make([]T, len(original.values)),
	}

	copy(clone.nodes, original.nodes)
	copy(clone.values, original.values)

	return clone
}

// Vec is a persistent vector.
type Vec[T any] struct {
	count int      // Number of elements in this vector
	depth int      // Depth of the tree under root
	root  *node[T] // Root of the tree; Contains either child nodes or elements
	tail  []T      // Quickly access elements at the end of the vector
}

// New creates a new persistent vector constructed from the values provided.
func New[T any](vals ...T) Vec[T] {
	var v TVec[T]

	for i := 0; i < len(vals); i++ {
		v = v.Conj(vals[i])
	}

	return v.Persistent()
}

// Len returns the number of values in v
func (v Vec[T]) Len() int {
	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v Vec[T]) Nth(i int) T {
	return findValues(v.count, v.depth, v.root, v.tail, i)[indexAt(0, i)]
}

// Peek returns the last value from a vector.
func (v Vec[T]) Peek() T {
	return v.Nth(v.count - 1)
}

// Assoc creates a new vector that contains val at the location indexed by key.
// The key must be greater than zero and less than v.Len().
func (v Vec[T]) Assoc(index int, value T) Vec[T] {
	if index < 0 || index >= v.count {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, v.count))
	}

	if index >= tailOffset(v.count, v.tail) {
		// The value to update is in the tail, so make a copy of the tail
		var newTail = make([]T, len(v.tail))
		copy(newTail, v.tail)
		newTail[indexAt(0, index)] = value

		return Vec[T]{
			depth: v.depth,
			count: v.count,
			root:  v.root,
			tail:  newTail,
		}
	}

	// Walk through the tree, cloning the path to the updated node.
	var newRoot = cloneNode(v.root)
	var walk = newRoot
	for level := v.depth; level > 0; level -= 1 {
		var i = indexAt(level, index)
		walk.nodes[i] = cloneNode(walk.nodes[i])
		walk = walk.nodes[i]
	}
	walk.values[indexAt(0, index)] = value

	return Vec[T]{
		depth: v.depth,
		count: v.count,
		root:  newRoot,
		tail:  v.tail,
	}
}

// Conj creates a new vector with a value appended to the end.
func (v Vec[T]) Conj(val T) Vec[T] {
	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < nodeWidth {
		// The tail can still be grown, so make a copy to add the new value to.
		var newTail = make([]T, len(v.tail))
		copy(newTail, v.tail)
		newTail = append(newTail, val)

		return Vec[T]{
			depth: v.depth,
			count: v.count + 1,
			root:  v.root,
			tail:  newTail,
		}
	}

	var newDepth = v.depth
	var newRoot = v.root

	// There is no room in the tail, so move the tail into the tree.
	if !isDeepEnoughToAppend(v.depth, v.count) {
		// No space left in the current tree, so deepen the tree one level
		// with a new node containing the old root.
		newDepth = v.depth + 1
		newRoot = &node[T]{}
		newRoot.nodes = make([]*node[T], nodeWidth)
		newRoot.nodes[0] = v.root
	}

	// Walk through the tree with an indirect pointer to find location the
	// tail will end up being moved to, making copies of nodes along the
	// path so that other vector references aren't mutated.
	var indirect = &newRoot
	for level := newDepth; level > 0; level -= 1 {
		if *indirect == nil {
			*indirect = &node[T]{nodes: make([]*node[T], nodeWidth)}
		} else {
			*indirect = cloneNode(*indirect)
		}
		indirect = &(*indirect).nodes[indexAt(level, v.count-1)]
	}
	// Move the old tail as a new node into the trie. Since it has a new path,
	// other vectors sharing this trie won't be affected by this change.
	*indirect = &node[T]{values: v.tail}

	// Create a new tail that contains the newly conjugated value.
	var newTail = []T{val}

	return Vec[T]{
		depth: newDepth,
		count: v.count + 1,
		root:  newRoot,
		tail:  newTail,
	}
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

// TVec is a transient vector. This is similar in structure to a normal
// persistent vector, however it is used in places where persistence isn't
// needed, and more performant operations are required. Each time an operation
// on a TVec is performed, a new one is created using the same memory. The old
// TVec then becomes invalidated so if it is used again a panic occurs.
type TVec[T any] struct {
	invalid bool     // Use when the TVec becomes invalid after a mutation.
	count   int      // Number of elements in this vector
	depth   int      // Depth of the tree under root
	root    *node[T] // Root of the tree containg either child nodes or elements
	tail    []T      // Quickly access elements at the end of the vector
}

func (v TVec[T]) invalidate() {
	if v.invalid {
		panic("attempted operation on an invalid transient vector")
	} else {
		v.invalid = true
	}
}

// Persistent creates a new persistent Vector from a transient vector.
func (v TVec[T]) Persistent() Vec[T] {
	v.invalidate()

	return Vec[T]{
		depth: v.depth,
		count: v.count,
		root:  v.root,
		tail:  v.tail,
	}
}

// Len returns the number of values in v
func (v TVec[T]) Len() int {
	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v TVec[T]) Nth(i int) T {
	return findValues(v.count, v.depth, v.root, v.tail, i)[indexAt(0, i)]
}

// Peek returns the last value from a vector.
func (v TVec[T]) Peek() T {
	return v.Nth(v.count - 1)
}

// String returns a representation of a vector in the same form as a Go slice
// when using the "%v" formatting verb as in the standard fmt package:
//     With no elements: []
//     With one element: [1]
//     With more than one element: [1 2 3]
func (v TVec[T]) String() string {
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

// Assoc returns a transient vector with a value updated at the given index,
// invalidating the transient vector that was operated on.
func (v TVec[T]) Assoc(index int, value T) TVec[T] {
	v.invalidate()
	if index < 0 || index >= v.count {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, v.count))
	}

	if index >= tailOffset(v.count, v.tail) {
		v.tail[indexAt(0, index)] = value
		return TVec[T]{
			invalid: false,
			depth:   v.depth,
			count:   v.count,
			root:    v.root,
			tail:    v.tail,
		}
	}

	// Walk through the tree and update the leaf value found.
	var walk = v.root
	for level := v.depth; level > 0; level -= 1 {
		walk = walk.nodes[indexAt(level, index)]
	}
	walk.values[indexAt(0, index)] = value

	return TVec[T]{
		invalid: false,
		depth:   v.depth,
		count:   v.count,
		root:    v.root,
		tail:    v.tail,
	}
}

// Conj returns a transient vector with a value appended to the end,
// invalidating the transient vector operated on.
func (v TVec[T]) Conj(val T) TVec[T] {
	v.invalidate()

	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < nodeWidth {
		// The tail still has space, so just append to it.

		return TVec[T]{
			invalid: false,
			depth:   v.depth,
			count:   v.count + 1,
			root:    v.root,
			tail:    append(v.tail, val),
		}
	}

	// There is no room in the tail, so move the tail into the tree.

	var newDepth = v.depth
	var newRoot = v.root

	if !isDeepEnoughToAppend(v.depth, v.count) {
		// No space left in the current tree, so deepen the tree one level
		// with a new root node to contain the old root.
		newDepth = v.depth + 1
		newRoot = &node[T]{}
		newRoot.nodes = make([]*node[T], nodeWidth)
		newRoot.nodes[0] = v.root
	}

	// Walk through the tree with an indirect pointer to find the location the
	// tail will end up being moved to.
	var indirect = &newRoot
	for level := newDepth; level > 0; level -= 1 {
		if *indirect == nil {
			*indirect = &node[T]{nodes: make([]*node[T], nodeWidth)}
		}
		indirect = &(*indirect).nodes[indexAt(level, v.count-1)]
	}
	// Move the old tail into the trie
	*indirect = &node[T]{values: v.tail}

	// Create a new tail for conjugating the new value to.
	var newTail = make([]T, 0, nodeWidth)
	newTail = append(newTail, val)

	return TVec[T]{
		invalid: false,
		depth:   newDepth,
		count:   v.count + 1,
		root:    newRoot,
		tail:    newTail,
	}
}

func printNode[T any](n *node[T]) {
	fmt.Printf("%#v\n", n)
	if n == nil {
		return
	}
	if n.nodes != nil {
		fmt.Printf("children {\n")
		for _, child := range n.nodes {
			printNode(child)
		}
		fmt.Printf("}\n")
	}
	if n.values != nil {
		fmt.Printf("values [")
		for _, val := range n.values {
			fmt.Printf("%v, ", val)
		}
		fmt.Printf("]\n")
	}
}

func (v Vec[T]) Printd() {
	fmt.Printf("%#v\n", v)
	printNode(v.root)
}
