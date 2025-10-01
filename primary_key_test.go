package lexkey

import (
	"testing"

	"github.com/fgrzl/lexkey/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldPanicWhenNewPrimaryKeyReceivesNilParts(t *testing.T) {
	tests := []struct {
		name  string
		part1 LexKey
		part2 LexKey
	}{
		{"nil partition", nil, Encode("r")},
		{"nil row", Encode("p"), nil},
		{"both nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act / Assert
			require.Panics(t, func() { NewPrimaryKey(tt.part1, tt.part2) })
		})
	}
}

func TestShouldEncodePrimaryKeyAndDecodeRoundTrip(t *testing.T) {
	// Arrange
	pk := NewPrimaryKey(Encode("partition"), Encode("row"))

	// Act
	enc := pk.Encode()

	// Assert: expected encoding is partition + Seperator + row
	test.AssertHexEqual(t, "706172746974696f6e00726f77", enc)

	// Act
	decoded, err := DecodePrimaryKey(enc)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, pk.PartitionKey, decoded.PartitionKey)
	assert.Equal(t, pk.RowKey, decoded.RowKey)
}

func TestShouldErrorWhenDecodingPrimaryKeyWithNoSeparator(t *testing.T) {
	// Act
	_, err := DecodePrimaryKey([]byte("no_separator"))

	// Assert
	require.Error(t, err)
}

func TestShouldReturnIndependentCopiesWhenDecodingPrimaryKey(t *testing.T) {
	// Arrange
	origP := Encode("p")
	origR := Encode("r")
	pk := NewPrimaryKey(origP, origR)
	enc := pk.Encode()

	// Act
	decoded, err := DecodePrimaryKey(enc)
	require.NoError(t, err)

	// mutate decoded slices
	decoded.PartitionKey[0] ^= 0xff
	decoded.RowKey[0] ^= 0xff

	// Assert originals unchanged
	assert.Equal(t, Encode("p"), origP)
	assert.Equal(t, Encode("r"), origR)
}

func TestShouldPanicWithExpectedMessageForNewPrimaryKey(t *testing.T) {
	// Table-driven cases for different nil combinations
	cases := []struct {
		name     string
		part1    LexKey
		part2    LexKey
		expected string
	}{
		{"nil partition", nil, Encode("r"), "partitionKey and rowKey cannot be nil"},
		{"nil row", Encode("p"), nil, "partitionKey and rowKey cannot be nil"},
		{"both nil", nil, nil, "partitionKey and rowKey cannot be nil"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Act
			msg, ok := test.CapturePanicMessage(func() { NewPrimaryKey(c.part1, c.part2) })

			// Assert
			require.True(t, ok)
			assert.Equal(t, c.expected, msg)
		})
	}
}
