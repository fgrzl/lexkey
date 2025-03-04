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
	separatorByte = 0x00
	endMarkerByte = 0xFF
)

// LexKey represents an encoded key as a byte slice, optimized for lexicographic sorting.
// An empty LexKey (length 0) is valid and distinct from nil.
type LexKey []byte

// NewLexKey constructs a LexKey from a list of parts, ensuring lexicographic sorting.
// Returns an error if parts is empty or contains unsupported types.
func NewLexKey(parts ...any) (LexKey, error) {
	if len(parts) == 0 {
		return nil, errors.New("empty keys are not allowed")
	}

	var result []byte
	for i, part := range parts {
		encoded, err := encodeToBytes(part)
		if err != nil {
			return nil, err
		}
		result = append(result, encoded...)
		if i < len(parts)-1 {
			result = append(result, separatorByte)
		}
	}
	return result, nil
}

// Encode constructs a LexKey, returning an error if encoding fails.
func Encode(parts ...any) (LexKey, error) {
	return NewLexKey(parts...)
}

// IsEmpty checks if the LexKey is empty (length 0). A nil LexKey is considered empty.
func (e LexKey) IsEmpty() bool {
	return len(e) == 0
}

// ToHexString converts the LexKey to a hexadecimal string.
// Returns empty string for empty or nil LexKey.
func (e LexKey) ToHexString() string {
	if len(e) == 0 {
		return ""
	}
	return hex.EncodeToString(e)
}

// FromHexString decodes a hexadecimal string back into a LexKey.
// Sets to empty slice (not nil) for empty input string.
func (e *LexKey) FromHexString(hexStr string) error {
	if len(hexStr) == 0 {
		*e = []byte{} // Empty slice instead of nil
		return nil
	}
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		if errors.Is(err, hex.ErrLength) {
			return fmt.Errorf("invalid hex string length: %w", err)
		}
		return fmt.Errorf("failed to decode hex string: %w", err)
	}
	*e = bytes
	return nil
}

// MarshalJSON encodes LexKey as a hex string for JSON.
func (e LexKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.ToHexString())
}

// UnmarshalJSON decodes a hex string from JSON into a LexKey.
// Handles JSON null by setting to empty slice.
func (e *LexKey) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*e = []byte{}
		return nil
	}
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return fmt.Errorf("failed to unmarshal LexKey hex string: %w", err)
	}
	return e.FromHexString(hexStr)
}

// EncodeLast returns the last lexicographically sortable key in a range.
func (e LexKey) EncodeLast() []byte {
	newKey := make([]byte, len(e)+1)
	copy(newKey, e)
	newKey[len(e)] = endMarkerByte
	return newKey
}

// encodeToBytes converts a value to a lexicographically sortable byte representation.
func encodeToBytes(v any) ([]byte, error) {
	switch v := v.(type) {
	case string:
		return []byte(v), nil
	case uuid.UUID:
		return v[:], nil
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
		return []byte{0x00}, nil
	case struct{}:
		return []byte{0xFF}, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", v)
	}
}

func encodeInt64(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v)^0x8000000000000000) // Flip sign bit
	return buf
}

func encodeInt32(v int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v)^0x80000000) // Flip sign bit
	return buf
}

func encodeInt16(v int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(v)^0x8000) // Flip sign bit
	return buf
}

func encodeUint64(v uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	return buf
}

func encodeUint32(v uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return buf
}

func encodeUint16(v uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return buf
}

func encodeFloat64(v float64) []byte {
	buf := make([]byte, 8)
	if math.IsNaN(v) {
		// Encode NaN as a special value - choose a consistent representation
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

func encodeFloat32(v float32) []byte {
	buf := make([]byte, 4)
	if math.IsNaN(float64(v)) {
		// Encode NaN as a special value
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

// PrimaryKey represents a composite key for key-value storage.
type PrimaryKey struct {
	PartitionKey LexKey
	RowKey       LexKey
}

// NewPrimaryKey creates a new PrimaryKey from partition and row keys.
// Returns an error if either key is nil.
func NewPrimaryKey(partitionKey, rowKey LexKey) (PrimaryKey, error) {
	if partitionKey == nil || rowKey == nil {
		return PrimaryKey{}, errors.New("partitionKey and rowKey cannot be nil")
	}
	return PrimaryKey{
		PartitionKey: partitionKey,
		RowKey:       rowKey,
	}, nil
}

// Encode concatenates PartitionKey and RowKey with a separator.
func (pk PrimaryKey) Encode() LexKey {
	result := make(LexKey, len(pk.PartitionKey)+len(pk.RowKey)+1)
	n := copy(result, pk.PartitionKey)
	result[n] = separatorByte
	copy(result[n+1:], pk.RowKey)
	return result
}

// RangeKey defines a range query over keys.
type RangeKey struct {
	PartitionKey LexKey
	StartRowKey  LexKey
	EndRowKey    LexKey
}

// Encode encodes the range boundaries for range queries.
func (rk RangeKey) Encode(withPartitionKey bool) (lower, upper LexKey) {
	lower = encodeBoundary(rk.PartitionKey, rk.StartRowKey, false, withPartitionKey)
	upper = encodeBoundary(rk.PartitionKey, rk.EndRowKey, true, withPartitionKey)
	return lower, upper
}

// encodeBoundary encodes range boundaries for lexicographic ordering.
func encodeBoundary(partitionKey, rowKey LexKey, isUpper, withPartitionKey bool) LexKey {
	var size int
	if withPartitionKey {
		size = len(partitionKey)
	}
	if len(rowKey) > 0 {
		size += 1 + len(rowKey) // separator + rowKey
		if isUpper {
			size++ // extra byte for end marker
		}
	} else {
		size += 1 // Just separator or end marker
	}
	result := make(LexKey, size)

	n := 0
	if withPartitionKey {
		n += copy(result, partitionKey)
	}
	if len(rowKey) == 0 {
		result[n] = ternary(isUpper, endMarkerByte, separatorByte)
	} else {
		result[n] = separatorByte
		n++
		copy(result[n:], rowKey)
		if isUpper {
			result[len(result)-1] = endMarkerByte
		}
	}
	return result
}

func ternary(cond bool, a, b byte) byte {
	if cond {
		return a
	}
	return b
}
