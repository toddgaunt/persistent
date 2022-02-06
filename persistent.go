// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package persistent provides data structures for go that that always
// preserves the previous version of themselves when performing operations on
// them. Since these structures are immutable, this allows for liberal memory
// sharing between instances as they are operated upon.
package persistent

type Collection[T any] interface {
	Count() int
	Conj(T) T
}

func Empty[T any](c Collection[T]) bool {
	return c.Count() == 0
}
