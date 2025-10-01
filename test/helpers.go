package test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// CapturePanicMessage is a test helper shared across tests to capture a panic string.
func CapturePanicMessage(f func()) (msg string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprintf("%v", r)
			ok = true
		}
	}()
	f()
	return "", false
}

// HexEncode returns the lowercase hex encoding of the provided bytes.
// Exported so other packages in tests can reuse it.
func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

// AssertHexEqual compares expected hex string with bytes.
func AssertHexEqual(t *testing.T, expected string, b []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Equal(t, expected, hex.EncodeToString(b), msgAndArgs...)
}

// AssertHexLess asserts hex(a) < hex(b) lexicographically.
func AssertHexLess(t *testing.T, a, b []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Less(t, hex.EncodeToString(a), hex.EncodeToString(b), msgAndArgs...)
}

// AssertHexLessOrEqual asserts hex(a) <= hex(b) lexicographically.
func AssertHexLessOrEqual(t *testing.T, a, b []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assert.LessOrEqual(t, hex.EncodeToString(a), hex.EncodeToString(b), msgAndArgs...)
}

// AssertHexGreater asserts hex(a) > hex(b) lexicographically.
func AssertHexGreater(t *testing.T, a, b []byte, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Greater(t, hex.EncodeToString(a), hex.EncodeToString(b), msgAndArgs...)
}
