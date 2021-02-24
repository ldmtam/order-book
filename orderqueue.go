package orderbook

import (
	"fmt"
)

// OrderQueue ...
type OrderQueue struct {
	_padding0 [8]uint64
	r         uint64 // next position to read
	_padding1 [8]uint64
	w         uint64 // next position to write
	_padding2 [8]uint64
	len       uint64
	_padding3 [8]uint64
	cap       uint64
	_padding4 [8]uint64
	mask      uint64
	_padding5 [8]uint64
	contents  []*Order
	_padding6 [8]uint64
}

// NewOrderQueue ...
func NewOrderQueue(size uint64) *OrderQueue {
	return &OrderQueue{
		cap:      size,
		mask:     size - 1,
		contents: make([]*Order, int(size)),
	}
}

// Put ...
func (q *OrderQueue) Put(value *Order) bool {
	if q.len == q.cap || (q.w == q.r && q.len > 0) {
		return false
	}
	q.contents[q.w&q.mask] = value
	q.w++
	q.len++
	return true
}

// Get ...
func (q *OrderQueue) Get() *Order {
	if q.r == q.w {
		return nil
	}
	content := q.contents[q.r&q.mask]
	q.r++
	q.len--
	return content
}

// Reset ...
func (q *OrderQueue) Reset() {
	q.len = 0
	q.r = 0
	q.w = 0
}

// Len ...
func (q *OrderQueue) Len() int {
	return int(q.len)
}

func (q *OrderQueue) String() string {
	return fmt.Sprintf("Read: %d, Write: %d", q.r, q.w)
}
