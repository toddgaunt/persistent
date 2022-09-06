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

func cloneTail[T any](tail []T) []T {
	var newTail = make([]T, len(tail))
	copy(newTail, tail)
	return newTail
}

type id int

var persistent *id = nil

func newID() *id {
	return new(id)
}

type node[T any] struct {
	// id indicates if a node was made by transient vector if it is not zero.
	id     *id
	nodes  []*node[T]
	values []T
}

func newNode[T any](id *id) *node[T] {
	return &node[T]{
		id:    id,
		nodes: make([]*node[T], nodeWidth),
	}
}

func newLeaf[T any](id *id, values []T) *node[T] {
	return &node[T]{
		id:     id,
		values: values[:],
	}
}

func cloneNode[T any](id *id, original *node[T]) *node[T] {
	if original == nil {
		return nil
	}

	var clone = &node[T]{
		id:     id,
		nodes:  make([]*node[T], len(original.nodes)),
		values: make([]T, len(original.values)),
	}

	copy(clone.nodes, original.nodes)
	copy(clone.values, original.values)

	return clone
}

// Vector is a persistent vector.
type Vector[T any] struct {
	count int      // Number of elements in this vector
	depth int      // Depth of the tree under root
	root  *node[T] // Root of the tree; Contains either child nodes or elements
	tail  []T      // Quickly access elements at the end of the vector
}

// New creates a new persistent vector constructed from the values provided.
func New[T any](vals ...T) Vector[T] {
	var v TransientVector[T]

	for i := 0; i < len(vals); i++ {
		v = v.Conj(vals[i])
	}

	return v.Persistent()
}

func (v Vector[T]) Transient() TransientVector[T] {
	return TransientVector[T]{
		invalid: false,
		count:   v.count,
		depth:   v.depth,
		root:    cloneNode(newID(), v.root),
		tail:    cloneTail(v.tail),
	}
}

// Len returns the number of values in v
func (v Vector[T]) Len() int {
	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v Vector[T]) Nth(index int) T {
	return findValues(v.count, v.depth, v.root, v.tail, index)[indexAt(0, index)]
}

// Peek returns the last value from a vector.
func (v Vector[T]) Peek() T {
	return v.Nth(v.count - 1)
}

// Assoc creates a new vector that contains val at the location indexed by key.
// The key must be greater than zero and less than v.Len().
func (v Vector[T]) Assoc(index int, value T) Vector[T] {
	if index < 0 || index >= v.count {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, v.count))
	}

	if index >= tailOffset(v.count, v.tail) {
		// The value to update is in the tail, so make a copy of the tail
		var newTail = cloneTail(v.tail)
		newTail[indexAt(0, index)] = value

		return Vector[T]{
			depth: v.depth,
			count: v.count,
			root:  v.root,
			tail:  newTail,
		}
	}

	// Walk through the tree, cloning the path to the updated node.
	var newRoot = cloneNode(persistent, v.root)
	var walk = newRoot
	for level := v.depth; level > 0; level -= 1 {
		var i = indexAt(level, index)
		walk.nodes[i] = cloneNode(persistent, walk.nodes[i])
		walk = walk.nodes[i]
	}
	walk.values[indexAt(0, index)] = value

	return Vector[T]{
		depth: v.depth,
		count: v.count,
		root:  newRoot,
		tail:  v.tail,
	}
}

// Conj creates a new vector with a value appended to the end.
func (v Vector[T]) Conj(val T) Vector[T] {
	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < nodeWidth {
		// The tail can still be grown, so make a copy to add the new value to.
		var newTail = cloneTail(v.tail)

		return Vector[T]{
			depth: v.depth,
			count: v.count + 1,
			root:  v.root,
			tail:  append(newTail, val),
		}
	}

	var newDepth = v.depth
	var newRoot = v.root

	// There is no room in the tail, so move the tail into the tree.
	if !isDeepEnoughToAppend(v.depth, v.count) {
		// No space left in the current tree, so deepen the tree one level
		// with a new node containing the old root.
		newDepth = v.depth + 1
		newRoot = newNode[T](persistent)
		newRoot.nodes[0] = v.root
	}

	// Walk through the tree with an indirect pointer to find location the
	// tail will end up being moved to, making copies of nodes along the
	// path so that other vector references aren't mutated.
	var indirect = &newRoot
	for level := newDepth; level > 0; level -= 1 {
		if *indirect == nil {
			*indirect = newNode[T](persistent)
		} else {
			*indirect = cloneNode(persistent, *indirect)
		}
		indirect = &(*indirect).nodes[indexAt(level, v.count-1)]
	}
	// Move the old tail as a new node into the trie. Since it has a new path,
	// other vectors sharing this trie won't be affected by this change.
	*indirect = newLeaf(persistent, v.tail)

	// Create a new tail that contains the conjugated value.
	var newTail = []T{val}

	return Vector[T]{
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
func (v Vector[T]) String() string {
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

// TransientVector is a transient vector. This is similar in structure to a normal
// persistent vector, however it is used in places where persistence isn't
// needed, and more performant operations are required. Each time an operation
// on a TransientVector is performed, a new one is created using the same memory. The old
// TransientVector then becomes invalidated so if it is used again a panic occurs.
type TransientVector[T any] struct {
	id      *id      // Used to ensure transients mutate only nodes with their unique ID.
	invalid bool     // Set to true to after a mutation.
	count   int      // Number of elements in this vector
	depth   int      // Depth of the tree under root
	root    *node[T] // Root of the tree containg either child nodes or elements
	tail    []T      // Quickly access elements at the end of the vector
}

func (v TransientVector[T]) invalidate() {
	if v.invalid {
		panic("attempted operation on an invalid transient vector")
	} else {
		v.invalid = true
	}
}

// Persistent creates a new persistent Vector from a transient vector.
func (v TransientVector[T]) Persistent() Vector[T] {
	v.invalidate()

	return Vector[T]{
		depth: v.depth,
		count: v.count,
		root:  cloneNode(persistent, v.root),
		tail:  cloneTail(v.tail),
	}
}

// Len returns the number of values in v
func (v TransientVector[T]) Len() int {
	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v TransientVector[T]) Nth(index int) T {
	return findValues(v.count, v.depth, v.root, v.tail, index)[indexAt(0, index)]
}

// Peek returns the last value from a vector.
func (v TransientVector[T]) Peek() T {
	return v.Nth(v.count - 1)
}

// String returns a representation of a vector in the same form as a Go slice
// when using the "%v" formatting verb as in the standard fmt package:
//     With no elements: []
//     With one element: [1]
//     With more than one element: [1 2 3]
func (v TransientVector[T]) String() string {
	if v.invalid {
		panic("attempted operation on an invalid transient vector")
	}

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
func (v TransientVector[T]) Assoc(index int, value T) TransientVector[T] {
	if index < 0 || index >= v.count {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, v.count))
	}

	v.invalidate()

	if index >= tailOffset(v.count, v.tail) {
		v.tail[indexAt(0, index)] = value
		return TransientVector[T]{
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
		var i = indexAt(level, index)
		if walk.nodes[i].id == persistent {
			walk.nodes[i] = cloneNode(v.id, walk.nodes[i])
		}
		walk = walk.nodes[i]
	}
	walk.values[indexAt(0, index)] = value

	return TransientVector[T]{
		invalid: false,
		depth:   v.depth,
		count:   v.count,
		root:    v.root,
		tail:    v.tail,
	}
}

// Conj returns a transient vector with a value appended to the end,
// invalidating the transient vector operated on.
func (v TransientVector[T]) Conj(val T) TransientVector[T] {
	v.invalidate()

	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < nodeWidth {
		// The tail still has space, so just append to it.

		return TransientVector[T]{
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
		newRoot = newNode[T](v.id)
		newRoot.nodes[0] = v.root
	}

	// Walk through the tree with an indirect pointer to find the location the
	// old tail will end up being moved to, then move it in as a value node.
	var indirect = &newRoot
	for level := newDepth; level > 0; level -= 1 {
		if *indirect == nil {
			*indirect = newNode[T](v.id)
		}
		indirect = &(*indirect).nodes[indexAt(level, v.count-1)]
	}
	*indirect = newLeaf(v.id, v.tail)

	// Create a new tail for conjugating the new value to.
	var newTail = []T{val}

	return TransientVector[T]{
		invalid: false,
		depth:   newDepth,
		count:   v.count + 1,
		root:    newRoot,
		tail:    newTail,
	}
}
