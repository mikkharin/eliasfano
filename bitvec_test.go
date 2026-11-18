package eliasfano

import (
	"math"
	"math/bits"
	"math/rand/v2"
	"slices"
	"testing"
)

var boundaryPositions = []uint64{0, 1, 62, 63, 64, 65, 126, 127, 128, 129, 191, 192, 193}

func TestSetGetBit(t *testing.T) {
	for _, p := range boundaryPositions {
		v := make([]uint64, 4)
		setBit(v, p)
		for q := uint64(0); q < 256; q++ {
			if got, want := getBit(v, q), q == p; got != want {
				t.Fatalf("after setBit(%d): getBit(%d) = %v, want %v", p, q, got, want)
			}
		}
	}
}

func TestWordsFor(t *testing.T) {
	for nbits, want := range map[uint64]int{0: 0, 1: 1, 63: 1, 64: 1, 65: 2, 127: 2, 128: 2, 129: 3} {
		if got := wordsFor(nbits); got != want {
			t.Errorf("wordsFor(%d) = %d, want %d", nbits, got, want)
		}
	}
}

func TestSetGetBits(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	for width := uint(0); width <= 64; width++ {
		for _, pos := range boundaryPositions {
			v := make([]uint64, 5)
			val := r.Uint64()
			if width < 64 {
				val &= 1<<width - 1
			}
			setBits(v, pos, width, val)
			if got := getBits(v, pos, width); got != val {
				t.Fatalf("width=%d pos=%d: getBits = %#x, want %#x", width, pos, got, val)
			}
			// No bits outside the field may be set.
			total := 0
			for _, w := range v {
				total += bits.OnesCount64(w)
			}
			if want := bits.OnesCount64(val); total != want {
				t.Fatalf("width=%d pos=%d: %d bits set overall, want %d", width, pos, total, want)
			}
		}
	}
}

func TestSetBitsMasksValue(t *testing.T) {
	v := make([]uint64, 2)
	setBits(v, 60, 8, 0xFFFF) // only the low 8 bits may be written
	if got := getBits(v, 60, 8); got != 0xFF {
		t.Errorf("getBits = %#x, want 0xff", got)
	}
	if got := getBits(v, 68, 8); got != 0 {
		t.Errorf("bits beyond the field were written: %#x", got)
	}
}

func TestGetBitsAdjacentFields(t *testing.T) {
	for width := uint(1); width <= 64; width++ {
		const count = 150
		v := make([]uint64, wordsFor(count*uint64(width)))
		r := rand.New(rand.NewPCG(uint64(width), 99))
		want := make([]uint64, count)
		for i := range want {
			want[i] = r.Uint64()
			if width < 64 {
				want[i] &= 1<<width - 1
			}
			setBits(v, uint64(i)*uint64(width), width, want[i])
		}
		for i, w := range want {
			if got := getBits(v, uint64(i)*uint64(width), width); got != w {
				t.Fatalf("width=%d field %d: got %#x, want %#x", width, i, got, w)
			}
		}
	}
}

func TestSelectInWord(t *testing.T) {
	words := []uint64{1, 1 << 63, math.MaxUint64, 0x8000000000000001, 0xAAAAAAAAAAAAAAAA, 0x0123456789ABCDEF}
	for _, w := range words {
		k := 0
		for pos := 0; pos < 64; pos++ {
			if w>>pos&1 == 1 {
				if got := selectInWord(w, k); got != pos {
					t.Errorf("selectInWord(%#x, %d) = %d, want %d", w, k, got, pos)
				}
				k++
			}
		}
	}
}

func onesPositions(v []uint64) []uint64 {
	var out []uint64
	for wi, w := range v {
		for ; w != 0; w &= w - 1 {
			out = append(out, uint64(wi)*64+uint64(bits.TrailingZeros64(w)))
		}
	}
	return out
}

func checkSelect(t *testing.T, words []uint64) {
	t.Helper()
	want := onesPositions(words)
	bv := newBitVector(slices.Clone(words))
	for k, pos := range want {
		if got := bv.select1(uint64(k)); got != pos {
			t.Fatalf("select1(%d) = %d, want %d", k, got, pos)
		}
	}
}

func TestSelect1SingleBit(t *testing.T) {
	for _, p := range boundaryPositions {
		words := make([]uint64, 4)
		setBit(words, p)
		checkSelect(t, words)
	}
}

func TestSelect1BoundaryPairs(t *testing.T) {
	for _, a := range boundaryPositions {
		for _, b := range boundaryPositions {
			if a >= b {
				continue
			}
			words := make([]uint64, 4)
			setBit(words, a)
			setBit(words, b)
			checkSelect(t, words)
		}
	}
}

func TestSelect1Dense(t *testing.T) {
	words := make([]uint64, 100)
	for i := range words {
		words[i] = math.MaxUint64
	}
	checkSelect(t, words)
}

func TestSelect1Random(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for _, density := range []int{1, 2, 8, 64, 1000} {
		words := make([]uint64, 300)
		for i := 0; i < len(words)*64/density; i++ {
			setBit(words, r.Uint64N(uint64(len(words))*64))
		}
		checkSelect(t, words)
	}
}

func TestSelect1SparseAcrossManyWords(t *testing.T) {
	// Set bits far apart within one sample interval.
	words := make([]uint64, 2000)
	for i := 0; i < 3*selectSampleRate; i++ {
		setBit(words, uint64(i)*37)
	}
	checkSelect(t, words)
}

func TestNewBitVectorSamples(t *testing.T) {
	words := make([]uint64, 20)
	for i := 0; i < 2*selectSampleRate+1; i++ {
		setBit(words, uint64(i)*2%(20*64))
	}
	bv := newBitVector(words)
	total := 0
	for _, w := range words {
		total += bits.OnesCount64(w)
	}
	if want := (total + selectSampleRate - 1) / selectSampleRate; len(bv.samples) != want {
		t.Errorf("len(samples) = %d, want %d for %d set bits", len(bv.samples), want, total)
	}
}
