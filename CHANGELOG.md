# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/).

## [0.1.0]

### Added

- `eliasfano.New` to encode a sorted `[]uint64` (duplicates allowed).
- `Sequence.Len`, `Sequence.At`, `Sequence.LowerBound`, `Sequence.SizeInBytes`.
- `ErrNotSorted` for unsorted input.
