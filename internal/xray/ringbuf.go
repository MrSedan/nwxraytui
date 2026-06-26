package xray

import "sync"

type RingBuffer struct {
	mu   sync.RWMutex
	data []string
	head int
	size int
	cap  int
}

func NewRingBuffer(cap int) *RingBuffer {
	return &RingBuffer{data: make([]string, cap), cap: cap}
}

func (rb *RingBuffer) Push(s string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.data[rb.head] = s
	rb.head = (rb.head + 1) % rb.cap
	if rb.size < rb.cap {
		rb.size++
	}
}

func (rb *RingBuffer) Lines() []string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if rb.size == 0 {
		return nil
	}
	out := make([]string, rb.size)
	start := (rb.head - rb.size + rb.cap) % rb.cap
	for i := 0; i < rb.size; i++ {
		out[i] = rb.data[(start+i)%rb.cap]
	}
	return out
}
