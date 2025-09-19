// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package vectors provides a persistent vector similar to the one found in the
// Clojure programming language. The actual implementation uses data structures
// similar to Clojure's implementation as well, though implemented using Go
// idioms and techniques.
package vector

import "fmt"

// These constants determine the maximum width of vector nodes
const nodeBits = 5
const nodeWidth = 1 << nodeBits
const nodeMask = nodeWidth - 1

// indexAt extracts the bits from i that are needed to index a node at a given
// level in the tree.
func indexAt(level, i int) int {
	return (i >> (level * nodeBits)) & nodeMask
}

// indexInTail returns the total number of items within a Vec minus the tail.
func indexInTail[T any](index int, count int, tail []T) bool {
	return index >= (count - len(tail))
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

	if indexInTail(index, count, tail) {
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

	clone := &node[T]{
		id: id,
	}

	if original.nodes != nil {
		clone.nodes = make([]*node[T], len(original.nodes))
		copy(clone.nodes, original.nodes)
	}

	if original.values != nil {
		clone.values = make([]T, len(original.values))
		copy(clone.values, original.values)
	}

	return clone
}

// Vector is a persistent vector. Vector values can be treated as values, which
// means that no operation on a Vector will modify it. Instead a new vector
// will be created each time with the operation applied to using the old vector
// as the base. Vector shares memory between instances so these operations are
// quite fast.
type Vector[T any] struct {
	count int      // Number of items in this vector
	depth int      // Depth of the tree under root
	tail  []T      // Quickly access items at the end of the vector
	root  *node[T] // Root of the tree; Contains either child nodes or items
}

// New creates a new persistent vector constructed from the values provided.
func New[T any](vals ...T) Vector[T] {
	var v = Vector[T]{}.Transient()

	for i := 0; i < len(vals); i++ {
		v = v.Conj(vals[i])
	}

	return v.Persistent()
}

// Transient creates a new transient vector using v as its base
func (v Vector[T]) Transient() TransientVector[T] {
	id := new(id)
	return TransientVector[T]{
		id:      id,
		invalid: false,
		count:   v.count,
		depth:   v.depth,
		tail:    cloneTail(v.tail),
		root:    v.root,
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

	if indexInTail(index, v.count, v.tail) {
		// The value to update is in the tail, so make a copy of the tail
		var newTail = cloneTail(v.tail)
		newTail[indexAt(0, index)] = value

		return Vector[T]{
			depth: v.depth,
			count: v.count,
			tail:  newTail,
			root:  v.root,
		}
	}

	// Create a new root so the original vector isn't changed.
	var newRoot = cloneNode(persistent, v.root)

	// Walk through the tree, cloning the path to the updated node.
	var walk = newRoot
	for level := v.depth; level > 0; level -= 1 {
		var i = indexAt(level, index)
		walk.nodes[i] = cloneNode(persistent, walk.nodes[i])
		walk = walk.nodes[i]
	}
	// Finally, update the value in the leaf node.
	walk.values[indexAt(0, index)] = value

	return Vector[T]{
		depth: v.depth,
		count: v.count,
		tail:  v.tail,
		root:  newRoot,
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
			tail:  append(newTail, val),
			root:  v.root,
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

	// Walk through the tree with an indirect pointer to find location the tail
	// will end up being moved to, creating new nodes along the way as needed.
	var indirect = &newRoot
	for level := newDepth; level > 0; level -= 1 {
		if *indirect == nil {
			*indirect = newNode[T](persistent)
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
		tail:  newTail,
		root:  newRoot,
	}
}

// String returns a representation of a vector in the same form as a Go slice
// when using the "%v" formatting verb as in the standard fmt package:
//		With no items: []
//		With one item: [1]
//		With more than one item: [1 2 3]
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

// TransientVector provides the same API as a persistent vector, however a
// transient vector becomes invalid after any operation that creates a new
// vector from an itself. While transient vectors are similar in structure
// to a persistent vectors, they are meant to be used in places where
// persistence isn't needed, and faster performance for certain operations is
// required. Each time an operation on a TransientVector is performed, a new
// one is created using the same underlying memory. The old TransientVector is
// then marked invalidated so if it is used again with any of the operations
// this package provides, a panic occurs.
type TransientVector[T any] struct {
	// id is used to ensure transients mutate only nodes with their unique ID.
	// This works because a new ID is allocated whenever a transient vector is
	// made which uses a unique pointer address for the ID. This ID is only
	// deallocated when all nodes that reference the id are reclaimed as well.
	// This ensures that as long as a node exists with an already allocated ID,
	// then it won't be allocated by a different transient vector.
	//
	// Also note that the zero value of TransientVector is valid, even though it
	// isn't assigned an ID. This is because:
	//     1. An empty TransientVector can't possibly point to nodes owned by another vector.
	//     2. Once made persistent it's nodes will have a nil id, the same as persistent vectors.
	id      *id
	invalid bool     // Set to true to after a mutation.
	count   int      // Number of items in this vector
	depth   int      // Depth of the tree under root
	tail    []T      // Quickly access items at the end of the vector
	root    *node[T] // Root of the tree containg either child nodes or items
}

func (v TransientVector[T]) ensureValid() {
	if v.invalid {
		panic("attempted operation on an invalid transient vector")
	}
}

func (v TransientVector[T]) invalidate() {
	v.ensureValid()
	v.invalid = true
}

// Persistent creates a new persistent Vector from a transient vector.
func (v TransientVector[T]) Persistent() Vector[T] {
	v.invalidate()

	return Vector[T]{
		depth: v.depth,
		count: v.count,
		tail:  cloneTail(v.tail),
		root:  cloneNode(persistent, v.root),
	}
}

// Len returns the number of values in v
func (v TransientVector[T]) Len() int {
	v.ensureValid()

	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v TransientVector[T]) Nth(index int) T {
	v.ensureValid()

	return findValues(v.count, v.depth, v.root, v.tail, index)[indexAt(0, index)]
}

// Peek returns the last value from a vector.
func (v TransientVector[T]) Peek() T {
	return v.Nth(v.count - 1)
}

// String returns a representation of a vector in the same form as a Go slice
// when using the "%v" formatting verb as in the standard fmt package:
//     With no items: []
//     With one item: [1]
//     With more than one item: [1 2 3]
func (v TransientVector[T]) String() string {
	v.ensureValid()

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
	v.invalidate()

	if index < 0 || index >= v.count {
		panic(fmt.Sprintf("index out of range [%d] with length %d", index, v.count))
	}

	if indexInTail(index, v.count, v.tail) {
		v.tail[indexAt(0, index)] = value
		return TransientVector[T]{
			id:      v.id,
			invalid: false,
			depth:   v.depth,
			count:   v.count,
			tail:    v.tail,
			root:    v.root,
		}
	}

	if v.root.id != v.id {
		// Create a new root so the original vector isn't changed.
		v.root = cloneNode(v.id, v.root)
	}

	// Walk through the tree and update the leaf value found.
	var walk = v.root
	for level := v.depth; level > 0; level -= 1 {
		var i = indexAt(level, index)
		if walk.nodes[i].id != v.id {
			walk.nodes[i] = cloneNode(v.id, walk.nodes[i])
		}
		walk = walk.nodes[i]
	}
	walk.values[indexAt(0, index)] = value

	return TransientVector[T]{
		id:      v.id,
		invalid: false,
		depth:   v.depth,
		count:   v.count,
		tail:    v.tail,
		root:    v.root,
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
			id:      v.id,
			invalid: false,
			depth:   v.depth,
			count:   v.count + 1,
			tail:    append(v.tail, val),
			root:    v.root,
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
		if (*indirect).id != v.id {
			*indirect = cloneNode(v.id, *indirect)
		}
		indirect = &(*indirect).nodes[indexAt(level, v.count-1)]
	}
	*indirect = newLeaf(v.id, v.tail)

	// Create a new tail for conjugating the new value to. Allocate enough
	// space for a full tail up-front to optimize appending new values.
	var newTail = make([]T, 0, nodeWidth)
	newTail = append(newTail, val)

	return TransientVector[T]{
		id:      v.id,
		invalid: false,
		depth:   newDepth,
		count:   v.count + 1,
		tail:    newTail,
		root:    newRoot,
	}
}
