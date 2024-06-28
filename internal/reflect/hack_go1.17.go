//go:build go1.17 && !go1.18

package reflect

import (
	"reflect"
	"unsafe"
)

type hackMapIter struct {
	m reflect.Value
	// it's a pointer before go1.18,
	// it causes allocation when calling m.MapRange & the 1st it.Next()
	hitter *hitter
}

func (iter *hackMapIter) initialized() bool { return iter.hitter != nil }

func (iter *hackMapIter) Next() (unsafe.Pointer, unsafe.Pointer) {
	mapiternext(unsafe.Pointer(iter.hitter))
	return iter.hitter.k, iter.hitter.v
}
