package lexkey

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkNewLexKey benchmarks the creation of LexKey with various types
func BenchmarkNewLexKey(b *testing.B) {
	testCases := []struct {
		name  string
		parts []any
	}{
		{"SingleString", []any{"hello"}},
		{"SingleInt", []any{123}},
		{"SingleUUID", []any{uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")}},
		{"SingleFloat64", []any{3.14159}},
		{"SingleTime", []any{time.Now()}},
		{"TwoParts", []any{"user", 123}},
		{"ThreeParts", []any{"tenant", "user", 456}},
		{"FiveParts", []any{"tenant", "table", "user", 789, time.Now()}},
		{"MixedTypes", []any{"prefix", 42, uuid.New(), true, 3.14}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = NewLexKey(tc.parts...)
			}
		})
	}
}

// BenchmarkEncodeIntoPrealloc measures using encodeInto with a freshly allocated buffer each iteration
func BenchmarkEncodeIntoPrealloc(b *testing.B) {
	parts := []any{"tenant", "user", 123, uuid.New(), time.Now(), true, []byte("metadata")}
	size := estimateSize(parts)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, size)
		pos := 0
		for j, p := range parts {
			n, _ := encodeInto(buf[pos:], p)
			pos += n
			if j < len(parts)-1 {
				buf[pos] = Seperator
				pos++
			}
		}
		_ = buf[:pos]
	}
}

// BenchmarkEncodeIntoReuse measures using encodeInto with a single reused buffer
func BenchmarkEncodeIntoReuse(b *testing.B) {
	parts := []any{"tenant", "user", 123, uuid.New(), time.Now(), true, []byte("metadata")}
	size := estimateSize(parts)
	buf := make([]byte, size)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := 0
		for j, p := range parts {
			n, _ := encodeInto(buf[pos:], p)
			pos += n
			if j < len(parts)-1 {
				buf[pos] = Seperator
				pos++
			}
		}
		_ = buf[:pos]
	}
}

// BenchmarkEncodeParallel measures concurrent encoding using per-goroutine buffers
func BenchmarkEncodeParallel(b *testing.B) {
	parts := []any{"tenant", "user", 123, uuid.New(), time.Now(), true, []byte("metadata")}
	size := estimateSize(parts)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		localBuf := make([]byte, size)
		for pb.Next() {
			pos := 0
			for j, p := range parts {
				n, _ := encodeInto(localBuf[pos:], p)
				pos += n
				if j < len(parts)-1 {
					localBuf[pos] = Seperator
					pos++
				}
			}
			_ = localBuf[:pos]
		}
	})
}

// BenchmarkSortLargeKeys benchmarks sorting a large slice of keys
func BenchmarkSortLargeKeys(b *testing.B) {
	// prepare baseline keys
	N := 2000
	keys := make([]LexKey, N)
	for i := 0; i < N; i++ {
		keys[i] = Encode("user", i, uuid.New())
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copyKeys := make([]LexKey, len(keys))
		copy(copyKeys, keys)
		sort.Slice(copyKeys, func(i, j int) bool { return Compare(copyKeys[i], copyKeys[j]) < 0 })
	}
}

// BenchmarkEncode benchmarks the Encode function (no error handling)
func BenchmarkEncode(b *testing.B) {
	testCases := []struct {
		name  string
		parts []any
	}{
		{"SingleString", []any{"hello"}},
		{"SingleInt", []any{123}},
		{"SingleUUID", []any{uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")}},
		{"TwoParts", []any{"user", 123}},
		{"ThreeParts", []any{"tenant", "user", 456}},
		{"MixedTypes", []any{"prefix", 42, uuid.New(), true, 3.14}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Encode(tc.parts...)
			}
		})
	}
}

// BenchmarkEncodeTypes benchmarks encoding of different data types
func BenchmarkEncodeTypes(b *testing.B) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	testTime := time.Now()
	testBytes := []byte("hello world")

	testCases := []struct {
		name string
		part any
	}{
		{"String", "hello world"},
		{"Int", 123456789},
		{"Int64", int64(123456789)},
		{"Int32", int32(123456)},
		{"Int16", int16(12345)},
		{"Uint64", uint64(123456789)},
		{"Uint32", uint32(123456)},
		{"Uint16", uint16(12345)},
		{"Uint8", uint8(123)},
		{"Float64", 3.141592653589793},
		{"Float32", float32(3.14159)},
		{"Bool", true},
		{"UUID", testUUID},
		{"Time", testTime},
		{"Duration", time.Minute},
		{"ByteSlice", testBytes},
		{"Nil", nil},
		{"Struct", struct{}{}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Encode(tc.part)
			}
		})
	}
}

// BenchmarkEncodeFirst benchmarks the EncodeFirst function
func BenchmarkEncodeFirst(b *testing.B) {
	testCases := []struct {
		name  string
		parts []any
	}{
		{"SingleString", []any{"hello"}},
		{"TwoParts", []any{"user", 123}},
		{"ThreeParts", []any{"tenant", "user", 456}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EncodeFirst(tc.parts...)
			}
		})
	}
}

// BenchmarkEncodeLast benchmarks the EncodeLast function
func BenchmarkEncodeLast(b *testing.B) {
	testCases := []struct {
		name  string
		parts []any
	}{
		{"SingleString", []any{"hello"}},
		{"TwoParts", []any{"user", 123}},
		{"ThreeParts", []any{"tenant", "user", 456}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = EncodeLast(tc.parts...)
			}
		})
	}
}

// BenchmarkToHexString benchmarks hex string conversion
func BenchmarkToHexString(b *testing.B) {
	keys := []LexKey{
		Encode("hello"),
		Encode("user", 123),
		Encode("tenant", "user", 456),
		Encode(uuid.New(), time.Now(), 12345),
		Encode("prefix", 42, uuid.New(), true, 3.14, []byte("data")),
	}

	for i, key := range keys {
		b.Run(fmt.Sprintf("Key%d", i+1), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = key.ToHexString()
			}
		})
	}
}

// BenchmarkFromHexString benchmarks hex string parsing
func BenchmarkFromHexString(b *testing.B) {
	hexStrings := []string{
		"68656c6c6f",
		"75736572000000000000000000007b",
		"74656e616e740075736572000000000000000000000001c8",
		"550e8400e29b41d4a716446655440000008000000000000000000000000000007b",
	}

	for i, hexStr := range hexStrings {
		b.Run(fmt.Sprintf("HexString%d", i+1), func(b *testing.B) {
			var key LexKey
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = key.FromHexString(hexStr)
			}
		})
	}
}

// BenchmarkMarshalJSON benchmarks JSON marshaling
func BenchmarkMarshalJSON(b *testing.B) {
	keys := []LexKey{
		Encode("hello"),
		Encode("user", 123),
		Encode("tenant", "user", 456),
		Encode(uuid.New(), time.Now(), 12345),
		Encode("prefix", 42, uuid.New(), true, 3.14, []byte("data")),
	}

	for i, key := range keys {
		b.Run(fmt.Sprintf("Key%d", i+1), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = key.MarshalJSON()
			}
		})
	}
}

// BenchmarkUnmarshalJSON benchmarks JSON unmarshaling
func BenchmarkUnmarshalJSON(b *testing.B) {
	jsonData := [][]byte{
		[]byte(`"68656c6c6f"`),
		[]byte(`"75736572000000000000000000007b"`),
		[]byte(`"74656e616e740075736572000000000000000000000001c8"`),
		[]byte(`null`),
	}

	for i, data := range jsonData {
		b.Run(fmt.Sprintf("JSON%d", i+1), func(b *testing.B) {
			var key LexKey
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = key.UnmarshalJSON(data)
			}
		})
	}
}

// BenchmarkPrimaryKeyEncode benchmarks PrimaryKey encoding
func BenchmarkPrimaryKeyEncode(b *testing.B) {
	testCases := []struct {
		name string
		pk   PrimaryKey
	}{
		{
			"SimpleStrings",
			NewPrimaryKey(Encode("partition1"), Encode("row1")),
		},
		{
			"MixedTypes",
			NewPrimaryKey(Encode("tenant", 123), Encode("user", uuid.New(), time.Now())),
		},
		{
			"LargeKeys",
			NewPrimaryKey(
				Encode("tenant", "table", "partition", 12345),
				Encode("user", uuid.New(), time.Now(), 67890, "metadata"),
			),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tc.pk.Encode()
			}
		})
	}
}

// BenchmarkDecodePrimaryKey benchmarks PrimaryKey decoding
func BenchmarkDecodePrimaryKey(b *testing.B) {
	testKeys := []LexKey{
		NewPrimaryKey(Encode("partition1"), Encode("row1")).Encode(),
		NewPrimaryKey(Encode("tenant", 123), Encode("user", uuid.New())).Encode(),
		NewPrimaryKey(
			Encode("tenant", "table", "partition"),
			Encode("user", uuid.New(), time.Now()),
		).Encode(),
	}

	for i, key := range testKeys {
		b.Run(fmt.Sprintf("Key%d", i+1), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = DecodePrimaryKey(key)
			}
		})
	}
}

// BenchmarkRangeKeyEncode benchmarks RangeKey encoding
func BenchmarkRangeKeyEncode(b *testing.B) {
	partition := Encode("tenant", "table")
	lower := Encode("user1")
	upper := Encode("user9")
	rk := NewRangeKey(partition, lower, upper)

	b.Run("WithPartition", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rk.Encode(true)
		}
	})

	b.Run("WithoutPartition", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rk.Encode(false)
		}
	})
}

// BenchmarkNewRangeKeyFull benchmarks full range key creation
func BenchmarkNewRangeKeyFull(b *testing.B) {
	partition := Encode("tenant", "table", 123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRangeKeyFull(partition)
	}
}

// BenchmarkMemoryAllocations tests memory allocations for various operations
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("SmallKey", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Encode("hello")
		}
	})

	b.Run("MediumKey", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Encode("tenant", "user", 123, uuid.New())
		}
	})

	b.Run("LargeKey", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Encode("tenant", "table", "user", 123, uuid.New(), time.Now(), true, 3.14, []byte("metadata"))
		}
	})

	b.Run("PrimaryKeyEncode", func(b *testing.B) {
		pk := NewPrimaryKey(Encode("partition"), Encode("row"))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pk.Encode()
		}
	})

	b.Run("HexStringConversion", func(b *testing.B) {
		key := Encode("tenant", "user", 123)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = key.ToHexString()
		}
	})
}

// BenchmarkLexicographicOrdering tests sorting performance
func BenchmarkLexicographicOrdering(b *testing.B) {
	keys := make([]LexKey, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = Encode("user", i, uuid.New())
	}

	b.Run("KeyComparison", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < len(keys)-1; j++ {
				_ = string(keys[j]) < string(keys[j+1])
			}
		}
	})
}

// BenchmarkFloatSpecialValues benchmarks encoding of special float values
func BenchmarkFloatSpecialValues(b *testing.B) {
	specialValues := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		0.0,
		math.Copysign(0.0, -1), // Negative zero
		math.SmallestNonzeroFloat64,
		math.MaxFloat64,
	}

	for i, val := range specialValues {
		b.Run(fmt.Sprintf("Float64_%d", i), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Encode(val)
			}
		})
	}
}

// BenchmarkLongStrings benchmarks encoding of strings of various lengths
func BenchmarkLongStrings(b *testing.B) {
	strings := []string{
		"short",
		"this is a medium length string for testing purposes",
		func() string {
			s := make([]byte, 1024)
			for i := range s {
				s[i] = byte('a' + (i % 26))
			}
			return string(s)
		}(),
		func() string {
			s := make([]byte, 10240)
			for i := range s {
				s[i] = byte('a' + (i % 26))
			}
			return string(s)
		}(),
	}

	for _, str := range strings {
		b.Run(fmt.Sprintf("String%dBytes", len(str)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Encode(str)
			}
		})
	}
}
