package set

import (
	"errors"
	"math/bits"
)

const (
	BitsPerItem = bits.UintSize
)

func genBitMask(count int) uint {
	var result uint
	for i := 0; i < count; i++ {
		result <<= 1
		result |= 1
	}
	return result
}

var (
	BitsPerIndex = bits.Len(uint(BitsPerItem - 1))
	FullItemMask = genBitMask(BitsPerItem)
	IndexMask    = genBitMask(BitsPerIndex)
)

type UIntBitSet struct {
	items []uint
	max   uint
	size  uint
}

type UIntSet interface {
	Add(val uint)
	Delete(val uint)
	Has(val uint) bool
	Cap() uint
	FirstAbsent(from uint) (uint, error)
	Iterator() <-chan uint
}

func NewBitSet(max uint) *UIntBitSet {
	return &UIntBitSet{
		items: make([]uint, (max/BitsPerItem)+1, (max/BitsPerItem)+1),
		max:   max,
	}
}

func (bs *UIntBitSet) Cap() uint {
	return bs.size
}

func (bs *UIntBitSet) Add(val uint) {
	if bs.Has(val) {
		return
	}
	bs.items[val>>BitsPerIndex] |= 1 << (val & IndexMask)
	bs.size++
}

func (bs *UIntBitSet) Delete(val uint) {
	if bs.Has(val) {
		bs.items[val>>BitsPerIndex] &^= 1 << (val & IndexMask)
		bs.size--
	}
}

func (bs *UIntBitSet) Has(val uint) bool {
	return bs.items[val>>BitsPerIndex]&(1<<(val&IndexMask)) > 0
}

func (bs *UIntBitSet) Iterator() <-chan uint {
	result := make(chan uint)
	go func() {
		for i := uint(0); i < bs.max; i++ {
			if bs.Has(i) {
				result <- i
			}
		}
		close(result)
	}()
	return result
}

func (bs *UIntBitSet) FirstAbsent(from uint) (uint, error) {
	for i := int(from); i < len(bs.items); i++ {
		pos, found := bs.scanItemBitScanForward(i)
		if found {
			return uint(i)<<BitsPerIndex + pos, nil
		}
	}
	return 0, errors.New("unable to find absent item")
}

func (bs *UIntBitSet) scanItemBitScanForward(index int) (uint, bool) {
	if index < 0 || index >= len(bs.items) {
		return 0, false
	}
	item := bs.items[index]
	if item == 0 {
		return 0, true
	}
	if item == FullItemMask {
		return 0, false
	}
	var i uint = 0
	for item >= 0 {
		if item&1 == 0 {
			return i, true
		}
		item >>= 1
		i++
	}
	return i, false
}
