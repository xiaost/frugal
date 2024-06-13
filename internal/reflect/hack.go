package reflect

import (
	"reflect"
	"unsafe"
)

// this func should be called once to test compatibility with Go runtime
func testhack() {
	m := map[int]string{7: "hello"}
	rv := reflect.ValueOf(m)
	it := rv.MapRange()
	it.Next()
	kp, vp := mapIterKeyValue(it)
	if *(*int)(kp) != 7 || *(*string)(vp) != "hello" || maplen(rv.UnsafePointer()) != 1 {
		panic("compatibility issue found: mapIterKeyValue")
	}
	m[8] = "world"
	m[9] = "!"
	if maplen(rv.UnsafePointer()) != 3 {
		panic("compatibility issue found: maplen")
	}
	rv1 := reflect.New(rv.Type()).Elem()
	rv1 = reflectValueWithPointer(rv, rv.UnsafePointer())
	m1, ok := rv1.Interface().(map[int]string)
	if !ok || m1[8] != "world" {
		panic("compatibility issue found: reflectValueWithPointer")
	}
}

func mapIterKeyValue(m *reflect.MapIter) (unsafe.Pointer, unsafe.Pointer) {
	type hitter struct {
		// k and v is always the 1st two fields of hitter
		// it will not be changed easily even though in the future
		k unsafe.Pointer
		v unsafe.Pointer
	}

	type hackMapIter struct {
		m      reflect.Value
		hitter hitter
	}
	p := (*hackMapIter)(unsafe.Pointer(m))
	return p.hitter.k, p.hitter.v
}

func maplen(p unsafe.Pointer) int {
	type hmap struct {
		count int // count is the 1st field
	}
	return (*hmap)(p).count
}

// reflectValueWithPointer returns reflect.Value with the unsafe.Pointer.
// Same reflect.NewAt().Elem() without the cost of getting abi.Type
func reflectValueWithPointer(rv reflect.Value, p unsafe.Pointer) reflect.Value {
	type rvtype struct { // reflect.Value
		abiType uintptr        // we don't touch it, leave it uintptr here.
		ptr     unsafe.Pointer // where the real pointer stored
	}
	(*rvtype)(unsafe.Pointer(&rv)).ptr = p
	return rv
}
