package internal

import (
	"errors"
	"github.com/pavlo-iuriichuk/cron/internal/set"
	"sync"
)

type Allocator interface {
	Init()
	Alloc() (uint, error)
	Free(id uint) error
	Cap() uint
	Iterator() <-chan uint
}

const MinId = uint(0)

type HAllocator struct {
	lock  sync.RWMutex
	size  uint
	items set.UIntSet
}

func NewHAllocator(size uint) *HAllocator {
	return &HAllocator{
		size:  size,
		items: set.NewHBitSet(size),
	}
}

func (alc *HAllocator) Init() {
	defer alc.lock.Unlock()
	alc.lock.Lock()

	alc.items = set.NewHBitSet(alc.size)
}

func (alc *HAllocator) Alloc() (uint, error) {
	probe, err := alc.items.FirstAbsent(0)
	if err != nil {
		return probe, err
	}
	alc.items.Add(probe)
	return probe, nil
}

func (alc *HAllocator) Free(id uint) error {
	defer alc.lock.Unlock()
	alc.lock.Lock()

	if id >= alc.size {
		return errors.New("id is out of bounds")
	}

	alc.items.Delete(id)
	return nil
}

func (alc *HAllocator) Cap() uint {
	defer alc.lock.RUnlock()
	alc.lock.RLock()

	return alc.size - alc.items.Cap()
}

func (alc *HAllocator) Iterator() <-chan uint {
	return alc.items.Iterator()
}
