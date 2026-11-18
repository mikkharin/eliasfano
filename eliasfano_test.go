package eliasfano_test

import (
	"bytes"
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/mikkharin/eliasfano"
)

func mustNew(t testing.TB, values []uint64) *eliasfano.Sequence {
	t.Helper()
	s, err := eliasfano.New(values)
	if err != nil {
		t.Fatalf("New(%v): unexpected error: %v", abbreviate(values), err)
	}
	return s
}

func abbreviate(values []uint64) []uint64 {
	if len(values) > 16 {
		return values[:16]
	}
	return values
}

func refLowerBound(values []uint64, x uint64) int {
	return sort.Search(len(values), func(i int) bool { return values[i] >= x })
}

func probes(values []uint64) []uint64 {
	p := []uint64{0, 1, math.MaxUint64 - 1, math.MaxUint64}
	for _, v := range values {
		p = append(p, v)
		if v > 0 {
			p = append(p, v-1)
		}
		if v < math.MaxUint64 {
			p = append(p, v+1)
		}
	}
	return p
}

func verify(t testing.TB, values []uint64) {
	t.Helper()
	s := mustNew(t, values)
	if got := s.Len(); got != len(values) {
		t.Fatalf("Len() = %d, want %d", got, len(values))
	}
	for i, want := range values {
		if got := s.At(i); got != want {
			t.Fatalf("At(%d) = %d, want %d (n=%d)", i, got, want, len(values))
		}
	}
	for _, x := range probes(values) {
		if got, want := s.LowerBound(x), refLowerBound(values, x); got != want {
			t.Fatalf("LowerBound(%d) = %d, want %d (n=%d)", x, got, want, len(values))
		}
	}
}

func TestEmpty(t *testing.T) {
	for name, in := range map[string][]uint64{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			s := mustNew(t, in)
			if s.Len() != 0 {
				t.Errorf("Len() = %d, want 0", s.Len())
			}
			for _, x := range []uint64{0, 1, math.MaxUint64} {
				if got := s.LowerBound(x); got != 0 {
					t.Errorf("LowerBound(%d) = %d, want 0", x, got)
				}
			}
			mustPanic(t, func() { s.At(0) })
		})
	}
}

func TestZeroValueSequence(t *testing.T) {
	var s eliasfano.Sequence
	if s.Len() != 0 || s.LowerBound(5) != 0 {
		t.Errorf("zero Sequence: Len=%d LowerBound(5)=%d, want 0, 0", s.Len(), s.LowerBound(5))
	}
	mustPanic(t, func() { s.At(0) })
}

func TestSingleElement(t *testing.T) {
	for _, v := range []uint64{0, 1, 2, 63, 64, 65, 1 << 32, 1<<63 - 1, 1 << 63, math.MaxUint64} {
		verify(t, []uint64{v})
	}
}

func TestZeroValues(t *testing.T) {
	verify(t, []uint64{0})
	verify(t, []uint64{0, 0, 0, 0, 0})
	verify(t, []uint64{0, 0, 1})
	verify(t, []uint64{0, 1, 2, 3})
	verify(t, make([]uint64, 1000))
}

func TestDuplicates(t *testing.T) {
	verify(t, []uint64{7, 7, 7, 7})
	verify(t, []uint64{1, 1, 2, 2, 3, 3})
	verify(t, []uint64{0, 5, 5, 5, 100, 100, math.MaxUint64, math.MaxUint64})

	s := mustNew(t, []uint64{1, 3, 3, 3, 9})
	// First of equal values.
	if got := s.LowerBound(3); got != 1 {
		t.Errorf("LowerBound(3) = %d, want 1", got)
	}
}

func TestConsecutiveIntegers(t *testing.T) {
	for _, base := range []uint64{0, 1, 1000, 1 << 40, math.MaxUint64 - 300} {
		values := make([]uint64, 300)
		for i := range values {
			values[i] = base + uint64(i)
		}
		verify(t, values)
	}
}

func TestSparseWithLargeGaps(t *testing.T) {
	verify(t, []uint64{0, 1 << 20, 1 << 40, 1 << 60, 1<<63 + 5})
	verify(t, []uint64{5, 1_000_000_000_000})
	verify(t, []uint64{0, 1, 2, 1 << 50, 1<<50 + 1, 1 << 62})

	values := make([]uint64, 200)
	for i := range values {
		values[i] = uint64(i) * (1 << 53)
	}
	verify(t, values)
}

func TestLargeValues(t *testing.T) {
	verify(t, []uint64{1 << 63})
	verify(t, []uint64{1<<63 - 1, 1 << 63, 1<<63 + 1})
	verify(t, []uint64{math.MaxUint64 - 2, math.MaxUint64 - 1, math.MaxUint64})
	verify(t, []uint64{0, math.MaxUint64})
	verify(t, []uint64{1 << 32, 1<<32 + 1, 1<<33 - 1, 1 << 33})
}

func TestMaxUint64(t *testing.T) {
	verify(t, []uint64{math.MaxUint64})
	verify(t, []uint64{math.MaxUint64, math.MaxUint64, math.MaxUint64})

	s := mustNew(t, []uint64{10, 20, math.MaxUint64})
	if got := s.At(2); got != math.MaxUint64 {
		t.Errorf("At(2) = %d, want MaxUint64", got)
	}
	if got := s.LowerBound(math.MaxUint64); got != 2 {
		t.Errorf("LowerBound(MaxUint64) = %d, want 2", got)
	}
	if got := s.LowerBound(21); got != 2 {
		t.Errorf("LowerBound(21) = %d, want 2", got)
	}

	// No value >= MaxUint64, so the result is Len().
	s = mustNew(t, []uint64{1, 2, math.MaxUint64 - 1})
	if got := s.LowerBound(math.MaxUint64); got != 3 {
		t.Errorf("LowerBound(MaxUint64) = %d, want 3", got)
	}
}

func TestUnsortedInput(t *testing.T) {
	cases := map[string][]uint64{
		"two descending":   {2, 1},
		"last smaller":     {1, 2, 3, 4, 3},
		"first bigger":     {9, 1, 2, 3},
		"middle dip":       {1, 5, 4, 6},
		"max then smaller": {math.MaxUint64, 0},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			orig := slices.Clone(in)
			s, err := eliasfano.New(in)
			if err == nil || s != nil {
				t.Fatalf("New(%v) = (%v, %v), want nil and an error", in, s, err)
			}
			if !errors.Is(err, eliasfano.ErrNotSorted) {
				t.Errorf("error %q does not wrap ErrNotSorted", err)
			}
			if !slices.Equal(in, orig) {
				t.Errorf("input was modified: %v, want %v", in, orig)
			}
		})
	}
}

func TestNewDoesNotRetainInput(t *testing.T) {
	in := []uint64{1, 2, 3, 100}
	s := mustNew(t, in)
	in[0], in[3] = 999, 999
	if got := s.At(0); got != 1 {
		t.Errorf("At(0) = %d after mutating input, want 1", got)
	}
	if got := s.At(3); got != 100 {
		t.Errorf("At(3) = %d after mutating input, want 100", got)
	}
}

func TestAtInvalidIndex(t *testing.T) {
	s := mustNew(t, []uint64{1, 2, 3})
	for _, i := range []int{-1, 3, 4, math.MinInt, math.MaxInt} {
		t.Run("", func(t *testing.T) {
			msg := mustPanic(t, func() { s.At(i) })
			if !strings.Contains(msg, "out of range") {
				t.Errorf("At(%d) panic = %q, want it to mention \"out of range\"", i, msg)
			}
		})
	}
	if s.At(0) != 1 || s.At(2) != 3 {
		t.Error("valid edge indexes returned wrong values")
	}
}

// mustPanic fails unless f panics; it returns the panic message.
func mustPanic(t *testing.T, f func()) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected a panic, got none")
			return
		}
		msg, _ = r.(string)
	}()
	f()
	return ""
}

func TestLowerBound(t *testing.T) {
	values := []uint64{10, 20, 20, 30, 1000, 1 << 40}
	s := mustNew(t, values)

	cases := []struct {
		name string
		x    uint64
		want int
	}{
		{"zero", 0, 0},
		{"below min", 9, 0},
		{"at min", 10, 0},
		{"between values", 11, 1},
		{"just below existing", 19, 1},
		{"at duplicated value", 20, 1},
		{"after duplicates", 21, 3},
		{"at middle value", 30, 3},
		{"between far values", 31, 4},
		{"at large value", 1 << 40, 5},
		{"just above max", 1<<40 + 1, 6},
		{"MaxUint64", math.MaxUint64, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := s.LowerBound(tc.x); got != tc.want {
				t.Errorf("LowerBound(%d) = %d, want %d", tc.x, got, tc.want)
			}
		})
	}
}

func TestLowerBoundSingleElement(t *testing.T) {
	s := mustNew(t, []uint64{42})
	for x, want := range map[uint64]int{0: 0, 41: 0, 42: 0, 43: 1, math.MaxUint64: 1} {
		if got := s.LowerBound(x); got != want {
			t.Errorf("LowerBound(%d) = %d, want %d", x, got, want)
		}
	}
}

// TestBitBoundaries covers bit positions around word boundaries in both the
// high and low bit vectors.
func TestBitBoundaries(t *testing.T) {
	sizes := []int{1, 2, 31, 32, 33, 62, 63, 64, 65, 66, 126, 127, 128, 129, 130, 191, 192, 193}
	strides := []uint64{0, 1, 2, 3, 5, 7, 31, 33, 64, 65, 127, 129, 1000, 1 << 20, 1 << 40}

	for _, n := range sizes {
		for _, stride := range strides {
			for _, base := range []uint64{0, 1, 63, 64, 65, 127, 128} {
				values := make([]uint64, n)
				for i := range values {
					values[i] = base + uint64(i)*stride
				}
				verify(t, values)
			}
		}
	}
}

// TestSingleValueAtBoundaryPositions puts the last high bit around word
// boundaries.
func TestSingleValueAtBoundaryPositions(t *testing.T) {
	for _, v := range []uint64{62, 63, 64, 65, 66, 126, 127, 128, 129} {
		verify(t, []uint64{0, v})
		verify(t, []uint64{v})
		verify(t, []uint64{v, v})
	}
}

func TestSizeInBytes(t *testing.T) {
	if got := mustNew(t, nil).SizeInBytes(); got != 0 {
		t.Errorf("empty SizeInBytes() = %d, want 0", got)
	}

	// Must be well below the raw []uint64 size.
	r := rand.New(rand.NewPCG(1, 2))
	values := make([]uint64, 10_000)
	var cur uint64
	for i := range values {
		cur += uint64(r.IntN(200))
		values[i] = cur
	}
	s := mustNew(t, values)
	if got, raw := s.SizeInBytes(), 8*len(values); got >= raw/3 {
		t.Errorf("SizeInBytes() = %d, want well below raw size %d", got, raw)
	}
}

func randomSorted(r *rand.Rand, n, kind int) []uint64 {
	values := make([]uint64, n)
	switch kind {
	case 0: // tiny universe: lots of duplicates
		for i := range values {
			values[i] = uint64(r.IntN(8))
		}
	case 1: // dense with some duplicates
		for i := range values {
			values[i] = uint64(r.IntN(n + 1))
		}
	case 2: // sparse
		for i := range values {
			values[i] = r.Uint64N(1 << 40)
		}
	case 3: // full 64-bit range
		for i := range values {
			values[i] = r.Uint64()
		}
	case 4: // cumulative small gaps, including zero gaps
		var cur uint64
		for i := range values {
			cur += uint64(r.IntN(4))
			values[i] = cur
		}
	default: // clustered: a few big jumps between dense runs
		var cur uint64
		for i := range values {
			if r.IntN(20) == 0 {
				cur += r.Uint64N(1 << 50)
			} else {
				cur += uint64(r.IntN(3))
			}
			values[i] = cur
		}
	}
	slices.Sort(values)
	return values
}

func TestRandomProperties(t *testing.T) {
	r := rand.New(rand.NewPCG(0xE1145, 0xFA40))
	const kinds = 6
	iterations := 300
	if testing.Short() {
		iterations = 50
	}
	for it := 0; it < iterations; it++ {
		n := r.IntN(1500)
		kind := it % kinds
		values := randomSorted(r, n, kind)

		s := mustNew(t, values)
		if s.Len() != n {
			t.Fatalf("iter %d kind %d: Len() = %d, want %d", it, kind, s.Len(), n)
		}
		for i, want := range values {
			if got := s.At(i); got != want {
				t.Fatalf("iter %d kind %d n=%d: At(%d) = %d, want %d", it, kind, n, i, got, want)
			}
		}

		queries := probes(values)
		for range 200 {
			queries = append(queries, r.Uint64(), r.Uint64N(uint64(n)+2))
		}
		for _, x := range queries {
			got, want := s.LowerBound(x), refLowerBound(values, x)
			if got != want {
				t.Fatalf("iter %d kind %d n=%d: LowerBound(%d) = %d, want %d", it, kind, n, x, got, want)
			}
		}
	}
}

func FuzzSequence(f *testing.F) {
	f.Add(uint8(0), uint64(0), []byte{})
	f.Add(uint8(0), uint64(5), []byte{1, 0, 0, 0, 0, 0, 0, 0})
	f.Add(uint8(60), uint64(math.MaxUint64), []byte("0123456789abcdef0123456789abcdef"))
	f.Add(uint8(63), uint64(1), bytes.Repeat([]byte{0xff}, 64))
	f.Add(uint8(40), uint64(1<<40), bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 20))

	f.Fuzz(func(t *testing.T, shift uint8, probe uint64, data []byte) {
		const maxValues = 128
		if len(data) > maxValues*8 {
			data = data[:maxValues*8]
		}
		values := make([]uint64, 0, len(data)/8)
		for ; len(data) >= 8; data = data[8:] {
			var v uint64
			for j := 7; j >= 0; j-- {
				v = v<<8 | uint64(data[j])
			}
			// The shift varies magnitude: dense, duplicate-heavy or sparse.
			values = append(values, v>>(shift%64))
		}

		// New must fail exactly when the input is unsorted.
		s, err := eliasfano.New(values)
		if sorted := slices.IsSorted(values); sorted != (err == nil) {
			t.Fatalf("New(%v): sorted=%v but err=%v", values, sorted, err)
		}
		if err != nil {
			if !errors.Is(err, eliasfano.ErrNotSorted) {
				t.Fatalf("unexpected error type: %v", err)
			}
			slices.Sort(values)
			s, err = eliasfano.New(values)
			if err != nil {
				t.Fatalf("New(sorted %v): %v", values, err)
			}
		}

		if s.Len() != len(values) {
			t.Fatalf("Len() = %d, want %d", s.Len(), len(values))
		}
		for i, want := range values {
			if got := s.At(i); got != want {
				t.Fatalf("At(%d) = %d, want %d; values=%v", i, got, want, values)
			}
		}
		queries := append(probes(values), probe, probe>>(shift%64))
		for _, x := range queries {
			if got, want := s.LowerBound(x), refLowerBound(values, x); got != want {
				t.Fatalf("LowerBound(%d) = %d, want %d; values=%v", x, got, want, values)
			}
		}
	})
}
