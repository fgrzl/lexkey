package lexkey

import (
	"encoding/hex"
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test encoding various types into LexKey
func TestNewLexKey(t *testing.T) {
	tests := []struct {
		name     string
		parts    []any
		expected string
		wantErr  bool
	}{
		{"String encoding", []any{"hello"}, "68656c6c6f", false},
		{"UUID encoding", []any{uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")}, "550e8400e29b41d4a716446655440000", false},
		{"Int encoding", []any{123}, "800000000000007b", false},
		{"Negative int encoding", []any{-123}, "7fffffffffffff85", false},

		{"Int16 encoding", []any{int16(123)}, "807b", false},
		{"Negative int16 encoding", []any{int16(-123)}, "7f85", false},

		{"Int32 encoding", []any{int32(123)}, "8000007b", false},
		{"Negative int32 encoding", []any{int32(-123)}, "7fffff85", false},
		{"Int64 encoding", []any{int64(123)}, "800000000000007b", false},
		{"Negative int64 encoding", []any{int64(-123)}, "7fffffffffffff85", false},
		{"UInt8 encoding", []any{uint8(123)}, "7b", false},
		{"UInt16 encoding", []any{uint16(123)}, "007b", false},
		{"UInt32 encoding", []any{uint32(123)}, "0000007b", false},
		{"UInt64 encoding", []any{uint64(123)}, "000000000000007b", false},
		{"Float32 encoding", []any{float32(3.14)}, "c048f5c3", false},
		{"Float64 encoding", []any{3.14}, "c0091eb851eb851f", false},
		{"Negative float32 encoding", []any{float32(-3.14)}, "3fb70a3c", false},
		{"Negative float64 encoding", []any{-3.14}, "3ff6e147ae147ae0", false},
		{"Boolean true", []any{true}, "01", false},
		{"Boolean false", []any{false}, "00", false},
		{"Byte slice encoding", []any{[]byte("data")}, "64617461", false},
		{"Time encoding", []any{time.Unix(0, 0)}, "8000000000000000", false},
		{"Future time encoding", []any{time.Unix(1700000000, 0)}, "97979cfe362a0000", false},
		{"Duration encoding", []any{time.Duration(42)}, "800000000000002a", false},
		{"Nil input", []any{nil}, "00", false},
		{"Empty input", []any{}, "", true},
		{"Multiple types", []any{"foo", 42, true}, "666f6f00800000000000002a0001", false},
		{"Complex mixed types", []any{"key", uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), -99, 4.2, true}, "6b657900550e8400e29b41d4a716446655440000007fffffffffffff9d00c010cccccccccccd0001", false},
		{"Unsupported type", []any{struct{}{}}, "ff", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLexKey(tt.parts...)
			if tt.wantErr {
				require.Error(t, err, "Expected an error but got nil")
			} else {
				require.NoError(t, err, "Unexpected error: %v", err)
				assert.Equal(t, tt.expected, hex.EncodeToString(got), "Encoding mismatch")
			}
		})
	}
}

// Test Encode function
func TestEncode(t *testing.T) {
	key, err := Encode("hello", 42)
	require.NoError(t, err)
	expected := "68656c6c6f00800000000000002a" // Corrected to match "hello" and 42
	assert.Equal(t, expected, hex.EncodeToString(key))
}

// Test IsEmpty method
func TestLexKey_IsEmpty(t *testing.T) {
	assert.True(t, LexKey{}.IsEmpty())
	assert.False(t, LexKey{0x01}.IsEmpty())
}

// Test JSON serialization and deserialization
func TestLexKey_JSON(t *testing.T) {
	key, _ := Encode("test")
	data, err := json.Marshal(key)
	require.NoError(t, err)

	var decoded LexKey
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, key, decoded)
}

// Test lexicographic ordering
func TestLexKey_Ordering(t *testing.T) {
	key1, _ := Encode("a")
	key2, _ := Encode("b")
	assert.True(t, string(key1) < string(key2))
}

// Test EncodeFirst and EncodeLast
func TestLexKey_EncodeLast(t *testing.T) {
	key, _ := Encode("middle")
	last := key.EncodeLast()

	assert.True(t, string(key) < string(last))                         // Existing check
	assert.True(t, hex.EncodeToString(last) > hex.EncodeToString(key)) // Additional verification
}

// Test PrimaryKey encoding
func TestPrimaryKey(t *testing.T) {
	pk, err := NewPrimaryKey(LexKey("partition"), LexKey("row"))
	require.NoError(t, err)
	encoded := pk.Encode()
	assert.Equal(t, "706172746974696f6e00726f77", hex.EncodeToString(encoded))
}

// Test RangeKey encoding
func TestRangeKey(t *testing.T) {
	rk := RangeKey{
		PartitionKey: LexKey("part"),
		StartRowKey:  LexKey("start"),
		EndRowKey:    LexKey("end"),
	}
	lower, upper := rk.Encode(true)
	assert.Equal(t, "70617274007374617274", hex.EncodeToString(lower))
	assert.Equal(t, "7061727400656e64ff", hex.EncodeToString(upper))
}

// Test encoding numbers// Test encoding numbers
func TestNumberEncoding(t *testing.T) {
	intKey, _ := Encode(42)
	assert.Equal(t, "800000000000002a", hex.EncodeToString(intKey))

	floatKey, _ := Encode(3.14)
	assert.Equal(t, "c0091eb851eb851f", hex.EncodeToString(floatKey)) // Corrected

	negativeIntKey, _ := Encode(-42)
	assert.Equal(t, "7fffffffffffffd6", hex.EncodeToString(negativeIntKey))
}

// Test boolean encoding
func TestBooleanEncoding(t *testing.T) {
	trueKey, _ := Encode(true)
	falseKey, _ := Encode(false)
	assert.Equal(t, "01", hex.EncodeToString(trueKey))
	assert.Equal(t, "00", hex.EncodeToString(falseKey))
}

// Test error cases
func TestErrorCases(t *testing.T) {
	var key LexKey

	// Invalid hex string
	err := key.FromHexString("invalidhex")
	assert.Error(t, err)

	// Unsupported type
	_, err = Encode(make(chan int))
	assert.Error(t, err)
}

// Test nil values
func TestNilValues(t *testing.T) {
	nilKey, _ := Encode(nil)
	assert.Equal(t, "00", hex.EncodeToString(nilKey))
}

// Test encodeBoundary function
func TestEncodeBoundary(t *testing.T) {
	partKey := LexKey("partition")
	rowKey := LexKey("row")

	lower := encodeBoundary(partKey, rowKey, false, true)
	upper := encodeBoundary(partKey, rowKey, true, true)

	assert.Equal(t, "706172746974696f6e00726f77", hex.EncodeToString(lower))
	assert.Equal(t, "706172746974696f6e00726f77ff", hex.EncodeToString(upper))
}

func TestEncodeBoundary_WithoutRowKey(t *testing.T) {
	partKey := LexKey("partition")

	lower := encodeBoundary(partKey, nil, false, true)
	upper := encodeBoundary(partKey, nil, true, true)

	assert.Equal(t, "706172746974696f6e00", hex.EncodeToString(lower))
	assert.Equal(t, "706172746974696f6eff", hex.EncodeToString(upper))
}

func TestLexKey_Int64Sorting(t *testing.T) {
	// Generate a range of int64 values from negative to positive
	values := []int64{-9223372036854775808, -1000000000000, -1000000, -1, 0, 1, 1000000, 1000000000000, 9223372036854775807}

	// Encode each value
	var encodedKeys []LexKey
	for _, v := range values {
		encoded, err := Encode(v)
		require.NoError(t, err)
		encodedKeys = append(encodedKeys, encoded)
	}

	// Ensure the encoded values are sorted in the expected order
	for i := 0; i < len(encodedKeys)-1; i++ {
		assert.True(t, string(encodedKeys[i]) < string(encodedKeys[i+1]),
			"Encoded int64 values are not sorted correctly: %d vs %d", values[i], values[i+1])
	}
}

func TestLexKey_Int32VsInt64Sorting(t *testing.T) {
	// Define a mix of int32 and int64 values
	values := []any{int32(-2147483648), int64(-9223372036854775808), int32(-100000), int64(-1), int32(0), int64(1), int32(100000), int64(9223372036854775807)}

	// Encode each value
	var encodedKeys []LexKey
	for _, v := range values {
		encoded, err := Encode(v)
		require.NoError(t, err)
		encodedKeys = append(encodedKeys, encoded)
	}

	// Ensure the encoded values are sorted in the expected order
	for i := 0; i < len(encodedKeys)-1; i++ {
		assert.True(t, string(encodedKeys[i]) < string(encodedKeys[i+1]),
			"Encoded int32/int64 values are not sorted correctly: %v vs %v", values[i], values[i+1])
	}
}

func TestEncodeFloat32_NaN(t *testing.T) {
	// Create a NaN float32 value
	nan := float32(math.NaN())

	// Encode the NaN value
	encoded := encodeFloat32(nan)

	// Verify encoding matches expected canonical NaN representation
	assert.Equal(t, "7fc00001", hex.EncodeToString(encoded), "NaN encoding mismatch")
}

func TestEncodeFloat64_NaN(t *testing.T) {
	// Create a NaN float32 value
	nan := float64(math.NaN())

	// Encode the NaN value
	encoded := encodeFloat64(nan)

	// Verify encoding matches expected canonical NaN representation
	assert.Equal(t, "7ff8000000000001", hex.EncodeToString(encoded), "NaN encoding mismatch")
}

func TestNewPrimaryKey_NilValues(t *testing.T) {
	// Attempt to create a PrimaryKey with nil values
	_, err := NewPrimaryKey(nil, nil)

	// Expect an error
	require.Error(t, err, "Expected error when both partitionKey and rowKey are nil")
	assert.Equal(t, "partitionKey and rowKey cannot be nil", err.Error())
}

func TestLexKey_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LexKey
		wantErr  bool
	}{
		{"Null JSON", "null", LexKey{}, false},
		{"Valid hex string", `"68656c6c6f"`, LexKey("hello"), false},
		{"Invalid JSON format", `123`, nil, true},
		{"Malformed hex string", `"invalidhex"`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key LexKey
			err := key.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got nil")
			} else {
				require.NoError(t, err, "Unexpected error: %v", err)
				assert.Equal(t, tt.expected, key, "Unmarshaled value mismatch")
			}
		})
	}
}

func TestLexKey_ToHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    LexKey
		expected string
	}{
		{"Empty LexKey", LexKey{}, ""},
		{"Valid LexKey", LexKey("hello"), "68656c6c6f"},
		{"Nil LexKey", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.input.ToHexString()
			assert.Equal(t, tt.expected, output, "Hex encoding mismatch")
		})
	}
}

func TestLexKey_FromHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LexKey
		wantErr  bool
	}{
		{"Empty string", "", LexKey{}, false},
		{"Valid hex string", "68656c6c6f", LexKey("hello"), false},
		{"Invalid hex string", "invalidhex", nil, true},
		{"Odd-length hex string", "123", nil, true},
		{"Large hex input", "746573746b6579", LexKey("testkey"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key LexKey
			err := key.FromHexString(tt.input)

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got nil")
			} else {
				require.NoError(t, err, "Unexpected error: %v", err)
				assert.Equal(t, tt.expected, key, "Hex decoding mismatch")
			}
		})
	}
}
