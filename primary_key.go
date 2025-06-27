package lexkey

import (
	"bytes"
	"log/slog"
)

// NewPrimaryKey creates a new PrimaryKey from partition and row keys.
// Returns an error if either key is nil.
func NewPrimaryKey(partitionKey, rowKey LexKey) PrimaryKey {
	if partitionKey == nil || rowKey == nil {
		slog.Error("partitionKey and rowKey cannot be nil")
		return PrimaryKey{}
	}
	return PrimaryKey{
		PartitionKey: partitionKey,
		RowKey:       rowKey,
	}
}

// PrimaryKey represents a composite key for key-value storage.
type PrimaryKey struct {
	PartitionKey LexKey
	RowKey       LexKey
}

// Encode concatenates PartitionKey and RowKey with a Seperator.
func (pk PrimaryKey) Encode() LexKey {
	result := make(LexKey, len(pk.PartitionKey)+len(pk.RowKey)+1)
	n := copy(result, pk.PartitionKey)
	result[n] = Seperator
	copy(result[n+1:], pk.RowKey)
	return result
}

func DecodePrimaryKey(raw []byte) PrimaryKey {
	sep := bytes.IndexByte(raw, Seperator)
	if sep < 0 {
		// Invalid format: no separator found
		slog.Warn("DecodePrimaryKey: missing separator", "raw", raw)
		return PrimaryKey{}
	}

	return NewPrimaryKey(
		append([]byte(nil), raw[:sep]...),
		append([]byte(nil), raw[sep+1:]...),
	)
}
