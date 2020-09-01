package set

import "errors"

type HBitSet struct {
	byLevel []*UIntBitSet
	max     uint
	size    uint
}

func NewHBitSet(max uint) *HBitSet {
	levelsCount := countLevels(max)
	result := &HBitSet{byLevel: make([]*UIntBitSet, levelsCount, levelsCount), max: max, size: 0}
	itemsAtLevel := max
	for i := 0; i < levelsCount; i++ {
		result.byLevel[i] = NewBitSet(itemsAtLevel)
		itemsAtLevel >>= BitsPerIndex
	}
	return result
}

func (bs *HBitSet) Add(val uint) {
	if bs.Has(val) {
		return
	}
	bs.byLevel[0].Add(val)
	bs.size++
	indexAtLevel := val >> BitsPerIndex
	for i := 1; i < len(bs.byLevel); i++ {
		if bs.byLevel[i-1].items[indexAtLevel] == FullItemMask {
			bs.byLevel[i].Add(indexAtLevel)
		}
		indexAtLevel >>= BitsPerIndex
	}
}

func (bs *HBitSet) Delete(val uint) {
	if bs.Has(val) {
		bs.byLevel[0].Delete(val)
		bs.size--
		indexAtLevel := val >> BitsPerIndex
		for i := 1; i < len(bs.byLevel); i++ {
			bs.byLevel[i].Delete(indexAtLevel)
			indexAtLevel >>= BitsPerIndex
		}
	}
}

func (bs *HBitSet) FirstAbsent(from uint) (uint, error) {
	if bs.size == bs.max {
		return 0, errors.New("all items are present")
	}

	var current uint
	for level := len(bs.byLevel) - 1; level >= 0; level-- {
		next, err := bs.byLevel[level].FirstAbsent(current)
		if err != nil {
			return next, err
		}
		current = next
	}
	return current, nil
}

func (bs *HBitSet) Has(val uint) bool {
	return bs.byLevel[0].Has(val)
}

func (bs *HBitSet) Cap() uint {
	return bs.size
}

func (bs *HBitSet) Iterator() <-chan uint {
	return bs.byLevel[0].Iterator()
}

func countLevels(max uint) int {
	levels := 0
	for max > 0 {
		max >>= BitsPerIndex
		levels++
	}
	return levels
}
