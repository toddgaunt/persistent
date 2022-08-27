package vectors_test

import (
	"fmt"
	"testing"

	"bastionburrow.com/persistent/vectors"
)

func TestVectorScratch(t *testing.T) {
	/*
		var slice = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}
		//var slice = []int{1, 2, 3, 4, 5}
		var vec = vectors.New(slice...)
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
	*/
}

func TestVecNth(t *testing.T) {
	var slice = []int{
		11, 12, 13, 14, 15, 16, 17, 18,
		21, 22, 23, 24, 25, 26, 27, 28,
		31, 32, 33, 34, 35, 36, 37, 38,
		41, 42, 43, 44, 45, 46, 47, 48,
		51, 52, 53, 54, 55, 56, 57, 58,
		61, 62, 63, 64, 65, 66, 67, 68,
		71, 72, 73, 74, 75, 76, 77, 78,
		81, 82, 83, 84, 85, 86, 87, 88,
	}

	var vec = vectors.New(slice...)

	for i := 0; i < len(slice); i++ {
		if vec.Nth(i) != slice[i] {
			t.Fatalf("want element %d at index %d, got %d", slice[i], i, vec.Nth(i))
		}
	}
}

func TestVectorConj(t *testing.T) {
}

func TestVectorAssoc(t *testing.T) {
}

func TestVectorString(t *testing.T) {
	type testStruct struct {
		name string
		num  int
		x    float64
		y    float64
	}

	var intSlice = []int{1, 2, 3, 4, 5}
	var stringSlice = []string{"hello", " ", "world"}
	var structSlice = []testStruct{
		{"one", 1, 1.0, 1.0},
		{"Adams", 42, 3.14, 2.71},
		{"Jdoe", 185, 6.2, 14},
	}

	var intVec = vectors.New(intSlice...)
	var stringVec = vectors.New(stringSlice...)
	var structVec = vectors.New(structSlice...)

	if got, want := fmt.Sprintf("%v", intSlice), intVec.String(); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	if got, want := fmt.Sprintf("%v", stringSlice), stringVec.String(); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	if got, want := fmt.Sprintf("%v", structSlice), structVec.String(); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
