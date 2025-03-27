package lexkey

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
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

// LexKey represents an encoded key as a byte slice, optimized for lexicographic sorting.
// An empty LexKey (length 0) is valid and distinct from nil.
type LexKey []byte

// NewLexKey constructs a LexKey from a list of parts, ensuring lexicographic sorting.
// Returns an error if parts is empty or contains unsupported types.
// The resulting key is a concatenation of encoded parts separated by Seperator bytes.
func NewLexKey(parts ...any) (LexKey, error) {
	if len(parts) == 0 {
		return LexKey([]byte{}), errors.New("cannot create LexKey: no parts provided")
	}

	// Pre-allocate result slice based on estimated size
	size := estimateSize(parts)
	result := make([]byte, 0, size)

	for i, part := range parts {
		encoded, err := encodeToBytes(part)
		if err != nil {
			return LexKey([]byte{}), fmt.Errorf("cannot encode part %d (%T): %w", i, part, err)
		}
		result = append(result, encoded...)
		if i < len(parts)-1 {
			result = append(result, Seperator)
		}
	}
	return result, nil
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

// EncodeFirst returns the first lexicographically sortable key in a range.
// Adds a Seperator byte to the prefix to ensure it sorts before any extension.
func EncodeFirst(parts ...any) LexKey {
	prefix := Encode(parts...)
	return append(prefix, Seperator)
}

// EncodeLast returns the last lexicographically sortable key in a range.
// Adds an EndMarker byte to the prefix to ensure it sorts after any extension.
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
	return json.Marshal(e.ToHexString())
}

// UnmarshalJSON decodes a hex string from JSON into a LexKey.
// Handles JSON null by setting to an empty slice.
func (e *LexKey) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*e = []byte{}
		return nil
	}
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return fmt.Errorf("cannot unmarshal JSON into LexKey: %w", err)
	}
	return e.FromHexString(hexStr)
}

// encodeToBytes converts a value to a lexicographically sortable byte representation.
// Returns an error if the type is unsupported.
func encodeToBytes(v any) ([]byte, error) {
	switch v := v.(type) {
	case string:
		return []byte(v), nil
	case uuid.UUID:
		return v[:], nil
	case LexKey:
		return v, nil
	case []byte:
		return v, nil
	case int:
		return encodeInt64(int64(v)), nil
	case int64:
		return encodeInt64(v), nil
	case int32:
		return encodeInt32(v), nil
	case int16:
		return encodeInt16(v), nil
	case uint64:
		return encodeUint64(v), nil
	case uint32:
		return encodeUint32(v), nil
	case uint16:
		return encodeUint16(v), nil
	case uint8:
		return []byte{v}, nil
	case float64:
		return encodeFloat64(v), nil
	case float32:
		return encodeFloat32(v), nil
	case bool:
		if v {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	case time.Time:
		return encodeInt64(v.UTC().UnixNano()), nil
	case time.Duration:
		return encodeInt64(int64(v)), nil
	case nil:
		return []byte{Seperator}, nil
	case struct{}:
		return []byte{EndMarker}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
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

// encodeInt64 encodes an int64 into 8 bytes, flipping the sign bit for lexicographic ordering.
// Negative numbers sort before positive ones by XORing with 0x8000000000000000.
func encodeInt64(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v)^0x8000000000000000)
	return buf
}

// encodeInt32 encodes an int32 into 4 bytes, flipping the sign bit for lexicographic ordering.
func encodeInt32(v int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v)^0x80000000)
	return buf
}

// encodeInt16 encodes an int16 into 2 bytes, flipping the sign bit for lexicographic ordering.
func encodeInt16(v int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(v)^0x8000)
	return buf
}

// encodeUint64 encodes a uint64 into 8 bytes, preserving natural order.
func encodeUint64(v uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	return buf
}

// encodeUint32 encodes a uint32 into 4 bytes, preserving natural order.
func encodeUint32(v uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return buf
}

// encodeUint16 encodes a uint16 into 2 bytes, preserving natural order.
func encodeUint16(v uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return buf
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
