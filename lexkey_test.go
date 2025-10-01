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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, hex.EncodeToString(got), "Encoding mismatch")
			}
		})
	}
}

// Test IsEmpty method
func TestLexKeyIsEmpty(t *testing.T) {
	assert.True(t, LexKey{}.IsEmpty())
	assert.False(t, LexKey{0x01}.IsEmpty())
}

// Test JSON serialization and deserialization
func TestLexKeyJSON(t *testing.T) {
	key := Encode("test")
	data, err := json.Marshal(key)
	require.NoError(t, err)

	var decoded LexKey
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, key, decoded)
}

// Test lexicographic ordering
func TestLexKeyOrdering(t *testing.T) {
	key1 := Encode("a")
	key2 := Encode("b")
	assert.True(t, string(key1) < string(key2))
}

// Test EncodeFirst and EncodeLast
func TestLexKeyEncodeLast(t *testing.T) {
	key := Encode("prefix", "a")
	first := EncodeFirst("prefix")
	last := EncodeLast("prefix")

	assert.True(t, hex.EncodeToString(first) <= hex.EncodeToString(key))
	assert.True(t, hex.EncodeToString(last) > hex.EncodeToString(key))
	assert.True(t, hex.EncodeToString(first) < hex.EncodeToString(last))
}

// Test PrimaryKey encoding
func TestPrimaryKey(t *testing.T) {
	pk := NewPrimaryKey(LexKey("partition"), LexKey("row"))
	encoded := pk.Encode()
	assert.Equal(t, "706172746974696f6e00726f77", hex.EncodeToString(encoded))
}

// Test RangeKey encoding
func TestRangeKey(t *testing.T) {
	rk := NewRangeKey(LexKey("part"), LexKey("start"), LexKey("end"))

	lower, upper := rk.Encode(true)
	assert.Equal(t, "70617274007374617274", hex.EncodeToString(lower))
	assert.Equal(t, "7061727400656e64ff", hex.EncodeToString(upper))
}

// Test encoding numbers// Test encoding numbers
func TestNumberEncoding(t *testing.T) {
	intKey := Encode(42)
	assert.Equal(t, "800000000000002a", hex.EncodeToString(intKey))

	floatKey := Encode(3.14)
	assert.Equal(t, "c0091eb851eb851f", hex.EncodeToString(floatKey)) // Corrected

	negativeIntKey := Encode(-42)
	assert.Equal(t, "7fffffffffffffd6", hex.EncodeToString(negativeIntKey))
}

// Test boolean encoding
func TestBooleanEncoding(t *testing.T) {
	trueKey := Encode(true)
	falseKey := Encode(false)
	assert.Equal(t, "01", hex.EncodeToString(trueKey))
	assert.Equal(t, "00", hex.EncodeToString(falseKey))
}

// Test error cases
func TestErrorCases(t *testing.T) {
	var key LexKey

	// Invalid hex string
	err := key.FromHexString("invalidhex")
	assert.Error(t, err)

	assert.Panics(t, func() {
		Encode(make(chan int))
	})
}

// Test nil values
func TestNilValues(t *testing.T) {
	nilKey := Encode(nil)
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

func TestEncodeBoundaryWithoutRowKey(t *testing.T) {
	partKey := LexKey("partition")

	lower := encodeBoundary(partKey, nil, false, true)
	upper := encodeBoundary(partKey, nil, true, true)

	assert.Equal(t, "706172746974696f6e00", hex.EncodeToString(lower))
	assert.Equal(t, "706172746974696f6eff", hex.EncodeToString(upper))
}

func TestLexKeyInt64Sorting(t *testing.T) {
	// Generate a range of int64 values from negative to positive
	values := []int64{-9223372036854775808, -1000000000000, -1000000, -1, 0, 1, 1000000, 1000000000000, 9223372036854775807}

	// Encode each value
	var encodedKeys []LexKey
	for _, v := range values {
		encoded := Encode(v)
		encodedKeys = append(encodedKeys, encoded)
	}

	// Ensure the encoded values are sorted in the expected order
	for i := 0; i < len(encodedKeys)-1; i++ {
		assert.True(t, string(encodedKeys[i]) < string(encodedKeys[i+1]),
			"Encoded int64 values are not sorted correctly: %d vs %d", values[i], values[i+1])
	}
}

func TestLexKeyInt32VsInt64Sorting(t *testing.T) {
	// Define a mix of int32 and int64 values
	values := []any{int32(-2147483648), int64(-9223372036854775808), int32(-100000), int64(-1), int32(0), int64(1), int32(100000), int64(9223372036854775807)}

	// Encode each value
	var encodedKeys []LexKey
	for _, v := range values {
		encoded := Encode(v)
		encodedKeys = append(encodedKeys, encoded)
	}

	// Ensure the encoded values are sorted in the expected order
	for i := 0; i < len(encodedKeys)-1; i++ {
		assert.True(t, string(encodedKeys[i]) < string(encodedKeys[i+1]),
			"Encoded int32/int64 values are not sorted correctly: %v vs %v", values[i], values[i+1])
	}
}

func TestEncodeFloat32NaN(t *testing.T) {
	// Create a NaN float32 value
	nan := float32(math.NaN())

	// Encode the NaN value
	encoded := encodeFloat32(nan)

	// Verify encoding matches expected canonical NaN representation
	assert.Equal(t, "7fc00001", hex.EncodeToString(encoded), "NaN encoding mismatch")
}

func TestEncodeFloat64NaN(t *testing.T) {
	// Create a NaN float32 value
	nan := float64(math.NaN())

	// Encode the NaN value
	encoded := encodeFloat64(nan)

	// Verify encoding matches expected canonical NaN representation
	assert.Equal(t, "7ff8000000000001", hex.EncodeToString(encoded), "NaN encoding mismatch")
}

func TestNewPrimaryKeyNilValues(t *testing.T) {
	// Attempt to create a PrimaryKey with nil values
	assert.Panics(t, func() {
		_ = NewPrimaryKey(nil, nil)
	})
}

func TestLexKeyUnmarshalJSON(t *testing.T) {
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
				require.NoError(t, err)
				assert.Equal(t, tt.expected, key, "Unmarshaled value mismatch")
			}
		})
	}
}

func TestLexKeyToHexString(t *testing.T) {
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

func TestLexKeyFromHexString(t *testing.T) {
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

func TestDecodePrimaryKeySuccessAndError(t *testing.T) {
	pk := NewPrimaryKey(Encode("part"), Encode("row"))
	encoded := pk.Encode()

	decoded, err := DecodePrimaryKey(encoded)
	require.NoError(t, err)
	assert.Equal(t, pk.PartitionKey, decoded.PartitionKey)
	assert.Equal(t, pk.RowKey, decoded.RowKey)

	// invalid input (no separator)
	_, err = DecodePrimaryKey([]byte("invalid"))
	require.Error(t, err)
}

func TestNewRangeKeyPanicsAndFull(t *testing.T) {
	// Panics for nil partition
	assert.Panics(t, func() { _ = NewRangeKey(nil, Encode("a"), Encode("b")) })
	// Panics for nil lower
	assert.Panics(t, func() { _ = NewRangeKey(Encode("p"), nil, Encode("b")) })
	// Panics for nil upper
	assert.Panics(t, func() { _ = NewRangeKey(Encode("p"), Encode("a"), nil) })

	// NewRangeKeyFull panics when partition is nil
	assert.Panics(t, func() { _ = NewRangeKeyFull(nil) })

	// Normal NewRangeKeyFull
	rk := NewRangeKeyFull(Encode("tenant"))
	assert.Equal(t, Empty, rk.StartRowKey)
	assert.Equal(t, Last, rk.EndRowKey)
}

func TestEncodeToBytesAndEncodeIntoUnsupported(t *testing.T) {
	// encodeToBytes wrapper should return an error for unsupported types
	_, err := encodeToBytes(map[int]int{1: 2})
	require.Error(t, err)

	// encodeInto should also return error for unsupported types
	buf := make([]byte, 64)
	_, err = encodeInto(buf, map[string]int{"a": 1})
	require.Error(t, err)
}

func TestEncodeIntoSpecialCasesAndEstimateSizeDefault(t *testing.T) {
	dst := make([]byte, 16)

	// nil encodes to separator
	n, err := encodeInto(dst, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte(Seperator), dst[0])

	// struct{} encodes to EndMarker
	n, err = encodeInto(dst, struct{}{})
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte(EndMarker), dst[0])

	// estimateSize with unsupported/default types
	parts := []any{"a", struct{}{}, map[int]int{1: 2}, int64(5)}
	// expected size: len("a")=1 + struct{}=1 + default(map)=1 + int64=8
	// separators between 4 parts = 3
	// compute manually: 1 + 1 + 1 + 8 + 3 = 14
	expected := 14
	got := estimateSize(parts)
	assert.Equal(t, expected, got)
}

func TestEncodeToBytesSuccessCases(t *testing.T) {
	// string
	bs, err := encodeToBytes("abc")
	require.NoError(t, err)
	assert.Equal(t, "616263", hexEncode(bs))

	// int
	bs, err = encodeToBytes(int64(123))
	require.NoError(t, err)
	assert.Equal(t, "800000000000007b", hexEncode(bs))

	// uint8
	bs, err = encodeToBytes(uint8(255))
	require.NoError(t, err)
	assert.Equal(t, "ff", hexEncode(bs))

	// bool
	bs, err = encodeToBytes(true)
	require.NoError(t, err)
	assert.Equal(t, "01", hexEncode(bs))
}

func TestEstimateSizeSinglePart(t *testing.T) {
	// single string should not add separator
	parts := []any{"single"}
	sz := estimateSize(parts)
	assert.Equal(t, len("single"), sz)
}

// helper for hex encoding bytes in tests
func hexEncode(b []byte) string {
	// simple inline hex encoding to avoid extra imports
	const hextable = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i := 0; i < len(b); i++ {
		out[i*2] = hextable[b[i]>>4]
		out[i*2+1] = hextable[b[i]&0x0f]
	}
	return string(out)
}

func TestEncodeIntoAllSupportedTypes(t *testing.T) {
	buf := make([]byte, 64)
	// prepare values for each case
	vals := []any{
		"somestring",
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		LexKey("lk"),
		[]byte("bytes"),
		int(123),
		int64(1234567890123),
		int32(12345),
		int16(1234),
		uint64(9876543210),
		uint32(1234567),
		uint16(54321),
		uint8(7),
		float64(1.2345),
		float32(2.71828),
		true,
		time.Now(),
		time.Duration(42),
		nil,
		struct{}{},
	}

	for _, v := range vals {
		n, err := encodeInto(buf, v)
		require.NoError(t, err)
		require.Greater(t, n, 0)
	}
}

func TestEstimateSizeAllCases(t *testing.T) {
	parts := []any{
		"abc",
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		LexKey("lk"),
		[]byte("b"),
		int64(1),
		int32(2),
		int16(3),
		uint64(4),
		uint32(5),
		uint16(6),
		uint8(7),
		float64(1.1),
		float32(2.2),
		true,
		time.Now(),
		time.Duration(3),
		nil,
		struct{}{},
		map[int]int{1: 1}, // default branch
	}
	// just ensure it runs and returns > 0
	sz := estimateSize(parts)
	require.Greater(t, sz, 0)
}

// Ensures monotonic ordering across a large int64 span using a reasonable stride
func TestLexKeyInt64LargeRangeSorting(t *testing.T) {
	const min = int64(-10_000_000)
	const max = int64(10_000_000)
	const step = int64(10_000) // 2001 values; keeps test fast

	prev := Encode(min)
	for v := min + step; v <= max; v += step {
		cur := Encode(v)
		if !(Compare(prev, cur) < 0) {
			t.Fatalf("ordering violated at %d: prev=%x cur=%x", v, prev, cur)
		}
		prev = cur
	}

	// Also probe a small dense window around zero to catch sign flip edge-cases
	for v := int64(-5); v < 5; v++ {
		a := Encode(v)
		b := Encode(v + 1)
		if !(Compare(a, b) < 0) {
			t.Fatalf("local ordering violated between %d and %d: %x vs %x", v, v+1, a, b)
		}
	}
}

func TestEncodeSizeAndEncodeInto(t *testing.T) {
	parts := []any{"tenant", "table", "user", int64(42), true}
	need := EncodeSize(parts...)
	buf := make([]byte, need)
	n, err := EncodeInto(buf, parts...)
	require.NoError(t, err)
	require.Equal(t, need, n)

	// Compare with Encode for exact bytes
	got := buf[:n]
	want := Encode(parts...)
	assert.Equal(t, want, LexKey(got))

	// Too small buffer should error
	_, err = EncodeInto(buf[:need-1], parts...)
	require.Error(t, err)
}
