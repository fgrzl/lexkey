[![ci](https://github.com/fgrzl/lexkey/actions/workflows/ci.yml/badge.svg)](https://github.com/fgrzl/lexkey/actions/workflows/ci.yml)
[![Dependabot Updates](https://github.com/fgrzl/lexkey/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/fgrzl/lexkey/actions/workflows/dependabot/dependabot-updates)

# lexkey

**`lexkey`** is a lightweight **lexicographically sortable key encoding library** for Go.  
It provides **consistent, ordered, and efficient** encoding for various data types, ensuring they sort **correctly** when stored in databases, key-value stores, or other ordered storage systems.



## ✨ **Features**
- 🚀 **Lexicographically sortable** encoding for structured keys.
- 🔑 Supports **strings, integers, floats, UUIDs, booleans, timestamps, durations, and byte slices**.
- 🔄 **Consistent ordering** for mixed types like `int32` and `int64`.
- 📦 **Optimized encoding** for space efficiency and speed.
- 📡 **JSON serialization support** for interoperability.



## 📦 **Installation**
```sh
go get github.com/fgrzl/lexkey
```



## 🛠 **Usage**
### **Create a LexKey**
```go
package main

import (
	"fmt"
	"github.com/fgrzl/lexkey"
)

func main() {
	key, err := lexkey.Encode("user", 123, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Encoded Key (Hex):", key.ToHexString())
}
```

### **Sorting Keys**
```go
key1, _ := lexkey.Encode("apple")
key2, _ := lexkey.Encode("banana")

fmt.Println(string(key1) < string(key2)) // ✅ True (correct lexicographic order)
```

### **Handling Numbers**
```go
key1, _ := lexkey.Encode(int64(-100))
key2, _ := lexkey.Encode(int64(50))

fmt.Println(string(key1) < string(key2)) // ✅ True (correct sorting for signed integers)
```

### **Using UUIDs**
```go
import "github.com/google/uuid"

id := uuid.New()
key, _ := lexkey.Encode("order", id)

fmt.Println("Encoded UUID Key:", key.ToHexString())
```

### **LexKey JSON Serialization**
```go
import "encoding/json"

key, _ := lexkey.Encode("session", 42)
jsonData, _ := json.Marshal(key)
fmt.Println(string(jsonData)) // ✅ Encoded as a hex string
```



## 🔍 Supported Data Types

| Type            | Supported? | Encoding Details |
|----------------|-----------|------------------|
| `string`       | ✅ Yes    | Stored as raw UTF-8 bytes |
| `int32`        | ✅ Yes    | Converted to `int64` for uniform sorting |
| `int64`        | ✅ Yes    | Sign-bit flipped for correct ordering |
| `uint32`       | ✅ Yes    | Big-endian encoded |
| `uint64`       | ✅ Yes    | Big-endian encoded |
| `float32`      | ✅ Yes    | IEEE 754 encoded with sign-bit transformation |
| `float64`      | ✅ Yes    | IEEE 754 encoded with sign-bit transformation |
| `bool`         | ✅ Yes    | `true → 0x01`, `false → 0x00` |
| `uuid.UUID`    | ✅ Yes    | 16-byte raw representation |
| `[]byte`       | ✅ Yes    | Stored as-is |
| `time.Time`    | ✅ Yes    | Encoded as `int64` nanoseconds since Unix epoch |
| `time.Duration`| ✅ Yes    | Encoded as `int64` nanoseconds |


## 📌 **Key Functions**
### **Encoding Keys**
```go
func Encode(parts ...any) (LexKey, error)
```
Encodes multiple values into a **single lexicographically sortable** key.

### **Sorting Helpers**
```go
func (e LexKey) EncodeFirst() []byte // Appends a NULL byte for range queries
func (e LexKey) EncodeLast() []byte  // Appends a MAX byte for range queries
```

### **Hex Encoding**
```go
func (e LexKey) ToHexString() string
func (e *LexKey) FromHexString(hexStr string) error
```

### **JSON Serialization**
```go
func (e LexKey) MarshalJSON() ([]byte, error)
func (e *LexKey) UnmarshalJSON(data []byte) error
```



## 🏆 **Why Use `lexkey`?**
✅ **Fast & Efficient** → Uses compact, binary-safe encoding.  
✅ **Correct Ordering** → Works across all supported types.  
✅ **Minimal Dependencies** → Only `uuid` and standard Go packages.  



## 🛠 **Testing**
Run the full test suite:
```sh
go test -cover ./...
```
**Test Coverage:** ✅ **100%** 🎯



## 📜 **License**
This project is licensed under the **MIT License**.



## 💡 **Contributing**
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature-name`).
3. Make changes and run tests (`go test -cover ./...`).
4. Open a pull request!

