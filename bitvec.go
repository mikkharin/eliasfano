package eliasfano

import "math/bits"

const (
	wordBits = 64

	// Set bits between select samples: larger means less memory, longer scans.
	selectSampleRate = 512
)

func wordsFor(nbits uint64) int {
	return int((nbits + wordBits - 1) / wordBits)
}

func setBit(v []uint64, pos uint64) {
	v[pos/wordBits] |= 1 << (pos % wordBits)
}

func getBit(v []uint64, pos uint64) bool {
	return v[pos/wordBits]>>(pos%wordBits)&1 == 1
}

// getBits reads a width-bit field (0..64) at pos; it may straddle two words.
func getBits(v []uint64, pos uint64, width uint) uint64 {
	if width == 0 {
		return 0
	}
	w, off := pos/wordBits, uint(pos%wordBits)
	x := v[w] >> off
	if off+width > wordBits {
		x |= v[w+1] << (wordBits - off)
	}
	if width < wordBits {
		x &= 1<<width - 1
	}
	return x
}

// setBits ORs the low width bits of val in at pos; the target must be zero.
func setBits(v []uint64, pos uint64, width uint, val uint64) {
	if width == 0 {
		return
	}
	if width < wordBits {
		val &= 1<<width - 1
	}
	w, off := pos/wordBits, uint(pos%wordBits)
	v[w] |= val << off
	if off+width > wordBits {
		v[w+1] |= val >> (wordBits - off)
	}
}

// selectInWord returns the position of the k-th (0-based) set bit of w.
func selectInWord(w uint64, k int) int {
	for ; k > 0; k-- {
		w &= w - 1
	}
	return bits.TrailingZeros64(w)
}

type bitVector struct {
	words []uint64
	// samples[j] is the position of the (j*selectSampleRate)-th set bit.
	samples []uint64
}

// newBitVector takes ownership of words.
func newBitVector(words []uint64) bitVector {
	var samples []uint64
	count := 0
	for wi, w := range words {
		for ; w != 0; w &= w - 1 {
			if count%selectSampleRate == 0 {
				pos := uint64(wi)*wordBits + uint64(bits.TrailingZeros64(w))
				samples = append(samples, pos)
			}
			count++
		}
	}
	return bitVector{words: words, samples: samples}
}

// select1 returns the position of the k-th (0-based) set bit; k must exist.
func (b bitVector) select1(k uint64) uint64 {
	pos := b.samples[k/selectSampleRate]
	rem := int(k % selectSampleRate)

	wi := pos / wordBits
	// Skip bits below the sample.
	w := b.words[wi] >> (pos % wordBits) << (pos % wordBits)
	for {
		c := bits.OnesCount64(w)
		if rem < c {
			return wi*wordBits + uint64(selectInWord(w, rem))
		}
		rem -= c
		wi++
		w = b.words[wi]
	}
}
