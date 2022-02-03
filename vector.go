// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package persistent

import "fmt"

// These constants determine the maximum width of vector nodes
const nodeBits = 5
const nodeWidth = 1 << nodeBits
const nodeMask = nodeWidth - 1

type vecNode struct {
	children []*vecNode
	values   []int
}

type Vector struct {
	count int      // Number of elements in this vector
	depth int      // Depth of the tree under root
	root  *vecNode // Root of the tree; Contains either child nodes or elements
	tail  []int    // Quickly access elements at the end of the vector
}

// idx extracts the bits from i that are needed to index a node at a given
// level in the tree.
func idx(i, level int) int {
	return (i >> (level * nodeBits)) & nodeMask
}

// tailOffset returns the total number of elements within a Vector minus the tail.
func (v *Vector) tailOffset() int {
	return v.count - len(v.tail)
}

// isDeepEnough returns true if the tree within the vector is deep enough for
// another element to be added given the amount of elements it contains.
// Otherwise false is returned when a new node needs to be added in order to
// make room.
func isDeepEnough(length, shift int) bool {
	return (length >> nodeBits) <= (1 << shift)
}

// findValues returns the slice of values within the vector which contains the
// value i is associated with.
func (v Vector) findValues(i int) []int {
	if i < 0 || i >= v.count {
		panic("index out of bounds")
	}

	if i >= v.tailOffset() {
		return v.tail
	}

	// The index is not associated with the tail, so do a slow lookup for the
	// node it is associated with.
	walk := v.root
	for level := v.depth; level > 0; level -= 1 {
		walk = walk.children[idx(i, level)]
	}

	return walk.values
}

// String returns a string representation of a vector in the same form as a Go slice:
//     With no elements: []
//     With one element: [1]
//     With more than one element: [1, 2, 3]
func (v *Vector) String() string {
	s := "["
	for i := 0; i < v.count; i += 1 {
		if i == 0 {
			s += fmt.Sprintf("%d", v.Nth(i))
		} else {
			s += fmt.Sprintf(", %d", v.Nth(i))
		}
	}
	s += "]"

	return s
}

func (v Vector) Count() int {
	return v.count
}

// Nth returns from the vector the value at the index provided. The index must
// be greater than zero and less than v.count.
func (v Vector) Nth(i int) int {
	return v.findValues(i)[idx(i, 0)]
}

// Peek returns the last value from a vector.
func (v Vector) Peek() int {
	return v.Nth(v.count - 1)
}

// Assoc creates a new vector that contains val at the location indexed by key.
// The key must be greater than zero and less than v.Count().
func (v Vector) Assoc(key int, val int) Vector {
	if key < 0 || key >= v.count {
		panic("index out of bounds")
	}

	newRoot := v.root
	newTail := v.tail

	var leaf []int
	// Either the tail is being updated, or a node in the tree is
	if key >= v.tailOffset() {
		// The value to update is in the tail, so make a copy of the tail
		newTail = make([]int, len(v.tail))
		copy(newTail, v.tail)
		leaf = newTail
	} else {
		// The value to update is in the tree, so create a new path of nodes

		// Clone the root node first so the changes to the path don't effect
		// the old vector
		newRoot = &vecNode{}
		newRoot.children = append([]*vecNode{}, v.root.children...)
		newRoot.values = append([]int{}, v.root.values...)

		walk := newRoot
		for level := v.depth; level > 0; level -= 1 {
			oldNode := walk.children[idx(key, level)]

			walk.children[idx(key, level)] = &vecNode{}
			walk.children = append([]*vecNode{}, oldNode.children...)
			walk.values = append([]int{}, oldNode.values...)

			walk = walk.children[idx(key, level)]
		}
		leaf = walk.values
	}

	// Update the value
	leaf[idx(key, 0)] = val

	return Vector{
		depth: v.depth,
		count: v.count,
		root:  newRoot,
		tail:  newTail,
	}
}

// Conj creates a new vector with a value appended to the end.
func (v Vector) Conj(value int) Vector {

	newDepth := v.depth
	newRoot := v.root
	var newTail []int

	// Either the tail is being appended to, or a node in the tree is.
	if len(v.tail) < nodeWidth {
		// The tail can still be grown, so make a copy to add the new value to.
		newTail = make([]int, len(v.tail)+1)
		copy(newTail, v.tail)
	} else {
		// There is no room in the tail, so move the tail into the tree.
		if !isDeepEnough(v.count, v.depth) {
			// No space left in the current tree, so deepen the tree one level
			// with a new node containing the old root.
			newDepth = v.depth + 1
			newRoot = &vecNode{}
			newRoot.children = []*vecNode{v.root}
		}

		// Walk through the tree with an indirect pointer to find location the
		// tail will end up being moved to, making copies of nodes along the path.
		indirect := &newRoot
		for level := newDepth; level > 0; level -= 1 {
			if *indirect == nil {
				*indirect = &vecNode{}
			} else {
				newNode := &vecNode{}
				newNode.children = append([]*vecNode{}, (*indirect).children...)
				newNode.values = append([]int{}, (*indirect).values...)

				*indirect = newNode
			}
			indirect = &(*indirect).children[idx(v.count-1, level)]
		}
		// Move the old tail into the trie
		*indirect = &vecNode{values: v.tail}

		// Create a new tail for conjugating the new value to.
		newTail = make([]int, 1)
	}
	newTail[idx(v.count, 0)] = value

	return Vector{
		depth: newDepth,
		count: v.count + 1,
		root:  newRoot,
		tail:  newTail,
	}
}
