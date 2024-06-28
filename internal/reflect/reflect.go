package reflect

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

func EncodedSize(v interface{}) int {
	testhackOnce.Do(testhack)
	rv := reflect.ValueOf(v)
	fd := getFieldDesc(rv) // copy get and create funcs here for inlining
	if fd == nil {
		var err error
		fd, err = createFieldDesc(rv)
		if err != nil {
			panic(fmt.Sprintf("unexpected err when parse fields: %s", err))
		}
	}

	// get underlying pointer
	var p unsafe.Pointer
	if rv.Kind() == reflect.Struct {
		// direct eface
		// need to copy to heap, and then get the ptr
		prv := fd.rvPool.Get().(*reflect.Value)
		defer fd.rvPool.Put(prv)
		(*prv).Elem().Set(rv)
		p = (*rvtype)(unsafe.Pointer(prv)).ptr // like `rvPtr` without copy
	} else {
		// we only supports one indirect like *struct
		// it doesn't support **struct
		p = rvPtr(rv)
		if p == nil {
			return 1 // tSTOP
		}
	}

	t := &tType{fd: fd}
	n, err := t.EncodedSize(p)
	if err != nil {
		panic(fmt.Sprintf("unexpected err: %s", err))
	}
	return n
}

func Encode(b []byte, v interface{}) (n int, err error) {
	testhackOnce.Do(testhack)
	rv := reflect.ValueOf(v)
	fd := getFieldDesc(rv) // copy get and create funcs here for inlining
	if fd == nil {
		fd, err = createFieldDesc(rv)
		if err != nil {
			return 0, err
		}
	}

	// get underlying pointer
	var p unsafe.Pointer
	if rv.Kind() == reflect.Struct {
		// direct eface
		// need to copy to heap, and then get the ptr
		prv := fd.rvPool.Get().(*reflect.Value)
		defer fd.rvPool.Put(prv)
		(*prv).Elem().Set(rv)
		p = (*rvtype)(unsafe.Pointer(prv)).ptr // like `rvPtr` without copy
	} else {
		// we only supports one indirect like *struct
		// it doesn't support **struct
		p = rvPtr(rv)
		if p == nil {
			b[0] = 0 // tSTOP
			return 1, nil
		}
	}
	e := Encoder{}
	return e.Encode(b, rvUnsafePointer(rv), fd)
}

func Decode(b []byte, v interface{}) (int, error) {
	testhackOnce.Do(testhack)
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return 0, errors.New("not a pointer")
	}
	if rv.IsNil() {
		return 0, errors.New("can't decode nil pointer")
	}
	if rv.Elem().Kind() != reflect.Struct {
		return 0, errors.New("not a pointer to a struct")
	}
	fd, err := getOrcreateFieldDesc(rv)
	if err != nil {
		return 0, err
	}
	d := decoderPool.Get().(*Decoder)
	n, err := d.Decode(b, rvUnsafePointer(rv), fd, maxDepthLimit)
	decoderPool.Put(d)
	return n, err
}
