# Overview

lexkey turns heterogeneous Go values into **`LexKey` byte strings** that preserve sort order when compared with `bytes.Compare` or stored in ordered KV systems.

## Supported types

Strings, signed/unsigned integers (canonical widths), floats, booleans, UUIDs, timestamps, durations, and byte slices. Mixed-type segments in one key sort predictably because each segment carries a type tag.

## Design goals

- **Correct ordering** — `int32` and `int64` compare consistently; negative integers sort below positive
- **Compact encoding** — type tags and payloads without JSON overhead
- **Interop** — JSON marshaling as hex for APIs and configs
- **KV integration** — `NewPrimaryKey(partition, row)` for `github.com/fgrzl/kv`

## Range helpers

`EncodeFirst` and `EncodeLast` produce boundary keys for prefix scans (for example all rows under a partition).

## Specification

Binary layout and type byte values are defined in [SPEC.md](../SPEC.md). Treat the spec as authoritative when implementing compatible encoders in other languages.
