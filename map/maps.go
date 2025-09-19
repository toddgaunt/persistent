// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package vectors provides a persistent vector similar to the one found in the
// Clojure programming language. The actual implementation uses data structures
// similar to Clojure's implementation as well, though implemented using Go
// idioms and techniques.
package map

import "github.com/toddgaunt/persistent/vectors"

type MapEntry[T any] struct {
}

type Map[T any] vectors.Vector[MapEntry[T]]

// Len TODO
func (m Map[T]) Len() int {
	return 0
}
