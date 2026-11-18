package eliasfano

import (
	"errors"
	"fmt"
	"math/bits"
)

// ErrNotSorted is returned by New for unsorted input.
var ErrNotSorted = errors.New("eliasfano: input is not sorted in non-decreasing order")

// Sequence is an immutable Elias–Fano encoded non-decreasing sequence.
// The zero value is empty.
type Sequence struct {
	n   int
	max uint64
	l   uint // number of low bits stored per value
	low []uint64
	// high holds unary-coded high parts: value i sets bit (value>>l)+i.
	high bitVector
}

// New encodes values, which must be sorted in non-decreasing order.
// The input is not retained.
func New(values []uint64) (*Sequence, error) {
	n := len(values)
	for i := 1; i < n; i++ {
		if values[i] < values[i-1] {
			return nil, fmt.Errorf("%w: values[%d]=%d < values[%d]=%d",
				ErrNotSorted, i, values[i], i-1, values[i-1])
		}
	}
	if n == 0 {
		return &Sequence{}, nil
	}

	last := values[n-1]
	// l = floor(log2(last/n)); last instead of last+1 avoids overflow.
	var l uint
	if q := last / uint64(n); q > 0 {
		l = uint(bits.Len64(q)) - 1
	}

	low := make([]uint64, wordsFor(uint64(n)*uint64(l)))
	highWords := make([]uint64, wordsFor(uint64(n)+last>>l+1))
	for i, v := range values {
		setBits(low, uint64(i)*uint64(l), l, v)
		setBit(highWords, v>>l+uint64(i))
	}

	return &Sequence{
		n:    n,
		max:  last,
		l:    l,
		low:  low,
		high: newBitVector(highWords),
	}, nil
}

// Len returns the number of values.
func (s *Sequence) Len() int { return s.n }

// At returns the i-th value. It panics if i is out of range.
func (s *Sequence) At(i int) uint64 {
	if i < 0 || i >= s.n {
		panic(fmt.Sprintf("eliasfano: index %d out of range [0:%d)", i, s.n))
	}
	high := s.high.select1(uint64(i)) - uint64(i)
	low := getBits(s.low, uint64(i)*uint64(s.l), s.l)
	return high<<s.l | low
}

// LowerBound returns the smallest i with At(i) >= x, or Len() if none.
// It runs in O(log n).
func (s *Sequence) LowerBound(x uint64) int {
	if x > s.max {
		return s.n
	}
	lo, hi := 0, s.n
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		if s.At(mid) < x {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// SizeInBytes returns the approximate size of the encoded data.
func (s *Sequence) SizeInBytes() int {
	return 8 * (len(s.low) + len(s.high.words) + len(s.high.samples))
}
