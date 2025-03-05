package lexkey

func NewRangeKey(partition, lower, upper LexKey) RangeKey {
	return RangeKey{
		PartitionKey: partition,
		StartRowKey:  lower,
		EndRowKey:    upper,
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
