# Getting started

## Install

```bash
go get github.com/fgrzl/lexkey
```

## Encode a key

```go
key := lexkey.Encode("user", int64(123), true)
fmt.Println(key.ToHexString())
```

## Compare keys

```go
k1 := lexkey.Encode("apple")
k2 := lexkey.Encode("banana")
ordered := bytes.Compare(k1, k2) < 0 // true
```

## Primary keys for KV

```go
pk := lexkey.NewPrimaryKey(
    lexkey.Encode("tenant", tenantID), // partition
    lexkey.Encode("user", userID),     // row
)
```

## Prefix range scan

```go
prefix := lexkey.Encode("orders", shopID)
from := lexkey.EncodeFirst(prefix)
to := lexkey.EncodeLast(prefix)
// use from/to with kv.Query range operators
```

## JSON

```go
data, _ := json.Marshal(key) // hex string form
```

See [SPEC.md](../SPEC.md) for encoding details.
