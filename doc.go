// Package eliasfano implements Elias–Fano encoding of non-decreasing uint64
// sequences.
//
// n values with maximum m take roughly n*(2 + log2(m/n)) bits, while At
// (random access) and LowerBound (successor search) work without
// decompressing the sequence.
//
// At takes O(1) time on typical data but is linear in the sequence length in
// the worst case; LowerBound makes O(log n) calls to At.
//
// A Sequence is immutable and safe for concurrent use.
package eliasfano
