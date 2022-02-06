package persistent_test

import (
	"testing"

	"bastionburrow.com/persistent"
)

func TestNewVec(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	var vec = persistent.NewVec(slice...)

	for i := 0; i < len(slice); i++ {
		if vec.Nth(i) !=  slice[i] {
			t.Fatalf("want element %d at index %d, got %s", slice[i], i, vec.String())
		}
	}
}

func TestVectorConj(t *testing.T) {
}

func TestVectorAssoc(t *testing.T) {
}
