Fast int64 -> int64 hash in golang.

[![GoDoc](https://godoc.org/github.com/brentp/intintmap?status.svg)](https://godoc.org/github.com/brentp/intintmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/brentp/intintmap)](https://goreportcard.com/report/github.com/brentp/intintmap)

# intintmap

    import "github.com/brentp/intintmap"

Package intintmap is a fast int64 key -> int64 value map.

It is copied nearly verbatim from
http://java-performance.info/implementing-world-fastest-java-int-to-int-hash-map/ .

It interleaves keys and values in the same underlying array to improve locality.

It is up to 3X faster than the builtin map:
```
name                             time/op
Map64Fill-8                       201ms ± 5%
IntIntMapFill-8                   207ms ±31%
StdMapFill-8                      371ms ±11%
Map64Get10PercentHitRate-8        148µs ±40%
IntIntMapGet10PercentHitRate-8    171µs ±50%
StdMapGet10PercentHitRate-8       171µs ±33%
Map64Get100PercentHitRate-8      4.50ms ± 5%
IntIntMapGet100PercentHitRate-8  4.82ms ± 6%
StdMapGet100PercentHitRate-8     15.5ms ±32%
```

## Usage

```go
m := intintmap.New(32768, 0.6)
m.Put(int64(1234), int64(-222))
m.Put(int64(123), int64(33))

v, ok := m.Get(int64(222))
v, ok := m.Get(int64(333))

m.Del(int64(222))
m.Del(int64(333))

fmt.Println(m.Size())

for k := range m.Keys() {
    fmt.Printf("key: %d\n", k)
}

for kv := range m.Items() {
    fmt.Printf("key: %d, value: %d\n", kv[0], kv[1])
}
```

#### type Map

```go
type Map struct {
}
```

Map is a map-like data-structure for int64s

#### func  New

```go
func New(size int, fillFactor float64) *Map
```
New returns a map initialized with n spaces and uses the stated fillFactor. The
map will grow as needed.

#### func (*Map) Get

```go
func (m *Map) Get(key int64) (int64, bool)
```
Get returns the value if the key is found.

#### func (*Map) Put

```go
func (m *Map) Put(key int64, val int64)
```
Put adds or updates key with value val.

#### func (*Map) Del

```go
func (m *Map) Del(key int64)
```
Del deletes a key and its value.

#### func (*Map) Keys

```go
func (m *Map) Keys() chan int64
```
Keys returns a channel for iterating all keys.

#### func (*Map) Items

```go
func (m *Map) Items() chan [2]int64
```
Items returns a channel for iterating all key-value pairs.


#### func (*Map) Size

```go
func (m *Map) Size() int
```
Size returns size of the map.
