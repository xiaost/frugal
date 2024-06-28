package reflect

import (
	"reflect"
	"unsafe"
)

// rvUnsafePointer backports rv.UnsafePointer for go1.17
// TODO: remove this func and use rv.UnsafePointer() directly when >= go1.18
func rvUnsafePointer(rv reflect.Value) unsafe.Pointer {
	return unsafe.Pointer(rv.Pointer())
}
