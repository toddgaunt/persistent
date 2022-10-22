package vectors_test

import (
	"fmt"
	"testing"

	"github.com/toddgaunt/persistent/vectors"
)

var testSlice = func() []int {
	var slice = make([]int, 65)
	for i := 0; i < len(slice); i++ {
		slice[i] = i + 1
	}
	return slice
}()

func TestVectorNth(t *testing.T) {
	var vec = vectors.New(testSlice...)

	for i := 0; i < len(testSlice); i++ {
		if vec.Nth(i) != testSlice[i] {
			t.Fatalf("want element %d at index %d, got %d", testSlice[i], i, vec.Nth(i))
		}
	}
}

func TestVectorAssoc(t *testing.T) {
	var testCases = []struct {
		name   string
		slice  []int
		index  int
		value  int
		panics bool
	}{
		{
			name:   "AssocTrie",
			slice:  make([]int, 32+32),
			index:  1,
			value:  42,
			panics: false,
		},
		{
			name:   "AssocDeepTrie",
			slice:  make([]int, 32*32+32),
			index:  1,
			value:  42,
			panics: false,
		},
		{
			name:   "AssocTail",
			slice:  make([]int, 32),
			index:  31,
			value:  42,
			panics: false,
		},
		{
			name:   "AssocEmpty",
			slice:  []int{},
			index:  0,
			value:  0,
			panics: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					if !tc.panics {
						t.Fatalf("got panic %v when none was expected", r)
					}
				} else {
					if tc.panics {
						t.Fatalf("got nil panic when one was expected")
					}
				}
			}()

			var vec1 = vectors.New(tc.slice...)
			var vec2 = vec1.Assoc(tc.index, tc.value)
			if got, want := vec1.Nth(tc.index), tc.slice[0]; got != want {
				t.Fatalf("got vec1.Nth(index)=%v, want vec1.Nth(index)=%v", got, want)
			}
			if got, want := vec2.Nth(tc.index), tc.value; got != want {
				t.Fatalf("got vec2.Nth(index)=%v, want vec2.Nth(index)=%v", got, want)
			}
			if got, want := vec1.Len(), len(tc.slice); got != want {
				t.Fatalf("got vec1.Len()=%v, want vec1.Len()=%v", got, want)
			}
			if got, want := vec2.Len(), len(tc.slice); got != want {
				t.Fatalf("got vec2.Len()=%v, want vec2.Len()=%v", got, want)
			}
		})
	}
}

func TestVectorConj(t *testing.T) {
	var testCases = []struct {
		name  string
		slice []int
		value int
	}{
		{
			name:  "ConjTrie",
			slice: make([]int, 32+32),
			value: 42,
		},
		{
			name:  "ConjDeepTrie",
			slice: make([]int, 32*32+32),
			value: 42,
		},
		{
			name:  "ConjTail",
			slice: make([]int, 32),
			value: 42,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var vec1 = vectors.New(tc.slice...)
			var vec2 = vec1.Conj(tc.value)
			if vec1.Len() > 0 {
				if got, want := vec1.Peek(), tc.slice[len(tc.slice)-1]; got != want {
					t.Fatalf("got vec1.Peek()=%v, want vec1.Peek()=%v", got, want)
				}
			}
			if got, want := vec2.Peek(), tc.value; got != want {
				t.Fatalf("got vec2.Peek()=%v, want vec2.Peek()=%v", got, want)
			}
			if got, want := vec1.Len(), len(tc.slice); got != want {
				t.Fatalf("got vec1.Len()=%v, want vec1.Len()=%v", got, want)
			}
			if got, want := vec2.Len(), len(tc.slice)+1; got != want {
				t.Fatalf("got vec2.Len()=%v, want vec2.Len()=%v", got, want)
			}
		})
	}
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

func TestTransientVectorAssoc(t *testing.T) {
	var vec = vectors.New(testSlice...)
	var want = vec.Nth(0)

	var tvec = vec.Transient()
	tvec.Assoc(0, 42)
	if got := vec.Nth(0); got != want {
		t.Fatalf("got vec.Nth(0)=%d, want vec.Nth(0)=%d", got, want)
	}
}

func TestTransientVectorConj(t *testing.T) {
	var vec = vectors.New(testSlice...)
	var want = vec.Nth(vec.Len() - 1)

	var tvec = vec.Transient()
	tvec.Conj(42)
	if got := vec.Nth(vec.Len() - 1); got != want {
		t.Fatalf("got vec.Nth(vec.Len()-1)=%d, want vec.Nth(vec.Len()-1)=%d", got, want)
	}
}

func FuzzVectorNth(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		var vec = vectors.New(b...)
		for i := 0; i < vec.Len(); i++ {
			if vec.Nth(i) != b[i] {
				t.Fatalf("want element %d at index %d, got %d", b[i], i, vec.Nth(i))
			}
		}
	})
}

func FuzzVectorAssoc(f *testing.F) {
	f.Fuzz(func(t *testing.T, init []byte, index int) {
		defer func() {
			r := recover()
			if r != nil {
				if r != fmt.Sprintf("index out of range [%d] with length %d", index, len(init)) {
					t.Fatalf("got panic %v, want out of bounds error", r)
				}
			}
		}()

		var vec = vectors.New(init...)

		var originalValue = vec.Nth(index)
		var value = originalValue + 1

		var result = vec.Assoc(index, value)

		if got, want := vec.Len(), result.Len(); got != want {
			t.Fatalf("got vec.Len() == %d != result.Len(), want vec.Len() == %d == result.Len()", got, want)
		}
		if got := vec.Nth(index); got != originalValue {
			t.Fatalf("got vec.Nth(index) == %v, want vec.Nth(index) == %v", got, originalValue)
		}
		if got, want := result.Nth(index), value; got != want {
			t.Fatalf("got result.Nth(index) == %v, result.Nth(index) == %v", got, want)
		}
	})
}

func FuzzVectorConj(f *testing.F) {
	f.Add([]byte{}, byte(0))
	f.Add(
		[]byte{1, 2, 3, 4},
		byte(5),
	)
	f.Add(
		[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		byte(33),
	)
	f.Fuzz(func(t *testing.T, init []byte, value byte) {
		var vec = vectors.New(init...)
		var result = vec.Conj(value)

		if got, want := vec.Len(), len(init); got != want {
			t.Fatalf("expected source to not be modified, got vec.Len() == %d, want vec.Len() == %d", got, want)
		}

		if got, want := result.Len(), vec.Len()+1; got != want {
			t.Fatalf("expected result one elem longer, got result.Len() == %d, want result.Len() == %d", got, want)
		}
	})
}

func FuzzTransientVectorNth(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		var vec = vectors.New(b...)
		var tvec = vec.Transient()
		for i := 0; i < tvec.Len(); i++ {
			if tvec.Nth(i) != b[i] {
				t.Fatalf("want element %d at index %d, got %d", b[i], i, tvec.Nth(i))
			}
		}
	})
}

func FuzzTransientVectorAssoc(f *testing.F) {
	f.Fuzz(func(t *testing.T, init []byte, index int) {
		defer func() {
			r := recover()
			if r != nil {
				if r != fmt.Sprintf("index out of range [%d] with length %d", index, len(init)) {
					t.Fatalf("got panic %v, want out of bounds error", r)
				}
			}
		}()

		var vec = vectors.New(init...)

		var originalValue = vec.Nth(index)
		var value = originalValue + 1

		var tvec = vec.Transient()

		var result = tvec.Assoc(index, value)

		if got, want := vec.Len(), result.Len(); got != want {
			t.Fatalf("got vec.Len() == %d != result.Len(), want vec.Len() == %d == result.Len()", got, want)
		}
		if got := vec.Nth(index); got != originalValue {
			t.Fatalf("got vec.Nth(index) == %v, want vec.Nth(index) == %v", got, originalValue)
		}
		if got, want := result.Nth(index), value; got != want {
			t.Fatalf("got result.Nth(index) == %v, result.Nth(index) == %v", got, want)
		}
	})
}

func FuzzTransientVectorConj(f *testing.F) {
	f.Add([]byte{}, byte(0))
	f.Add(
		[]byte{1, 2, 3, 4},
		byte(5),
	)
	f.Add(
		[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		byte(33),
	)
	f.Fuzz(func(t *testing.T, init []byte, value byte) {
		var vec = vectors.New(init...)
		var tvec = vec.Transient()
		var result = tvec.Conj(value)

		if got, want := vec.Len(), len(init); got != want {
			t.Fatalf("expected source to not be modified, got vec.Len() == %d, want vec.Len() == %d", got, want)
		}

		if got, want := result.Len(), vec.Len()+1; got != want {
			t.Fatalf("expected result one elem longer, got result.Len() == %d, want result.Len() == %d", got, want)
		}
	})
}

var benchmarkCases = []int{
	100,
	1000,
	10000,
	100000,
	1000000,
}

func newBenchmarkVec(n int) vectors.Vector[int] {
	init := make([]int, 0, n)
	for i := 0; i < n; i++ {
		init = append(init, i+1)
	}
	return vectors.New(init...)
}

func newBenchmarkTransientVector(n int) vectors.TransientVector[int] {
	vec := vectors.TransientVector[int]{}
	for i := 0; i < n; i++ {
		vec = vec.Conj(i + 1)
	}
	return vec
}

func newBenchmarkGoNative(n int) []int {
	slice := make([]int, n)
	for i := 0; i < n; i++ {
		slice[i] = i + 1
	}
	return slice
}

func BenchmarkNthPersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkVec(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Nth(i % n)
			}
		})
	}
}

func BenchmarkNthTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		tvec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tvec.Nth(i % n)
			}
		})
	}
}

func BenchmarkNthGoNative(b *testing.B) {
	for _, n := range benchmarkCases {
		slice := newBenchmarkGoNative(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = slice[i%n]
			}
		})
	}
}

func BenchmarkAssocPersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			vec := newBenchmarkVec(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				vec = vec.Assoc(i%n, i)
			}
		})
	}
}

func BenchmarkAssocTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			tvec := newBenchmarkTransientVector(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tvec = tvec.Assoc(i%n, i)
			}
		})
	}
}

func BenchmarkAssocGoNative(b *testing.B) {
	for _, n := range benchmarkCases {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			slice := newBenchmarkGoNative(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				slice[i%n] = i
			}
		})
	}
}

func BenchmarkConjPersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec := vectors.New[int]()
				for i := 0; i < n; i++ {
					vec = vec.Conj(i)
				}
			}
		})
	}
}

func BenchmarkConjTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tvec := vectors.New[int]().Transient()
				for i := 0; i < n; i++ {
					tvec = tvec.Conj(i)
				}
			}
		})
	}
}

func BenchmarkConjGoNative(b *testing.B) {
	for _, n := range benchmarkCases {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				slice := []int{}
				for i := 0; i < n; i++ {
					slice = append(slice, i)
				}
			}
		})
	}
}
