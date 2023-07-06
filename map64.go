package intintmap

import (
	"math"
)

// Val64 is a type constraint for values that are guaranteed to be 64 bits wide.
type Val64 interface {
	~int64 | ~uint64
}

type Pair64[K Val64, V any] struct {
	K K
	V V
}

const fillFactor64 = 0.7

func phiMix64(x int) int {
	h := x * 0x9E3779B9
	return h ^ (h >> 16)
}

// Map64 is a map-like data-structure for 64-bit types
type Map64[K Val64, V any] struct {
	data []Pair64[K, V] // key-value pairs
	size int

	zeroVal    V    // value of 'zero' key
	hasZeroKey bool // do we have 'zero' key in the map?
}

// New64 ...
func New64[K, V Val64](capacity int) *Map64[K, V] {
	return &Map64[K, V]{
		data: make([]Pair64[K, V], arraySize(capacity, fillFactor64)),
	}
}

// Get returns the value if the key is found.
func (m *Map64[K, V]) Get(key K) (V, bool) {
	if key == K(0) {
		if m.hasZeroKey {
			return m.zeroVal, true
		}
		var zero V
		return zero, false
	}

	idx := m.startIndex(key)
	pair := m.data[idx]

	if pair.K == K(0) { // end of chain already
		var zero V
		return zero, false
	}
	if pair.K == key { // we check zero prior to this call
		return pair.V, true
	}

	for {
		idx = m.nextIndex(idx)
		pair = m.data[idx]
		if pair.K == K(0) {
			var zero V
			return zero, false
		}
		if pair.K == key {
			return pair.V, true
		}
	}
}

// Put adds or updates key with value val.
func (m *Map64[K, V]) Put(key K, val V) {
	if key == K(0) {
		if !m.hasZeroKey {
			m.size++
		}
		m.zeroVal = val
		m.hasZeroKey = true
		return
	}

	idx := m.startIndex(key)
	pair := &m.data[idx]

	if pair.K == K(0) { // end of chain already
		pair.K = key
		pair.V = val
		if m.size >= m.sizeThreshold() {
			m.rehash()
		} else {
			m.size++
		}
		return
	} else if pair.K == key { // overwrite existing value
		pair.V = val
		return
	}

	// hash collision, seek next empty or key match
	for {
		idx = m.nextIndex(idx)
		pair = &m.data[idx]

		if pair.K == K(0) {
			pair.K = key
			pair.V = val
			if m.size >= m.sizeThreshold() {
				m.rehash()
			} else {
				m.size++
			}
			return
		} else if pair.K == key {
			pair.V = val
			return
		}
	}
}

func (m *Map64[K, V]) rehash() {
	oldData := m.data
	m.data = make([]Pair64[K, V], 2*len(m.data))

	// reset size
	if m.hasZeroKey {
		m.size = 1
	} else {
		m.size = 0
	}

	for _, p := range oldData {
		if p.K != K(0) {
			m.Put(p.K, p.V)
		}
	}
}

// Len returns the number of elements in the map.
func (m *Map64[K, V]) Len() int {
	return m.size
}

func (m *Map64[K, V]) sizeThreshold() int {
	return int(math.Floor(float64(len(m.data)) * fillFactor64))
}

func (m *Map64[K, V]) startIndex(key K) int {
	return phiMix64(int(key)) & (len(m.data) - 1)
}

func (m *Map64[K, V]) nextIndex(idx int) int {
	return (idx + 1) & (len(m.data) - 1)
}
