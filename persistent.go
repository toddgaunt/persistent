// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package persistent provides data structures for Go that that always preserve
// the previous version of themselves when mutating operations are performed on
// them. Since these structures are effectively immutable, this allows for
// liberal memory sharing between instances as operations are perfomed in lieu
// of memory duplication.
package persistent

type Collection[T any] interface {
	Count() int
	Conj(T) T
}

func Empty[T any](c Collection[T]) bool {
	return c.Count() == 0
}
