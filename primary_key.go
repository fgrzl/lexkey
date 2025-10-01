package lexkey

import (
	"bytes"
	"errors"
)

// NewPrimaryKey creates a new PrimaryKey from partition and row keys.
// Panics if either key is nil.
func NewPrimaryKey(partitionKey, rowKey LexKey) PrimaryKey {
	if partitionKey == nil || rowKey == nil {
		panic("partitionKey and rowKey cannot be nil")
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

// DecodePrimaryKey decodes a PrimaryKey from its byte encoding.
// Returns an error if the separator is missing or input is invalid.
func DecodePrimaryKey(raw []byte) (PrimaryKey, error) {
	sep := bytes.IndexByte(raw, Seperator)
	if sep < 0 {
		return PrimaryKey{}, errors.New("DecodePrimaryKey: missing separator")
	}
	return NewPrimaryKey(
		append([]byte(nil), raw[:sep]...),
		append([]byte(nil), raw[sep+1:]...),
	), nil
}
