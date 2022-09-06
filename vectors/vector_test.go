package vectors_test

import (
	"fmt"
	"testing"

	"bastionburrow.com/persistent/vectors"
)

func TestVectorNth(t *testing.T) {
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
	var vec = vectors.New[int]()
	vec = vec.Conj(2)
	if got, want := vec.Peek(), 2; got != want {
		t.Fatalf("got vec.Peek()=%v, want vec.Peek()=%v", got, want)
	}
	if got, want := vec.Len(), 1; got != want {
		t.Fatalf("got vec.Len()=%v, want vec.Len()=%v", got, want)
	}
}

func TestVectorAssoc(t *testing.T) {
	var vec = vectors.New(1, 2, 3, 4, 5, 6, 7, 8)
	vec = vec.Assoc(5, 42)
	if got, want := vec.Len(), 8; got != want {
		t.Fatalf("got vec.Len()=%v, want vec.Len()=%v", got, want)
	}
	if got, want := vec.Nth(5), 42; got != want {
		t.Fatalf("got vec.Nth(5)=%v, want vec.Nth(5)=%v", got, want)
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

func FuzzVectorNth(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		var vec = vectors.New(b...)
		for i := 0; i < len(b); i++ {
			if vec.Nth(i) != b[i] {
				t.Fatalf("want element %d at index %d, got %d", b[i], i, vec.Nth(i))
			}
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
		if got, want := result.Len(), vec.Len()+1; got != want {
			t.Fatalf("expected conj to make new vector one element longer, got %d, want %d", got, want)
		}
		vec = result
	})
}

func FuzzVectorAssoc(f *testing.F) {
	f.Fuzz(func(t *testing.T, init []byte, index int, value byte) {
		init = append(init, value)
		if index < 0 {
			index = -index
		}
		index = index % len(init)
		var vec = vectors.New(init...)
		var result = vec.Assoc(index, value)
		if got, want := vec.Len(), result.Len(); got != want {
			t.Fatalf("got len %d, want len %d", got, want)
		}
		if got, want := result.Nth(index), value; got != want {
			t.Fatalf("got value %v, want value %v", got, want)
		}
	})
}

var benchmarkCases = []int{
	32*0 + 32,        // Elements are all in tail
	32*1 + 32,        // Depth of 1 and full tail
	32*32 + 32,       // Depth of 2 and full tail
	32*32*32 + 32,    // Depth of 3 and full tail
	32*32*32*32 + 32, // Depth of 4 and full tail
}

func newBenchmarkVec(n int) vectors.Vector[int] {
	slice := make([]int, 0, n)
	for i := 0; i < n; i++ {
		slice = append(slice, i+1)
	}
	return vectors.New(slice...)
}

func BenchmarkNthTriePersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkVec(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Nth(n / 2)
			}
		})
	}
}

func BenchmarkNthTailPersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkVec(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Nth(n - 1)
			}
		})
	}
}

func BenchmarkConjTriePersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkVec(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Conj(42)
			}
		})
	}
}

func BenchmarkConjTailPersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkVec(n - 1)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Conj(42)
			}
		})
	}
}

func BenchmarkAssocTriePersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkVec(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Assoc(n/2, 42)
			}
		})
	}
}

func BenchmarkAssocTailPersistent(b *testing.B) {
	for _, n := range benchmarkCases {
		n = n - 1
		vec := newBenchmarkVec(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Assoc(n-1, 42)
			}
		})
	}
}

func newBenchmarkTransientVector(n int) vectors.TransientVector[int] {
	var vec vectors.TransientVector[int]
	for i := 0; i < n; i++ {
		vec = vec.Conj(i + 1)
	}
	return vec
}

func BenchmarkNthTrieTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Nth(n / 2)
			}
		})
	}
}

func BenchmarkNthTailTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Nth(n - 1)
			}
		})
	}
}

func BenchmarkConjTrieTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Conj(42)
			}
		})
	}
}

func BenchmarkConjTailTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		n = n - 1
		vec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Conj(42)
			}
		})
	}
}

func BenchmarkAssocTrieTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Assoc(n/2, 42)
			}
		})
	}
}

func BenchmarkAssocTailTransient(b *testing.B) {
	for _, n := range benchmarkCases {
		vec := newBenchmarkTransientVector(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vec.Assoc(n-1, 42)
			}
		})
	}
}

// Go slice baseline comparison

func newBenchmarkSlice(n int) []int {
	vec := make([]int, n)
	for i := 0; i < n; i++ {
		vec[i] = i + 1
	}
	return vec
}

func BenchmarkNthTailGoSlice(b *testing.B) {
	for _, n := range benchmarkCases {
		slice := newBenchmarkSlice(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = slice[n-1]
			}
		})
	}
}

func BenchmarkConjTailGoSlice(b *testing.B) {
	for _, n := range benchmarkCases {
		slice := newBenchmarkSlice(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				slice = append(slice, 42)
			}
		})
	}
}

func BenchmarkAssocTailGoSlice(b *testing.B) {
	for _, n := range benchmarkCases {
		slice := newBenchmarkSlice(n)
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				slice[n-1] = 42
			}
		})
	}
}
