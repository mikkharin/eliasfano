package eliasfano_test

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/mikkharin/eliasfano"
)

const benchN = 1 << 16

func benchValues() []uint64 {
	r := rand.New(rand.NewPCG(42, 1337))
	values := make([]uint64, benchN)
	var cur uint64
	for i := range values {
		cur += 1 + r.Uint64N(1<<10)
		values[i] = cur
	}
	return values
}

func benchIndexes(n int) []int {
	r := rand.New(rand.NewPCG(7, 7))
	idx := make([]int, 1024)
	for i := range idx {
		idx[i] = r.IntN(n)
	}
	return idx
}

var (
	sinkU64 uint64
	sinkInt int
	sinkSeq *eliasfano.Sequence
)

func BenchmarkNew(b *testing.B) {
	values := benchValues()
	b.ReportAllocs()
	b.SetBytes(int64(8 * len(values)))
	for i := 0; i < b.N; i++ {
		s, err := eliasfano.New(values)
		if err != nil {
			b.Fatal(err)
		}
		sinkSeq = s
	}
}

func BenchmarkAt(b *testing.B) {
	s, err := eliasfano.New(benchValues())
	if err != nil {
		b.Fatal(err)
	}
	idx := benchIndexes(s.Len())
	b.ReportAllocs()
	b.ResetTimer()
	var sum uint64
	for i := 0; i < b.N; i++ {
		sum += s.At(idx[i%len(idx)])
	}
	sinkU64 = sum
}

func BenchmarkLowerBound(b *testing.B) {
	values := benchValues()
	s, err := eliasfano.New(values)
	if err != nil {
		b.Fatal(err)
	}
	r := rand.New(rand.NewPCG(9, 9))
	queries := make([]uint64, 1024)
	for i := range queries {
		queries[i] = r.Uint64N(values[len(values)-1] + 1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sum int
	for i := 0; i < b.N; i++ {
		sum += s.LowerBound(queries[i%len(queries)])
	}
	sinkInt = sum
}

// BenchmarkSortSearch is the uncompressed baseline.
func BenchmarkSortSearch(b *testing.B) {
	values := benchValues()
	r := rand.New(rand.NewPCG(9, 9))
	queries := make([]uint64, 1024)
	for i := range queries {
		queries[i] = r.Uint64N(values[len(values)-1] + 1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sum int
	for i := 0; i < b.N; i++ {
		pos, _ := slices.BinarySearch(values, queries[i%len(queries)])
		sum += pos
	}
	sinkInt = sum
}
