package lexkey

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldPanicWhenNewRangeKeyReceivesNilArgs(t *testing.T) {
	tests := []struct {
		name      string
		partition LexKey
		lower     LexKey
		upper     LexKey
	}{
		{"nil partition", nil, Encode("a"), Encode("b")},
		{"nil lower", Encode("p"), nil, Encode("b")},
		{"nil upper", Encode("p"), Encode("a"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act / Assert
			require.Panics(t, func() { NewRangeKey(tt.partition, tt.lower, tt.upper) })
		})
	}
}

func TestShouldCreateFullRangeGivenPartition(t *testing.T) {
	// Arrange / Act
	rk := NewRangeKeyFull(Encode("tenant"))

	// Assert
	assert.Equal(t, Empty, rk.StartRowKey)
	assert.Equal(t, Last, rk.EndRowKey)
}

func TestShouldEncodeRangeKeyWithPartitionKey(t *testing.T) {
	// Arrange
	rk := NewRangeKey(Encode("part"), Encode("start"), Encode("end"))

	// Act
	lower, upper := rk.Encode(true)

	// Assert
	assert.Equal(t, "70617274007374617274", hex.EncodeToString(lower))
	assert.Equal(t, "7061727400656e64ff", hex.EncodeToString(upper))
}

func TestShouldEncodeRangeKeyWithoutPartitionKey(t *testing.T) {
	// Arrange
	rk := NewRangeKey(Encode("p"), Encode("r"), Encode("r"))

	// Act
	lower, upper := rk.Encode(false)

	// Assert
	assert.Equal(t, "0072", hex.EncodeToString(lower))
	assert.Equal(t, "0072ff", hex.EncodeToString(upper))
}

func TestShouldEncodeEmptyRowKeyBoundariesCorrectly(t *testing.T) {
	// Arrange / Act
	lower := encodeBoundary(Encode("partition"), nil, false, true)
	upper := encodeBoundary(Encode("partition"), nil, true, true)

	// Assert
	assert.Equal(t, "706172746974696f6e00", hex.EncodeToString(lower))
	assert.Equal(t, "706172746974696f6eff", hex.EncodeToString(upper))
}
