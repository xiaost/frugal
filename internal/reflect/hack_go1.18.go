//go:build go1.18

package reflect

import (
	"reflect"
	"unsafe"
)

type hackMapIter struct {
	m      reflect.Value
	hitter hitter // it's a pointer before go1.18
}

func (iter *hackMapIter) initialized() bool { return iter.hitter.k != nil }

func (iter *hackMapIter) Next() (unsafe.Pointer, unsafe.Pointer) {
	mapiternext(unsafe.Pointer(&iter.hitter))
	return iter.hitter.k, iter.hitter.v
}
