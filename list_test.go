package persistent_test

import (
	"testing"

	"bastionburrow.com/persistent"
)

func TestNewList(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5}
	var list = persistent.NewList(slice...)

	for i := 0; i < len(slice); i++ {
		if list.First() != slice[i] {
			t.Fatalf("want element %d at index %d, got %d", slice[i], i, list.First())
		}
		list = list.Rest()
	}
}

func TestListCount(t *testing.T) {
	type testCase struct {
		title string
		list  persistent.List[int]
		want  int
	}

	testCases := []testCase{
		{"Empty", persistent.NewList[int](), 0},
		{"SingleElement", persistent.NewList(42), 1},
		{"MultipleElements", persistent.NewList(1, 2, 3), 3},
	}

	for _, tc := range testCases {
		tc := tc
		f := func(t *testing.T) {
			t.Parallel()
			if got, want := tc.list.Count(), tc.want; got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		}
		t.Run(tc.title, f)
	}
}
