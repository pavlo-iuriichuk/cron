package internal

import "errors"

type Int interface {
	AsInt() int
}

type HeapItems []Int

func (h HeapItems) Less(i, j int) bool {
	return (h[i]).AsInt() < (h[j]).AsInt()
}

type Heap struct {
	size     int
	capacity int
	items    HeapItems
}

func NewHeap(size int) *Heap {
	return &Heap{size: 0, capacity: size, items: make(HeapItems, size, size)}
}

func (h *Heap) Len() int {
	return h.size
}

func (h *Heap) Push(item Int) error {
	if h.size == h.capacity {
		return errors.New("heap is full")
	}
	h.size++
	index := h.size - 1
	h.items[index] = item
	h.bubbleUp(index)
	return nil
}

func (h *Heap) parent(index int) int {
	return (index - 1) / 2
}

func (h *Heap) left(index int) int {
	return 2*index + 1
}

func (h *Heap) right(index int) int {
	return 2*index + 2
}

func (h *Heap) swap(i1, i2 int) {
	h.items[i1], h.items[i2] = h.items[i2], h.items[i1]
}

func (h *Heap) Top() (Int, error) {
	if h.size < 0 {
		return nil, errors.New("heap is empty")
	}
	return h.items[0], nil
}

func (h *Heap) Pop() (Int, error) {
	if h.size < 0 {
		return nil, errors.New("heap is empty")
	}
	if h.size == 0 {
		h.size--
		return h.items[0], nil
	}
	root := h.items[0]
	h.items[0] = h.items[h.size-1]
	h.size--
	h.bubbleDown(0)
	return root, nil
}

func (h *Heap) bubbleUp(index int) {
	for index != 0 && h.items.Less(index, h.parent(index)) {
		h.swap(index, h.parent(index))
		index = h.parent(index)
	}
}

func (h *Heap) bubbleDown(index int) {
	l := h.left(index)
	r := h.right(index)
	smallest := index
	if l < h.size && h.items.Less(l, index) {
		smallest = l
	}
	if r < h.size && h.items.Less(r, smallest) {
		smallest = r
	}
	if smallest != index {
		h.swap(index, smallest)
		h.bubbleDown(smallest)
	}
}
