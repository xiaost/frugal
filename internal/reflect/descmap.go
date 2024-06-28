package reflect

import (
	"sync/atomic"
	"unsafe"
)

// mapFieldDesc represents a read-lock-free hashmap for *FieldDesc like sync.Map.
// it's NOT designed for writes.
type mapFieldDesc struct {
	p unsafe.Pointer // for atomic, point to hashtable
}

// XXX: fixed size to make it simple,
// we may not so many structs that need to rehash it
const mapFieldDescBuckets = 0xffff

type mapFieldDescItem struct {
	abiType uintptr
	fd      *FieldDesc
}

func newMapFieldDesc() *mapFieldDesc {
	m := &mapFieldDesc{}
	buckets := make([][]mapFieldDescItem, mapFieldDescBuckets+1) // [0] - [0xffff]
	atomic.StorePointer(&m.p, unsafe.Pointer(&buckets))
	return m
}

// Get ...
func (m *mapFieldDesc) Get(abiType uintptr) *FieldDesc {
	buckets := *(*[][]mapFieldDescItem)(atomic.LoadPointer(&m.p))
	dd := buckets[abiType&mapFieldDescBuckets]
	for i := range dd {
		if dd[i].abiType == abiType {
			return dd[i].fd
		}
	}
	return nil
}

// Set ...
// createOrGetFieldDesc will protect calling Set with lock
func (m *mapFieldDesc) Set(abiType uintptr, fd *FieldDesc) {
	if m.Get(abiType) == fd {
		return
	}
	oldBuckets := *(*[][]mapFieldDescItem)(atomic.LoadPointer(&m.p))
	newBuckets := make([][]mapFieldDescItem, mapFieldDescBuckets+1)
	copy(newBuckets, oldBuckets)
	bk := abiType & mapFieldDescBuckets
	newBuckets[bk] = append(newBuckets[bk], mapFieldDescItem{abiType: abiType, fd: fd})
	atomic.StorePointer(&m.p, unsafe.Pointer(&newBuckets))
}
