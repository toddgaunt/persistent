package lists_test

import (
	"testing"

	"github.com/toddgaunt/persistent/lists"
)

func TestNew(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5}
	var list = lists.New(slice...)

	for i := 0; i < len(slice); i++ {
		if list.First() != slice[i] {
			t.Fatalf("want element %d at index %d, got %d", slice[i], i, list.First())
		}
		list = list.Rest()
	}
}

func TestListIsEmpty(t *testing.T) {
	var empty = lists.New[int]()
	if !lists.IsEmpty(empty) {
		t.Fatalf("want empty list, got %v", empty)
	}
}

func TestListLen(t *testing.T) {
	type testCase struct {
		title string
		list  lists.List[int]
		want  int
	}

	testCases := []testCase{
		{"Empty", lists.New[int](), 0},
		{"SingleElement", lists.New(42), 1},
		{"MultipleElements", lists.New(1, 2, 3), 3},
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
		list  lists.List[string]
		want  lists.List[string]
	}

	var empty = lists.New[string]()
	var world = empty.Conj("world")
	testCases := []testCase{
		{"Empty", empty, lists.New[string]()},
		{"ConjValue", empty.Conj("world"), lists.New("world")},
		{"ConjBranchOne", world.Conj("hello"), lists.New("hello", "world")},
		{"ConjBranchTwo", world.Conj("goodbye"), lists.New("goodbye", "world")},
	}

	for _, tc := range testCases {
		tc := tc
		f := func(t *testing.T) {
			if got, want := tc.list, tc.want; !lists.Equal(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
		t.Run(tc.title, f)
	}
}

func TestEqual(t *testing.T) {
	type testCase struct {
		title string
		a     lists.List[int]
		b     lists.List[int]
		want  bool
	}

	testCases := []testCase{
		{"Empty", lists.New[int](), lists.New[int](), true},
		{"SingleElementEqual", lists.New(42), lists.New(42), true},
		{"SingleElementDiffer", lists.New(42), lists.New(41), false},
		{"MultipleElementsEqual", lists.New(1, 2, 3), lists.New(1, 2, 3), true},
		{"MultipleElementsDiffer", lists.New(1, 2, 3), lists.New(2, 3, 1), false},
	}

	for _, tc := range testCases {
		tc := tc
		f := func(t *testing.T) {
			if got, want := lists.Equal(tc.a, tc.b), tc.want; got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
		t.Run(tc.title, f)
	}
}
