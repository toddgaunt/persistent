package persistent_test

import (
	"fmt"
	"testing"

	"bastionburrow.com/persistent"
)

func TestNewVec(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}
	//var slice = []int{1, 2, 3, 4, 5}
	var vec = persistent.NewVec(slice...)
	vec.Printd()

	for i := 0; i < len(slice); i++ {
		if vec.Nth(i) !=  slice[i] {
			t.Fatalf("want element %d at index %d, got %s", slice[i], i, vec.String())
		}
	}
	fmt.Printf("%s\n", vec.String())

	var vec2 = vec.Conj(42)

	fmt.Printf("vec: %s\n", vec.String())
	fmt.Printf("vec2: %s\n", vec2.String())
}

func TestVectorConj(t *testing.T) {
}

func TestVectorAssoc(t *testing.T) {
}
