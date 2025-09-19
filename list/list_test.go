package list_test

import (
	"testing"

	"github.com/toddgaunt/persistent/list"
)

func TestNew(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5}
	var list = list.New(slice...)

	for i := 0; i < len(slice); i++ {
		if list.First() != slice[i] {
			t.Fatalf("want element %d at index %d, got %d", slice[i], i, list.First())
		}
		list = list.Rest()
	}
}

func TestListIsEmpty(t *testing.T) {
	var empty = list.New[int]()
	if !list.IsEmpty(empty) {
		t.Fatalf("want empty list, got %v", empty)
	}
}

func TestListLen(t *testing.T) {
	type testCase struct {
		title string
		list  list.List[int]
		want  int
	}

	testCases := []testCase{
		{"Empty", list.New[int](), 0},
		{"SingleElement", list.New(42), 1},
		{"MultipleElements", list.New(1, 2, 3), 3},
	}

	for _, tc := range testCases {
		tc := tc
		f := func(t *testing.T) {
			if got, want := tc.list.Len(), tc.want; got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		}
		t.Run(tc.title, f)
	}
}

func TestListConj(t *testing.T) {
	type testCase struct {
		title string
		list  list.List[string]
		want  list.List[string]
	}

	var empty = list.New[string]()
	var world = empty.Conj("world")
	testCases := []testCase{
		{"Empty", empty, list.New[string]()},
		{"ConjValue", empty.Conj("world"), list.New("world")},
		{"ConjBranchOne", world.Conj("hello"), list.New("hello", "world")},
		{"ConjBranchTwo", world.Conj("goodbye"), list.New("goodbye", "world")},
	}

	for _, tc := range testCases {
		tc := tc
		f := func(t *testing.T) {
			if got, want := tc.list, tc.want; !list.Equal(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
		t.Run(tc.title, f)
	}
}

func TestEqual(t *testing.T) {
	type testCase struct {
		title string
		a     list.List[int]
		b     list.List[int]
		want  bool
	}

	testCases := []testCase{
		{"Empty", list.New[int](), list.New[int](), true},
		{"SingleElementEqual", list.New(42), list.New(42), true},
		{"SingleElementDiffer", list.New(42), list.New(41), false},
		{"MultipleElementsEqual", list.New(1, 2, 3), list.New(1, 2, 3), true},
		{"MultipleElementsDiffer", list.New(1, 2, 3), list.New(2, 3, 1), false},
	}

	for _, tc := range testCases {
		tc := tc
		f := func(t *testing.T) {
			if got, want := list.Equal(tc.a, tc.b), tc.want; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
		t.Run(tc.title, f)
	}
}
