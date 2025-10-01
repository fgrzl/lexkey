[![ci](https://github.com/fgrzl/lexkey/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/lexkey/actions/workflows/ci.yml)
[![Dependabot Updates](https://github.com/fgrzl/lexkey/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/lexkey/actions/workflows/dependabot/dependabot-updates)

# lexkey

`lexkey` is a lightweight lexicographically sortable key encoding library for Go.
It provides consistent, ordered, and efficient encoding for common data types so they sort correctly in byte-wise order.

## ✨ **Features**

- 🚀 **Lexicographically sortable** encoding for structured keys.
- 🔑 Supports **strings, integers, floats, UUIDs, booleans, timestamps, durations, and byte slices**.
- 🔄 **Consistent ordering** for mixed types like `int32` and `int64`.
- 📦 **Optimized encoding** for space efficiency and speed.
- 📡 **JSON serialization support** for interoperability.

## 📦 **Installation**

```sh
go get github.com/fgrzl/lexkey
```

## 🛠 Usage

### **Create a LexKey**

```go
key := lexkey.Encode("user", 123, true)
fmt.Println("Encoded Key (Hex):", key.ToHexString())
```

### **Sorting Keys**

```go
key1 := lexkey.Encode("apple")
key2 := lexkey.Encode("banana")

fmt.Println(string(key1) < string(key2)) // ✅ True (correct lexicographic order)
```

### **Handling Numbers**

```go
key1 := lexkey.Encode(int64(-100))
key2 := lexkey.Encode(int64(50))

fmt.Println(string(key1) < string(key2)) // ✅ True (correct sorting for signed integers)
```

### **Using UUIDs**

```go
import "github.com/google/uuid"

id := uuid.New()
key := lexkey.Encode("order", id)

fmt.Println("Encoded UUID Key:", key.ToHexString())
```

### **LexKey JSON Serialization**

```go
import "encoding/json"

key := lexkey.Encode("session", 42)
jsonData, _ := json.Marshal(key)
fmt.Println(string(jsonData)) // ✅ Encoded as a hex string
```

## 🔍 Supported Data Types

| Type            | Supported? | Encoding Details                                |
| --------------- | ---------- | ----------------------------------------------- |
| `string`        | ✅ Yes     | Stored as raw UTF-8 bytes                       |
| `int32`         | ✅ Yes     | Canonicalized to `int64` for uniform sorting    |
| `int64`         | ✅ Yes     | Sign-bit flipped for correct ordering           |
| `uint32`        | ✅ Yes     | Big-endian encoded                              |
| `uint64`        | ✅ Yes     | Big-endian encoded                              |
| `float32`       | ✅ Yes     | Canonicalized to `float64` then transformed     |
| `float64`       | ✅ Yes     | IEEE 754 encoded with sign-bit transformation   |
| `bool`          | ✅ Yes     | `true → 0x01`, `false → 0x00`                   |
| `uuid.UUID`     | ✅ Yes     | 16-byte raw representation                      |
| `[]byte`        | ✅ Yes     | Stored as-is                                    |
| `time.Time`     | ✅ Yes     | Encoded as `int64` nanoseconds since Unix epoch |
| `time.Duration` | ✅ Yes     | Encoded as `int64` nanoseconds                  |

## 📌 **Key Functions**

### Encoding Keys

```go
func Encode(parts ...any) LexKey
func NewLexKey(parts ...any) (LexKey, error)
func EncodeInto(dst []byte, parts ...any) (int, error)
func EncodeSize(parts ...any) int
```

Notes:
- As of 2025-10-01, numeric width canonicalization is the default (breaking change):
	- int, int8, int16, int32 → int64
	- uint8, uint16, uint32 → uint64
	- float32 → float64
- For explicit use, the following helpers are provided (equivalent to default behavior):
	- EncodeCanonicalWidth / NewLexKeyCanonicalWidth / EncodeIntoCanonicalWidth / EncodeSizeCanonicalWidth

### Sorting Helpers

```go
func EncodeFirst(parts ...any) LexKey // lower bound: prefix + 0x00 (sorts before any extension of the prefix)
func EncodeLast(parts ...any) LexKey  // upper bound: prefix + 0xFF (sorts after any extension of the prefix)
func Compare(a, b LexKey) int         // -1/0/1 without allocations
```

Prefix scans:

```go
// To scan all keys with a given prefix, use a half-open range [lower, upper):
lower := lexkey.EncodeFirst("tenant", "users") // ... 00
upper := lexkey.EncodeLast("tenant", "users")  // ... ff
// All keys that start with ("tenant", "users", ...) will satisfy: lower <= key && key < upper
```

### Hex Encoding

```go
func (e LexKey) ToHexString() string
func (e *LexKey) FromHexString(hexStr string) error
```

### JSON Serialization

```go
func (e LexKey) MarshalJSON() ([]byte, error)
func (e *LexKey) UnmarshalJSON(data []byte) error
```

## 🏆 Why Use `lexkey`?

✅ **Fast & Efficient** → Uses compact, binary-safe encoding.  
✅ **Correct Ordering** → Works across all supported types.  
✅ **Minimal Dependencies** → Only `uuid` and standard Go packages.

## 🛠 Testing

Run the full test suite:

```sh
go test -cover ./...
```

## 🔄 Breaking change (2025-10-01)

Numeric width canonicalization is now the default. This ensures logical numeric ordering across widths (e.g., uint32(1) and uint64(1) are equal on the wire). If you require the previous native-width bytes, pin an older release or re-encode using legacy rules as described in SPEC.md.

See SPEC.md for the full encoding specification and test vectors.

**Test Coverage:** ✅ **99.6%** 🎯
