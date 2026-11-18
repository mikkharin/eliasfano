# eliasfano

[![Go Reference](https://pkg.go.dev/badge/github.com/mikkharin/eliasfano.svg)](https://pkg.go.dev/github.com/mikkharin/eliasfano)

A small, dependency-free Go library for [Elias–Fano](https://en.wikipedia.org/wiki/Elias%E2%80%93Fano_coding)
encoding of monotonic (non-decreasing) `uint64` sequences. It stores `n`
values with maximum `m` in about `n·(2 + log₂(m/n))` bits and supports random
access and successor queries without decompression.

> Status: v0.1.0, an exercise project. The API may change before v1.

## Install

```sh
go get github.com/mikkharin/eliasfano
```

Requires Go 1.22+. Standard library only.

## Usage

```go
s, err := eliasfano.New([]uint64{3, 4, 7, 13, 14, 15, 21, 25, 36, 38, 54, 62})
if err != nil {
	log.Fatal(err) // input was not sorted
}

s.Len()          // 12
s.At(3)          // 13
s.LowerBound(22) // 7: index of the first value >= 22 (25)
s.LowerBound(99) // 12 == s.Len(): every value is smaller
```

## API

| Function / method | Description |
| --- | --- |
| `New(values []uint64) (*Sequence, error)` | Encodes a sorted slice; returns an error wrapping `ErrNotSorted` otherwise. The input is not retained. |
| `(*Sequence) Len() int` | Number of values. |
| `(*Sequence) At(i int) uint64` | The i-th value, effectively O(1). Panics if `i` is out of range, like a slice index. |
| `(*Sequence) LowerBound(x uint64) int` | Smallest index `i` with `At(i) >= x`, or `Len()`. O(log n). |
| `(*Sequence) SizeInBytes() int` | Approximate size of the encoded data. |

A `Sequence` is immutable and safe for concurrent use.

## Development

```sh
make check   # gofmt, go vet, golangci-lint, tests
make test
make cover
make bench
make fuzz FUZZTIME=1m
```

## License

[MIT](LICENSE)
