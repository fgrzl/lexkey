package lexkey

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// Special bytes used in lexicographic encoding
const (
	Seperator = 0x00 // Separates parts within a LexKey
	EndMarker = 0xFF // Marks the end of a range for lexicographic sorting
)

var (
	Empty = LexKey{}
	Last  = Encode(EndMarker)
)

// LexKey represents an encoded key as a byte slice, optimized for lexicographic sorting.
// An empty LexKey (length 0) is valid and distinct from nil.
type LexKey []byte

// NewLexKey constructs a LexKey from a list of parts, ensuring lexicographic sorting.
// Returns an error if parts is empty or contains unsupported types.
// The resulting key is a concatenation of encoded parts separated by Seperator bytes.
func NewLexKey(parts ...any) (LexKey, error) {
	// Default behavior: canonicalize numeric widths before encoding (BREAKING CHANGE).
	return NewLexKeyCanonicalWidth(parts...)
}

// Encode constructs a LexKey from pre-validated parts, panicking if encoding fails.
// Use this when inputs are guaranteed to be valid (e.g., no unsupported types).
// For fallible construction, use NewLexKey instead.
func Encode(parts ...any) LexKey {
	key, err := NewLexKey(parts...)
	if err != nil {
		panic(fmt.Sprintf("failed to encode key with pre-validated parts: %v", err))
	}
	return key
}

// canonicalizeNumericWidth converts narrower numeric types to 64-bit widths so
// cross-width numeric values sort identically (e.g., uint32(1) vs uint64(1)).
// It does NOT merge signed and unsigned domains, nor ints vs floats.
// Mapping:
// - int, int8, int16, int32 -> int64
// - uint8, uint16, uint32 -> uint64
// - float32 -> float64
// Other types unchanged.
func canonicalizeNumericWidth(v any) any {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case uint8:
		return uint64(x)
	case uint16:
		return uint64(x)
	case uint32:
		return uint64(x)
	case float32:
		return float64(x)
	default:
		return v
	}
}

// NewLexKeyCanonicalWidth constructs a LexKey after normalizing numeric widths
// so cross-width numeric values sort logically. Backwards-compatible: bytes differ
// from NewLexKey because widths are canonicalized; use only if you control both sides.
func NewLexKeyCanonicalWidth(parts ...any) (LexKey, error) {
	if len(parts) == 0 {
		return LexKey([]byte{}), errors.New("cannot create LexKey: no parts provided")
	}
	canon := make([]any, len(parts))
	for i, p := range parts {
		canon[i] = canonicalizeNumericWidth(p)
	}
	size := estimateSize(canon)
	result := make([]byte, size)
	pos := 0
	for i, part := range canon {
		n, err := encodeInto(result[pos:], part)
		if err != nil {
			return LexKey([]byte{}), fmt.Errorf("cannot encode part %d (%T): %w", i, part, err)
		}
		pos += n
		if i < len(canon)-1 {
			result[pos] = Seperator
			pos++
		}
	}
	return result[:pos], nil
}

// EncodeCanonicalWidth panics on error; see NewLexKeyCanonicalWidth.
func EncodeCanonicalWidth(parts ...any) LexKey {
	key, err := NewLexKeyCanonicalWidth(parts...)
	if err != nil {
		panic(fmt.Sprintf("failed to encode canonical-width key: %v", err))
	}
	return key
}

// EncodeSizeCanonicalWidth returns the size after width canonicalization.
func EncodeSizeCanonicalWidth(parts ...any) int {
	if len(parts) == 0 {
		return 0
	}
	canon := make([]any, len(parts))
	for i, p := range parts {
		canon[i] = canonicalizeNumericWidth(p)
	}
	return estimateSize(canon)
}

// EncodeIntoCanonicalWidth writes the canonical-width encoding into dst.
func EncodeIntoCanonicalWidth(dst []byte, parts ...any) (int, error) {
	if len(parts) == 0 {
		return 0, nil
	}
	canon := make([]any, len(parts))
	for i, p := range parts {
		canon[i] = canonicalizeNumericWidth(p)
	}
	need := estimateSize(canon)
	if len(dst) < need {
		return 0, fmt.Errorf("EncodeInto: dst too small: need %d bytes, have %d", need, len(dst))
	}
	pos := 0
	for i, part := range canon {
		n, err := encodeInto(dst[pos:], part)
		if err != nil {
			return 0, fmt.Errorf("cannot encode part %d (%T): %w", i, part, err)
		}
		pos += n
		if i < len(canon)-1 {
			dst[pos] = Seperator
			pos++
		}
	}
	return pos, nil
}

// EncodeFirst returns the first lexicographically sortable key in a range.
// Adds a Seperator byte to the prefix to ensure it sorts before any extension.
//
// Note: this calls Encode and will panic on encoding errors (i.e., when given unsupported types).
func EncodeFirst(parts ...any) LexKey {
	prefix := Encode(parts...)
	return append(prefix, Seperator)
}

// EncodeLast returns the last lexicographically sortable key in a range.
// Adds an EndMarker byte to the prefix to ensure it sorts after any extension.
//
// Note: this calls Encode and will panic on encoding errors (i.e., when given unsupported types).
func EncodeLast(parts ...any) LexKey {
	prefix := Encode(parts...)
	return append(prefix, EndMarker)
}

// IsEmpty checks if the LexKey is empty (length 0). A nil LexKey is considered empty.
func (e LexKey) IsEmpty() bool {
	return len(e) == 0
}

// ToHexString converts the LexKey to a hexadecimal string.
// Returns an empty string for an empty or nil LexKey.
func (e LexKey) ToHexString() string {
	if len(e) == 0 {
		return ""
	}
	return hex.EncodeToString(e)
}

// EncodeSize returns the exact number of bytes required to encode the given parts,
// including separators. Use this to pre-allocate a destination buffer for EncodeInto.
func EncodeSize(parts ...any) int {
	return EncodeSizeCanonicalWidth(parts...)
}

// EncodeInto writes the encoding of parts into dst and returns the number of bytes written.
// The dst slice must have length >= EncodeSize(parts...). No allocations are performed.
func EncodeInto(dst []byte, parts ...any) (int, error) {
	return EncodeIntoCanonicalWidth(dst, parts...)
}

// FromHexString decodes a hexadecimal string back into a LexKey.
// Sets to an empty slice (not nil) for an empty input string.
// Returns an error if the hex string is invalid.
func (e *LexKey) FromHexString(hexStr string) error {
	if len(hexStr) == 0 {
		*e = []byte{} // Empty slice instead of nil
		return nil
	}
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		if errors.Is(err, hex.ErrLength) {
			return fmt.Errorf("cannot decode hex string: odd length")
		}
		return fmt.Errorf("cannot decode hex string: %w", err)
	}
	*e = bytes
	return nil
}

// MarshalJSON encodes LexKey as a hex string for JSON serialization.
func (e LexKey) MarshalJSON() ([]byte, error) {
	// Use MarshalText to get hex bytes, then wrap with quotes without extra escaping
	text, err := e.MarshalText()
	if err != nil {
		return nil, err
	}
	// Build JSON string: "<hex>"
	out := make([]byte, len(text)+2)
	out[0] = '"'
	copy(out[1:], text)
	out[len(out)-1] = '"'
	return out, nil
}

// UnmarshalJSON decodes a hex string from JSON into a LexKey.
// Handles JSON null by setting to an empty slice.
func (e *LexKey) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		*e = []byte{}
		return nil
	}
	// Expect a JSON string: "..."
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		inner := data[1 : len(data)-1]
		return e.UnmarshalText(inner)
	}
	return fmt.Errorf("cannot unmarshal JSON into LexKey: invalid format")
}

// MarshalText implements encoding.TextMarshaler, returning a hex encoding of the key.
func (e LexKey) MarshalText() ([]byte, error) {
	if len(e) == 0 {
		return []byte{}, nil
	}
	dst := make([]byte, hex.EncodedLen(len(e)))
	hex.Encode(dst, e)
	return dst, nil
}

// UnmarshalText implements encoding.TextUnmarshaler, decoding a hex input into the key.
func (e *LexKey) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*e = []byte{}
		return nil
	}
	// hex.Decode expects even-length; it returns an error on malformed input
	dst := make([]byte, hex.DecodedLen(len(text)))
	n, err := hex.Decode(dst, text)
	if err != nil {
		return fmt.Errorf("cannot decode hex string: %w", err)
	}
	*e = dst[:n]
	return nil
}

// encodeToBytes converts a value to a lexicographically sortable byte representation.
// Returns an error if the type is unsupported.
func encodeToBytes(v any) ([]byte, error) {
	// Backwards-compatible wrapper: allocate a sufficiently large temp buffer and use encodeInto.
	// Most types encode into at most 16 bytes (UUID) plus separators.
	buf := make([]byte, 32)
	// Apply canonical width so encodeToBytes matches default behavior
	n, err := encodeInto(buf, canonicalizeNumericWidth(v))
	if err != nil {
		return nil, err
	}
	// return a copy of the used portion
	out := make([]byte, n)
	copy(out, buf[:n])
	return out, nil
}

// encodeInto writes the lexicographic encoding of v into dst and returns the number of bytes written.
// dst must be large enough to hold the encoding; caller is responsible for sizing it (estimateSize).
func encodeInto(dst []byte, v any) (int, error) {
	switch v := v.(type) {
	case string:
		n := copy(dst, v)
		return n, nil
	case uuid.UUID:
		n := copy(dst, v[:])
		return n, nil
	case LexKey:
		n := copy(dst, v)
		return n, nil
	case []byte:
		n := copy(dst, v)
		return n, nil
	case int:
		// encode as int64
		binary.BigEndian.PutUint64(dst, uint64(int64(v))^0x8000000000000000)
		return 8, nil
	case int64:
		binary.BigEndian.PutUint64(dst, uint64(v)^0x8000000000000000)
		return 8, nil
	case int32:
		binary.BigEndian.PutUint32(dst, uint32(v)^0x80000000)
		return 4, nil
	case int16:
		binary.BigEndian.PutUint16(dst, uint16(v)^0x8000)
		return 2, nil
	case uint64:
		binary.BigEndian.PutUint64(dst, v)
		return 8, nil
	case uint32:
		binary.BigEndian.PutUint32(dst, v)
		return 4, nil
	case uint16:
		binary.BigEndian.PutUint16(dst, v)
		return 2, nil
	case uint8:
		dst[0] = v
		return 1, nil
	case float64:
		if math.IsNaN(v) {
			binary.BigEndian.PutUint64(dst, 0x7FF8000000000001)
			return 8, nil
		}
		bits := math.Float64bits(v)
		if v < 0 {
			bits = ^bits
		} else {
			bits ^= 1 << 63
		}
		binary.BigEndian.PutUint64(dst, bits)
		return 8, nil
	case float32:
		if math.IsNaN(float64(v)) {
			binary.BigEndian.PutUint32(dst, 0x7FC00001)
			return 4, nil
		}
		bits := math.Float32bits(v)
		if v < 0 {
			bits = ^bits
		} else {
			bits ^= 1 << 31
		}
		binary.BigEndian.PutUint32(dst, bits)
		return 4, nil
	case bool:
		if v {
			dst[0] = 1
		} else {
			dst[0] = 0
		}
		return 1, nil
	case time.Time:
		// encode as int64 of UnixNano with sign flip
		t := v.UTC().UnixNano()
		binary.BigEndian.PutUint64(dst, uint64(t)^0x8000000000000000)
		return 8, nil
	case time.Duration:
		binary.BigEndian.PutUint64(dst, uint64(int64(v))^0x8000000000000000)
		return 8, nil
	case nil:
		dst[0] = Seperator
		return 1, nil
	case struct{}:
		dst[0] = EndMarker
		return 1, nil
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

// estimateSize predicts the byte size of a LexKey based on its parts, including Seperators.
func estimateSize(parts []any) int {
	size := 0
	for i, part := range parts {
		switch v := part.(type) {
		case string:
			size += len(v)
		case uuid.UUID:
			size += 16
		case LexKey:
			size += len(v)
		case []byte:
			size += len(v)
		case int, int64, uint64, time.Time, time.Duration:
			size += 8
		case int32, uint32:
			size += 4
		case int16, uint16:
			size += 2
		case uint8:
			size += 1
		case float64:
			size += 8
		case float32:
			size += 4
		case bool:
			size += 1
		case nil, struct{}:
			size += 1
		default:
			// Unsupported types will error later; assume minimal size
			size += 1
		}
		if i < len(parts)-1 {
			size += 1 // Seperator
		}
	}
	return size
}

// encodeFloat64 encodes a float64 into 8 bytes, ensuring lexicographic ordering.
// Flips the sign bit for positive numbers and all bits for negative numbers.
// NaN is encoded as a canonical value (0x7FF8000000000001).
func encodeFloat64(v float64) []byte {
	buf := make([]byte, 8)
	if math.IsNaN(v) {
		binary.BigEndian.PutUint64(buf, 0x7FF8000000000001) // Canonical NaN
		return buf
	}
	bits := math.Float64bits(v)
	if v < 0 {
		bits = ^bits // Flip all bits for negative numbers
	} else {
		bits ^= 1 << 63 // Flip sign bit for positive numbers
	}
	binary.BigEndian.PutUint64(buf, bits)
	return buf
}

// encodeFloat32 encodes a float32 into 4 bytes, ensuring lexicographic ordering.
// Flips the sign bit for positive numbers and all bits for negative numbers.
// NaN is encoded as a canonical value (0x7FC00001).
func encodeFloat32(v float32) []byte {
	buf := make([]byte, 4)
	if math.IsNaN(float64(v)) {
		binary.BigEndian.PutUint32(buf, 0x7FC00001) // Canonical NaN
		return buf
	}
	bits := math.Float32bits(v)
	if v < 0 {
		bits = ^bits // Flip all bits for negative numbers
	} else {
		bits ^= 1 << 31 // Flip sign bit for positive numbers
	}
	binary.BigEndian.PutUint32(buf, bits)
	return buf
}

// Compare returns -1, 0, 1 for a < b, a == b, a > b respectively without allocations.
func Compare(a, b LexKey) int {
	return bytes.Compare(a, b)
}
