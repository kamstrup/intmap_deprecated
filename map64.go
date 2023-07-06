package intintmap

import (
	"fmt"
	"math"
)

// Val64 is a type constraint for values that are guaranteed to be 64 bits wide.
type Val64 interface {
	~int64 | ~uint64
}

const fillFactor64 = 0.7

func phiMix64(x int) int {
	h := x * 0x9E3779B9
	return h ^ (h >> 16)
}

// Map64 is a map-like data-structure for 64-bit types
type Map64[T Val64] struct {
	data [][2]T // key-value pairs
	size int

	zeroVal    T    // value of 'zero' key
	hasZeroKey bool // do we have 'zero' key in the map?
}

// New64 ...
func New64[T Val64](capacity int) *Map64[T] {
	return &Map64[T]{
		data: make([][2]T, arraySize(capacity, fillFactor64)),
	}
}

// Get returns the value if the key is found.
func (m *Map64[T]) Get(key T) (T, bool) {
	if key == T(0) {
		if m.hasZeroKey {
			return m.zeroVal, true
		}
		return 0, false
	}

	idx := m.startIndex(key)
	pair := m.data[idx]

	if pair[0] == T(0) { // end of chain already
		return 0, false
	}
	if pair[0] == key { // we check zero prior to this call
		return pair[1], true
	}

	for {
		idx = m.nextIndex(idx)
		pair = m.data[idx]
		if pair[0] == T(0) {
			return 0, false
		}
		if pair[0] == key {
			return pair[1], true
		}
	}
}

// Put adds or updates key with value val.
func (m *Map64[T]) Put(key, val T) {
	fmt.Println("SZTH", m.size, m.sizeThreshold())
	if key == T(0) {
		if !m.hasZeroKey {
			m.size++
		}
		m.zeroVal = val
		m.hasZeroKey = true
		return
	}

	idx := m.startIndex(key)
	pair := &m.data[idx]

	if pair[0] == T(0) { // end of chain already
		pair[0] = key
		pair[1] = val
		if m.size >= m.sizeThreshold() {
			m.rehash()
		} else {
			m.size++
		}
		return
	} else if pair[0] == key { // overwrite existing value
		pair[1] = val
		return
	}

	// hash collision, seek next empty or key match
	for {
		idx = m.nextIndex(idx)
		pair = &m.data[idx]

		if pair[0] == T(0) {
			pair[0] = key
			pair[1] = val
			if m.size >= m.sizeThreshold() {
				m.rehash()
			} else {
				m.size++
			}
			return
		} else if pair[0] == key {
			pair[1] = val
			return
		}
	}

}

func (m *Map64[T]) rehash() {
	oldData := m.data
	m.data = make([][2]T, 2*len(m.data))

	// reset size
	if m.hasZeroKey {
		m.size = 1
	} else {
		m.size = 0
	}

	for _, p := range oldData {
		if p[0] != T(0) {
			m.Put(p[0], p[1])
		}
	}
}

// Len returns the number of elements in the map.
func (m *Map64[T]) Len() int {
	return m.size
}

func (m *Map64[T]) sizeThreshold() int {
	return int(math.Floor(float64(len(m.data)) * fillFactor64))
}

func (m *Map64[T]) startIndex(key T) int {
	return phiMix64(int(key)) & (len(m.data) - 1)
}

func (m *Map64[T]) nextIndex(idx int) int {
	return (idx + 1) & (len(m.data) - 1)
}
