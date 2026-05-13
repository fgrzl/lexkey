package lexkey

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/fgrzl/lexkey/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test encoding various types into LexKey
func TestShouldEncodePartsIntoLexKey(t *testing.T) {
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

		{"Int16 encoding", []any{int16(123)}, "800000000000007b", false},
		{"Negative int16 encoding", []any{int16(-123)}, "7fffffffffffff85", false},

		{"Int32 encoding", []any{int32(123)}, "800000000000007b", false},
		{"Negative int32 encoding", []any{int32(-123)}, "7fffffffffffff85", false},
		{"Int64 encoding", []any{int64(123)}, "800000000000007b", false},
		{"Negative int64 encoding", []any{int64(-123)}, "7fffffffffffff85", false},
		{"UInt8 encoding", []any{uint8(123)}, "000000000000007b", false},
		{"UInt16 encoding", []any{uint16(123)}, "000000000000007b", false},
		{"UInt32 encoding", []any{uint32(123)}, "000000000000007b", false},
		{"UInt64 encoding", []any{uint64(123)}, "000000000000007b", false},
		{"Float32 encoding", []any{float32(3.14)}, "c0091eb860000000", false},
		{"Float64 encoding", []any{3.14}, "c0091eb851eb851f", false},
		{"Negative float32 encoding", []any{float32(-3.14)}, "3ff6e1479fffffff", false},
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
				test.AssertHexEqual(t, tt.expected, got, "Encoding mismatch")
			}
		})
	}
}

// IsEmpty behavior
func TestShouldReportEmptyStateForLexKey(t *testing.T) {
	// Arrange
	empty := LexKey{}
	nonEmpty := LexKey{0x01}
	// Act / Assert
	assert.True(t, empty.IsEmpty())
	assert.False(t, nonEmpty.IsEmpty())
}

// JSON behaviors
func TestShouldMarshalLexKeyToJSON(t *testing.T) {
	// Arrange
	key := Encode("test")
	// Act
	data, err := json.Marshal(key)
	// Assert
	require.NoError(t, err)
	require.True(t, len(data) > 0)
}

func TestMarshalUnmarshalJSONVariants(t *testing.T) {
	tests := []struct {
		name     string
		input    LexKey
		expected string // expected JSON
	}{
		{"nil lexkey", nil, `""`},
		{"empty lexkey", LexKey{}, `""`},
		{"non-empty", Encode("hello"), `"68656c6c6f"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := tt.input.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Unmarshal back into a fresh LexKey
			var dest LexKey
			err = dest.UnmarshalJSON(data)
			require.NoError(t, err)

			// For nil and empty both should be empty slice (length 0)
			if len(tt.input) == 0 {
				assert.Equal(t, LexKey{}, dest)
			} else {
				assert.Equal(t, tt.input, dest)
			}
		})
	}

	// Ensure JSON null decodes to empty LexKey
	var d LexKey
	require.NoError(t, d.UnmarshalJSON([]byte("null")))
	assert.Equal(t, LexKey{}, d)
}

// Lexicographic ordering behavior
func TestShouldOrderLexicographicallyGivenDifferentStrings(t *testing.T) {
	// Arrange
	a := Encode("a")
	b := Encode("b")
	// Act
	cmp := Compare(a, b)
	// Assert
	assert.Less(t, cmp, 0)
}

// EncodeFirst / EncodeLast behaviors
func TestShouldEncodeFirstBeforeAnyExtension(t *testing.T) {
	// Arrange
	key := Encode("prefix", "a")
	first := EncodeFirst("prefix")
	// Act / Assert
	test.AssertHexLessOrEqual(t, first, key)
}

func TestShouldEncodeLastAfterAnyExtension(t *testing.T) {
	// Arrange
	key := Encode("prefix", "a")
	last := EncodeLast("prefix")
	// Act / Assert
	test.AssertHexGreater(t, last, key)
}

// PrimaryKey behavior
func TestShouldEncodePrimaryKeyGivenPartitionAndRow(t *testing.T) {
	// Arrange
	pk := NewPrimaryKey(LexKey("partition"), LexKey("row"))
	// Act
	encoded := pk.Encode()
	// Assert
	test.AssertHexEqual(t, "706172746974696f6e00726f77", encoded)
}

// RangeKey behavior
func TestShouldEncodeRangeKeyGivenPartitionAndBounds(t *testing.T) {
	// Arrange
	rk := NewRangeKey(LexKey("part"), LexKey("start"), LexKey("end"))
	// Act
	lower, upper := rk.Encode(true)
	// Assert
	test.AssertHexEqual(t, "70617274007374617274", lower)
	test.AssertHexEqual(t, "7061727400656e64ff", upper)
}

// Number encoding behaviors
func TestShouldEncodeIntAsLexOrderedBytes(t *testing.T) {
	// Arrange / Act
	intKey := Encode(42)
	// Assert
	test.AssertHexEqual(t, "800000000000002a", intKey)
}

func TestShouldEncodeNegativeIntAsLexOrderedBytes(t *testing.T) {
	// Arrange / Act
	negativeIntKey := Encode(-42)
	// Assert
	test.AssertHexEqual(t, "7fffffffffffffd6", negativeIntKey)
}

// Boolean encoding behavior
func TestShouldEncodeBooleansAsSingleByte(t *testing.T) {
	// Arrange / Act
	trueKey := Encode(true)
	falseKey := Encode(false)
	// Assert
	test.AssertHexEqual(t, "01", trueKey)
	test.AssertHexEqual(t, "00", falseKey)
}

func TestShouldPanicWhenEncodingUnsupportedType(t *testing.T) {
	// Arrange / Act / Assert
	assert.Panics(t, func() {
		Encode(make(chan int))
	})
}

// Nil encoding behavior
func TestShouldEncodeNilAsSeparatorByte(t *testing.T) {
	// Arrange / Act
	nilKey := Encode(nil)
	// Assert
	test.AssertHexEqual(t, "00", nilKey)
}

// encodeBoundary behaviors
func TestShouldEncodeBoundaryGivenPartitionAndRow(t *testing.T) {
	// Arrange
	partKey := LexKey("partition")
	rowKey := LexKey("row")
	// Act
	lower := encodeBoundary(partKey, rowKey, false, true)
	upper := encodeBoundary(partKey, rowKey, true, true)
	// Assert
	assert.Equal(t, "706172746974696f6e00726f77", hex.EncodeToString(lower))
	assert.Equal(t, "706172746974696f6e00726f77ff", hex.EncodeToString(upper))
}

func TestShouldEncodeBoundaryWithoutRowKey(t *testing.T) {
	// Arrange
	partKey := LexKey("partition")
	// Act
	lower := encodeBoundary(partKey, nil, false, true)
	upper := encodeBoundary(partKey, nil, true, true)
	// Assert
	assert.Equal(t, "706172746974696f6e00", hex.EncodeToString(lower))
	assert.Equal(t, "706172746974696f6eff", hex.EncodeToString(upper))
}

func TestShouldSortEncodedInt64ValuesInNumericOrder(t *testing.T) {
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
		assert.True(t, Compare(encodedKeys[i], encodedKeys[i+1]) < 0,
			"Encoded int64 values are not sorted correctly: %d vs %d", values[i], values[i+1])
	}
}

func TestShouldSortEncodedInt32AndInt64TogetherInNumericOrder(t *testing.T) {
	// Define a mix of int32 and int64 values
	values := []any{int64(-9223372036854775808), int32(-2147483648), int32(-100000), int64(-1), int32(0), int64(1), int32(100000), int64(9223372036854775807)}

	// Encode each value
	var encodedKeys []LexKey
	for _, v := range values {
		encoded := Encode(v)
		encodedKeys = append(encodedKeys, encoded)
	}

	// Ensure the encoded values are sorted in the expected order
	for i := 0; i < len(encodedKeys)-1; i++ {
		assert.True(t, Compare(encodedKeys[i], encodedKeys[i+1]) < 0,
			"Encoded int32/int64 values are not sorted correctly: %v vs %v", values[i], values[i+1])
	}
}

func TestShouldEncodeFloat32NaNAsCanonical(t *testing.T) {
	// Create a NaN float32 value
	nan := float32(math.NaN())

	// Encode the NaN value
	encoded := encodeFloat32(nan)

	// Verify encoding matches expected canonical NaN representation
	assert.Equal(t, "7fc00001", hex.EncodeToString(encoded), "NaN encoding mismatch")
}

func TestShouldEncodeFloat64NaNAsCanonical(t *testing.T) {
	// Create a NaN float32 value
	nan := float64(math.NaN())

	// Encode the NaN value
	encoded := encodeFloat64(nan)

	// Verify encoding matches expected canonical NaN representation
	assert.Equal(t, "7ff8000000000001", hex.EncodeToString(encoded), "NaN encoding mismatch")
}

func TestShouldPanicWhenPrimaryKeyHasNilValues(t *testing.T) {
	// Attempt to create a PrimaryKey with nil values
	assert.Panics(t, func() {
		_ = NewPrimaryKey(nil, nil)
	})
}

func TestShouldUnmarshalLexKeyFromJSON(t *testing.T) {
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

func TestShouldConvertLexKeyToHexString(t *testing.T) {
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

func TestShouldParseHexStringIntoLexKey(t *testing.T) {
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

func TestShouldDecodePrimaryKeyGivenValidInput(t *testing.T) {
	// Arrange
	pk := NewPrimaryKey(Encode("part"), Encode("row"))
	encoded := pk.Encode()
	// Act
	decoded, err := DecodePrimaryKey(encoded)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, pk.PartitionKey, decoded.PartitionKey)
	assert.Equal(t, pk.RowKey, decoded.RowKey)
}

func TestShouldReturnErrorWhenPrimaryKeyInputHasNoSeparator(t *testing.T) {
	// Arrange / Act
	_, err := DecodePrimaryKey([]byte("invalid"))
	// Assert
	require.Error(t, err)
}

func TestShouldPanicWhenRangeKeyPartitionIsNil(t *testing.T) {
	// Consolidated table-driven tests for nil argument panics
	tests := []struct {
		name string
		part LexKey
		low  LexKey
		up   LexKey
	}{
		{"nil partition", nil, Encode("a"), Encode("b")},
		{"nil lower", Encode("p"), nil, Encode("b")},
		{"nil upper", Encode("p"), Encode("a"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange/Act/Assert: constructing a RangeKey with any nil should panic
			assert.Panics(t, func() { _ = NewRangeKey(tt.part, tt.low, tt.up) })
		})
	}
}

func TestShouldPanicWhenRangeKeyFullPartitionIsNil(t *testing.T) {
	// Arrange / Act / Assert
	assert.Panics(t, func() { _ = NewRangeKeyFull(nil) })
}

func TestShouldCreateFullRangeKeyGivenPartition(t *testing.T) {
	// Arrange / Act
	rk := NewRangeKeyFull(Encode("tenant"))
	// Assert
	assert.Equal(t, Empty, rk.StartRowKey)
	assert.Equal(t, Last, rk.EndRowKey)
}

func TestShouldReturnErrorWhenEncodeToBytesReceivesUnsupportedType(t *testing.T) {
	// Arrange / Act
	_, err := encodeToBytes(map[int]int{1: 2})
	// Assert
	require.Error(t, err)
}

func TestShouldReturnErrorWhenEncodeIntoReceivesUnsupportedType(t *testing.T) {
	// Arrange
	buf := make([]byte, 64)
	// Act
	_, err := encodeInto(buf, map[string]int{"a": 1})
	// Assert
	require.Error(t, err)
}

func TestShouldEncodeNilAndStructMarkersWithEncodeInto(t *testing.T) {
	// Arrange
	dst := make([]byte, 16)
	// Act / Assert: nil -> separator
	n, err := encodeInto(dst, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte(Separator), dst[0])
	// Act / Assert: struct{} -> EndMarker
	n, err = encodeInto(dst, struct{}{})
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte(EndMarker), dst[0])
}

func TestShouldEstimateSizeForDefaultBranch(t *testing.T) {
	// Arrange
	parts := []any{"a", struct{}{}, map[int]int{1: 2}, int64(5)}
	// expected size: len("a")=1 + struct{}=1 + default(map)=1 + int64=8
	// separators between 4 parts = 3 => total 14
	expected := 14
	// Act
	got := estimateSize(parts)
	// Assert
	assert.Equal(t, expected, got)
}

func TestShouldEncodeToBytesForSupportedTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"String", "abc", "616263"},
		{"Int64", int64(123), "800000000000007b"},
		{"Uint8", uint8(255), "00000000000000ff"},
		{"BoolTrue", true, "01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange / Act
			bs, err := encodeToBytes(tt.input)
			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expected, test.HexEncode(bs))
		})
	}
}

func TestShouldEstimateSizeForSinglePart(t *testing.T) {
	// single string should not add separator
	parts := []any{"single"}
	sz := estimateSize(parts)
	assert.Equal(t, len("single"), sz)
}

// helper for hex encoding bytes in tests
// ...existing code...

func TestShouldEncodeAllSupportedTypesWithoutError(t *testing.T) {
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

func TestShouldEstimateSizeAcrossAllCases(t *testing.T) {
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
func TestShouldMaintainOrderingAcrossLargeInt64Range(t *testing.T) {
	const lo = int64(-10_000_000)
	const hi = int64(10_000_000)
	const step = int64(10_000) // 2001 values; keeps test fast

	prev := Encode(lo)
	for v := lo + step; v <= hi; v += step {
		cur := Encode(v)
		require.Less(t, Compare(prev, cur), 0, "ordering violated at %d: prev=%x cur=%x", v, prev, cur)
		prev = cur
	}

	// Also probe a small dense window around zero to catch sign flip edge-cases
	for v := int64(-5); v < 5; v++ {
		a := Encode(v)
		b := Encode(v + 1)
		require.Less(t, Compare(a, b), 0, "local ordering violated between %d and %d: %x vs %x", v, v+1, a, b)
	}
}

func TestShouldEncodeIntoGivenPreallocatedBuffer(t *testing.T) {
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
}

func TestShouldReturnErrorWhenEncodeIntoBufferIsTooSmall(t *testing.T) {
	// Arrange
	parts := []any{"tenant", "table", "user", int64(42), true}
	need := EncodeSize(parts...)
	buf := make([]byte, need-1)
	// Act
	_, err := EncodeInto(buf, parts...)
	// Assert
	require.Error(t, err)
}

func TestShouldSortCrossWidthUnsignedIntegersWithCanonicalWidth(t *testing.T) {
	// Arrange: values in increasing numeric order with mixed widths
	inputs := []any{uint32(1), uint64(1), uint32(2), uint64(3)}
	// Act: encode each using canonical width (all unsigned -> uint64)
	keys := make([]LexKey, len(inputs))
	for i, v := range inputs {
		keys[i] = EncodeCanonicalWidth(v)
	}
	// Assert: monotonic ordering
	for i := 0; i < len(keys)-1; i++ {
		cmp := Compare(keys[i], keys[i+1])
		// Equal numeric values across widths yield identical bytes after canonicalization
		if i == 0 { // 1 vs 1
			assert.LessOrEqual(t, cmp, 0, "index %d vs %d", i, i+1)
		} else {
			assert.Less(t, cmp, 0, "index %d vs %d", i, i+1)
		}
	}
}

func TestShouldSortCrossWidthSignedIntegersWithCanonicalWidth(t *testing.T) {
	// Arrange: mixed signed widths around zero
	inputs := []any{int32(-1), int64(0), int32(1), int64(2)}
	// Act
	keys := make([]LexKey, len(inputs))
	for i, v := range inputs {
		keys[i] = EncodeCanonicalWidth(v)
	}
	// Assert
	for i := 0; i < len(keys)-1; i++ {
		assert.Less(t, Compare(keys[i], keys[i+1]), 0)
	}
}

func TestShouldSortCrossWidthFloatsWithCanonicalWidth(t *testing.T) {
	// Arrange: float32 vs float64
	inputs := []any{float32(-3.5), float64(0.0), float32(1.25), float64(2.5)}
	// Act
	prev := EncodeCanonicalWidth(inputs[0])
	for i := 1; i < len(inputs); i++ {
		cur := EncodeCanonicalWidth(inputs[i])
		assert.Less(t, Compare(prev, cur), 0, "index %d vs %d", i-1, i)
		prev = cur
	}
}

func TestShouldEncodeIntoCanonicalWidthGivenPreallocatedBuffer(t *testing.T) {
	// Arrange
	parts := []any{"tenant", uint32(1), uint64(2), float32(1.5), float64(2.5)}
	need := EncodeSizeCanonicalWidth(parts...)
	buf := make([]byte, need)
	// Act
	n, err := EncodeIntoCanonicalWidth(buf, parts...)
	// Assert
	require.NoError(t, err)
	require.Equal(t, need, n)

	// Compare with EncodeCanonicalWidth for exact bytes
	want := EncodeCanonicalWidth(parts...)
	assert.Equal(t, want, LexKey(buf[:n]))
}

func TestShouldReturnErrorWhenEncodeIntoCanonicalWidthBufferIsTooSmall(t *testing.T) {
	// Arrange
	parts := []any{"t", uint32(1), uint64(2)}
	need := EncodeSizeCanonicalWidth(parts...)
	buf := make([]byte, need-1)
	// Act
	_, err := EncodeIntoCanonicalWidth(buf, parts...)
	// Assert
	require.Error(t, err)
}

// Ensure helper encoders for floats are covered (positive/negative finite paths)
func TestEncodeFloat64HelperMatchesExpected(t *testing.T) {
	gotPos := encodeFloat64(3.14)
	gotNeg := encodeFloat64(-3.14)
	assert.Equal(t, "c0091eb851eb851f", hex.EncodeToString(gotPos))
	assert.Equal(t, "3ff6e147ae147ae0", hex.EncodeToString(gotNeg))
}

func TestEncodeFloat32HelperMatchesExpected(t *testing.T) {
	gotPos := encodeFloat32(float32(3.14))
	gotNeg := encodeFloat32(float32(-3.14))
	// 32-bit transformed encodings
	assert.Equal(t, "c048f5c3", hex.EncodeToString(gotPos))
	assert.Equal(t, "3fb70a3c", hex.EncodeToString(gotNeg))
}

// Cover canonicalization of narrower signed types (int8)
func TestCanonicalizeInt8ViaEncodeCanonicalWidth(t *testing.T) {
	a := EncodeCanonicalWidth(int8(-5))
	b := EncodeCanonicalWidth(int64(-5))
	assert.Equal(t, b, a)
	assert.Equal(t, "7ffffffffffffffb", a.ToHexString())
}

// Exercise RangeKey.Encode with withPartitionKey=false
func TestRangeKeyEncodeWithoutPartitionKey(t *testing.T) {
	rk := RangeKey{PartitionKey: LexKey("p"), StartRowKey: LexKey("r"), EndRowKey: LexKey("r")}
	lower, upper := rk.Encode(false)
	assert.Equal(t, "0072", hex.EncodeToString(lower))
	assert.Equal(t, "0072ff", hex.EncodeToString(upper))
}

// Cover MarshalText/UnmarshalText empty branches
func TestMarshalUnmarshalTextEmpty(t *testing.T) {
	var e LexKey
	// MarshalText on empty
	txt, err := e.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, 0, len(txt))

	// UnmarshalText with empty input
	var d LexKey
	err = d.UnmarshalText([]byte(""))
	require.NoError(t, err)
	assert.Equal(t, 0, len(d))
}

func TestShouldEncodeIntoCanonicalWidthWithEmptyParts(t *testing.T) {
	// Arrange
	buf := make([]byte, 1)
	// Act
	n, err := EncodeIntoCanonicalWidth(buf)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestShouldReturnZeroSizeForZeroParts(t *testing.T) {
	// Arrange / Act
	a := EncodeSize()
	b := EncodeSizeCanonicalWidth()
	// Assert
	assert.Equal(t, 0, a)
	assert.Equal(t, 0, b)
}

func TestShouldErrorWhenNewLexKeyCanonicalWidthReceivesNoParts(t *testing.T) {
	// Arrange / Act
	_, err := NewLexKeyCanonicalWidth()
	// Assert
	require.Error(t, err)
}

func TestShouldErrorOnUnmarshalTextGivenInvalidHex(t *testing.T) {
	// Arrange
	var e LexKey
	// Act
	err := e.UnmarshalText([]byte("zzz"))
	// Assert
	require.Error(t, err)
}

func TestShouldCompareEqualAndGreater(t *testing.T) {
	// Arrange
	a := Encode("x", 1)
	b := Encode("x", 1)
	c := Encode("x", 2)
	// Act / Assert
	assert.Equal(t, 0, Compare(a, b))
	assert.Greater(t, Compare(c, b), 0)
}

func TestShouldEncodeToBytesGivenLexKey(t *testing.T) {
	// Arrange
	in := LexKey("ab")
	// Act
	out, err := encodeToBytes(in)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, "6162", LexKey(out).ToHexString())
}

func TestShouldCanonicalizeVariousWidthsViaEncodeCanonicalWidth(t *testing.T) {
	// Arrange
	vals := []any{int(1), int16(-2), uint16(3), uint32(4), float32(1.5)}
	// Act
	k := EncodeCanonicalWidth(vals...)
	// Assert: compare with manually widened equivalents
	want := Encode(int64(1), int64(-2), uint64(3), uint64(4), float64(1.5))
	assert.Equal(t, want, k)
}

func TestShouldMarshalEmptyLexKeyToJSON(t *testing.T) {
	// Arrange
	var e LexKey

	// Act
	data, err := e.MarshalJSON()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []byte(`""`), data)
}

func TestShouldReturnErrorWhenEncodeIntoCanonicalWidthDstTooSmall(t *testing.T) {
	// Arrange
	parts := []any{"t", uint32(1), uint64(2)}
	need := EncodeSizeCanonicalWidth(parts...)
	dst := make([]byte, need-1)

	// Act
	_, err := EncodeIntoCanonicalWidth(dst, parts...)

	// Assert
	require.Error(t, err)
}

func TestShouldReturnErrorWhenEncodeIntoCanonicalWidthUnsupportedPart(t *testing.T) {
	// Arrange
	dst := make([]byte, 64)
	parts := []any{"a", map[int]int{1: 2}}

	// Act
	_, err := EncodeIntoCanonicalWidth(dst, parts...)

	// Assert
	require.Error(t, err)
}

func TestShouldMarshalNonEmptyLexKeyToJSON(t *testing.T) {
	// Arrange
	key := Encode("hello")

	// Act
	data, err := key.MarshalJSON()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []byte(`"68656c6c6f"`), data)
}

func TestEncodeIntoReturnsErrorForUnsupportedTypeDirect(t *testing.T) {
	// Arrange
	dst := make([]byte, 64)

	// Act
	_, err := encodeInto(dst, map[string]int{"a": 1})

	// Assert
	require.Error(t, err)
}

func TestShouldPanicWithExpectedMessageForEncodeWrappers(t *testing.T) {
	// Arrange: obtain the underlying errors first
	_, err1 := NewLexKey(make(chan int))
	require.Error(t, err1)
	expected1 := fmt.Sprintf("failed to encode key with pre-validated parts: %v", err1)

	_, err2 := NewLexKeyCanonicalWidth(make(chan int))
	require.Error(t, err2)
	expected2 := fmt.Sprintf("failed to encode canonical-width key: %v", err2)

	// Act / Assert
	t.Run("Encode wrapper", func(t *testing.T) {
		msg, ok := test.CapturePanicMessage(func() { Encode(make(chan int)) })
		require.True(t, ok)
		assert.Equal(t, expected1, msg)
	})

	t.Run("EncodeCanonicalWidth wrapper", func(t *testing.T) {
		msg, ok := test.CapturePanicMessage(func() { EncodeCanonicalWidth(make(chan int)) })
		require.True(t, ok)
		assert.Equal(t, expected2, msg)
	})
}

// Target encodeInto branches for float infinities and UUID/uint8 boundaries.
func TestEncodeIntoFloat64Infinities(t *testing.T) {
	buf := make([]byte, 16)

	n, err := encodeInto(buf, math.Inf(1))
	require.NoError(t, err)
	require.Equal(t, 8, n)
	got := hex.EncodeToString(buf[:n])
	want := hex.EncodeToString(encodeFloat64(math.Inf(1)))
	require.Equal(t, want, got)

	n, err = encodeInto(buf, math.Inf(-1))
	require.NoError(t, err)
	require.Equal(t, 8, n)
	got = hex.EncodeToString(buf[:n])
	want = hex.EncodeToString(encodeFloat64(math.Inf(-1)))
	require.Equal(t, want, got)
}

func TestEncodeIntoFloat32Infinities(t *testing.T) {
	buf := make([]byte, 8)

	n, err := encodeInto(buf, float32(math.Inf(1)))
	require.NoError(t, err)
	require.Equal(t, 4, n)
	got := hex.EncodeToString(buf[:n])
	want := hex.EncodeToString(encodeFloat32(float32(math.Inf(1))))
	require.Equal(t, want, got)

	n, err = encodeInto(buf, float32(math.Inf(-1)))
	require.NoError(t, err)
	require.Equal(t, 4, n)
	got = hex.EncodeToString(buf[:n])
	want = hex.EncodeToString(encodeFloat32(float32(math.Inf(-1))))
	require.Equal(t, want, got)
}

func TestEncodeIntoFloat64NaN(t *testing.T) {
	buf := make([]byte, 16)
	n, err := encodeInto(buf, float64(math.NaN()))
	require.NoError(t, err)
	require.Equal(t, 8, n)
	// Should match canonical encoding
	got := hex.EncodeToString(buf[:n])
	want := hex.EncodeToString(encodeFloat64(math.NaN()))
	require.Equal(t, want, got)
}

func TestEncodeIntoFloat64Finite(t *testing.T) {
	cases := []struct {
		name string
		val  float64
	}{
		{"positive", 3.14},
		{"negative", -3.14},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := make([]byte, 16)
			n, err := encodeInto(buf, c.val)
			require.NoError(t, err)
			require.Equal(t, 8, n)
			got := hex.EncodeToString(buf[:n])
			want := hex.EncodeToString(encodeFloat64(c.val))
			require.Equal(t, want, got)
		})
	}
}

func TestEncodeIntoUUIDAndUint8Boundaries(t *testing.T) {
	buf := make([]byte, 32)
	u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	n, err := encodeInto(buf, u)
	require.NoError(t, err)
	require.Equal(t, 16, n)
	require.Equal(t, hex.EncodeToString(u[:]), hex.EncodeToString(buf[:n]))

	// uint8 boundaries
	n, err = encodeInto(buf, uint8(0))
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, "00", hex.EncodeToString(buf[:n]))

	n, err = encodeInto(buf, uint8(255))
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, "ff", hex.EncodeToString(buf[:n]))
}

func TestEncodeIntoIntegerWidthsAndTime(t *testing.T) {
	buf := make([]byte, 32)

	// int (positive)
	n, err := encodeInto(buf, int(12345))
	require.NoError(t, err)
	require.Equal(t, 8, n)
	require.Equal(t, hex.EncodeToString(Encode(int64(12345))), hex.EncodeToString(buf[:n]))

	// int (negative)
	n, err = encodeInto(buf, int(-12345))
	require.NoError(t, err)
	require.Equal(t, 8, n)

	// int32
	n, err = encodeInto(buf, int32(0x7fffffff))
	require.NoError(t, err)
	require.Equal(t, 4, n)

	// int16
	n, err = encodeInto(buf, int16(-1234))
	require.NoError(t, err)
	require.Equal(t, 2, n)

	// float NaN already covered via helpers; ensure float32 NaN path in encodeInto is covered
	n, err = encodeInto(buf, float32(float32(math.NaN())))
	require.NoError(t, err)
	require.Equal(t, 4, n)

	// time.Time extreme values
	n, err = encodeInto(buf, time.Unix(0, 0))
	require.NoError(t, err)
	require.Equal(t, 8, n)

	// time.Duration negative
	n, err = encodeInto(buf, time.Duration(-1))
	require.NoError(t, err)
	require.Equal(t, 8, n)
}
