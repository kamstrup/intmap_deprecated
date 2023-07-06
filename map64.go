package intintmap

import (
	"math"
)

// IntKey is a type constraint for values that can be used as keys in Map64
type IntKey interface {
	~int | ~uint | ~int64 | ~uint64 | ~int32 | ~uint32 | ~int16 | ~uint16 | ~int8 | ~uint8 | ~uintptr
}

type pair[K IntKey, V any] struct {
	K K
	V V
}

const fillFactor64 = 0.7

func phiMix64(x int) int {
	h := x * 0x9E3779B9
	return h ^ (h >> 16)
}

// Map64 is a hashmap where the keys are some any integer type.
type Map64[K IntKey, V any] struct {
	data []pair[K, V] // key-value pairs
	size int

	zeroVal    V    // value of 'zero' key
	hasZeroKey bool // do we have 'zero' key in the map?
}

// New64 creates a new map with keys being any integer subtype.
// The map can store up to the given capacity before reallocation and rehashing occurs.
func New64[K IntKey, V any](capacity int) *Map64[K, V] {
	return &Map64[K, V]{
		data: make([]pair[K, V], arraySize(capacity, fillFactor64)),
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
	p := m.data[idx]

	if p.K == K(0) { // end of chain already
		var zero V
		return zero, false
	}
	if p.K == key { // we check zero prior to this call
		return p.V, true
	}

	// hash collision, seek next hash match, bailing on first empty
	for {
		idx = m.nextIndex(idx)
		p = m.data[idx]
		if p.K == K(0) {
			var zero V
			return zero, false
		}
		if p.K == key {
			return p.V, true
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
	p := &m.data[idx]

	if p.K == K(0) { // end of chain already
		p.K = key
		p.V = val
		if m.size >= m.sizeThreshold() {
			m.rehash()
		} else {
			m.size++
		}
		return
	} else if p.K == key { // overwrite existing value
		p.V = val
		return
	}

	// hash collision, seek next empty or key match
	for {
		idx = m.nextIndex(idx)
		p = &m.data[idx]

		if p.K == K(0) {
			p.K = key
			p.V = val
			if m.size >= m.sizeThreshold() {
				m.rehash()
			} else {
				m.size++
			}
			return
		} else if p.K == key {
			p.V = val
			return
		}
	}
}

func (m *Map64[K, V]) ForEach(f func(K, V)) {
	if m.hasZeroKey {
		f(K(0), m.zeroVal)
	}
	forEach64(m.data, f)
}

// Clear removes all items from the map, but keeps the internal buffers for reuse.
func (m *Map64[K, V]) Clear() {
	var zero V
	m.hasZeroKey = false
	m.zeroVal = zero

	// compiles down to runtime.memclr()
	for i := range m.data {
		m.data[i] = pair[K, V]{}
	}

	m.size = 0
}

func (m *Map64[K, V]) rehash() {
	oldData := m.data
	m.data = make([]pair[K, V], 2*len(m.data))

	// reset size
	if m.hasZeroKey {
		m.size = 1
	} else {
		m.size = 0
	}

	forEach64(oldData, m.Put)
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

func forEach64[K IntKey, V any](pairs []pair[K, V], f func(k K, v V)) {
	for _, p := range pairs {
		if p.K != K(0) {
			f(p.K, p.V)
		}
	}
}

// Del deletes a key and its value, returning true iff the key was found
func (m *Map64[K, V]) Del(key K) bool {
	if key == K(0) {
		if m.hasZeroKey {
			m.hasZeroKey = false
			m.size--
			return true
		}
		return false
	}

	idx := m.startIndex(key)
	p := m.data[idx]

	if p.K == key {
		// any keys that were pushed back needs to be shifted nack into the empty slot
		// to avoid breaking the chain
		m.shiftKeys(idx)
		m.size--
		return true
	} else if p.K == K(0) { // end of chain already
		return false
	}

	for {
		idx = m.nextIndex(idx)
		p = m.data[idx]

		if p.K == key {
			// any keys that were pushed back needs to be shifted nack into the empty slot
			// to avoid breaking the chain
			m.shiftKeys(idx)
			m.size--
			return true
		} else if p.K == K(0) {
			return false
		}

	}
}

func (m *Map64[K, V]) shiftKeys(idx int) int {
	// Shift entries with the same hash.
	// We need to do this on deletion to ensure we don't have zeroes in the hash chain
	for {
		var p pair[K, V]
		lastIdx := idx
		idx = m.nextIndex(idx)
		for {
			p = m.data[idx]
			if p.K == K(0) {
				m.data[lastIdx] = pair[K, V]{}
				return lastIdx
			}

			slot := m.startIndex(p.K)
			if lastIdx <= idx {
				if lastIdx >= slot || slot > idx {
					break
				}
			} else {
				if lastIdx >= slot && slot > idx {
					break
				}
			}
			idx = m.nextIndex(idx)
		}
		m.data[lastIdx] = p
	}
}
