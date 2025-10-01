package lexkey

// NewRangeKey creates a RangeKey for a given partition and row key range.
// Panics if the partition key, lower, or upper key is nil.
func NewRangeKey(partition, lower, upper LexKey) RangeKey {
	if partition == nil {
		panic("partition key cannot be nil")
	}
	if lower == nil {
		panic("lower key cannot be nil")
	}
	if upper == nil {
		panic("upper key cannot be nil")
	}
	return RangeKey{
		PartitionKey: partition,
		StartRowKey:  lower,
		EndRowKey:    upper,
	}
}

// NewRangeKeyFull creates a RangeKey spanning the full partition.
// Panics if the partition key is nil.
func NewRangeKeyFull(partition LexKey) RangeKey {
	if partition == nil {
		panic("partition key cannot be nil")
	}
	return RangeKey{
		PartitionKey: partition,
		StartRowKey:  Empty,
		EndRowKey:    Last,
	}
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
		size += 1 + len(rowKey) // Seperator + rowKey
		if isUpper {
			size++ // extra byte for end marker
		}
	} else {
		size += 1 // Just Seperator or end marker
	}
	result := make(LexKey, size)

	n := 0
	if withPartitionKey {
		n += copy(result, partitionKey)
	}
	if len(rowKey) == 0 {
		result[n] = ternary(isUpper, EndMarker, Seperator)
	} else {
		result[n] = Seperator
		n++
		copy(result[n:], rowKey)
		if isUpper {
			result[len(result)-1] = EndMarker
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
