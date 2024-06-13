package reflect

import (
	"errors"
	"fmt"
	"reflect"
)

func EncodedSize(v interface{}) int {
	rv := normalizeEncodeRV(reflect.ValueOf(v))
	fd, err := createOrGetFieldDesc(rv.Type())
	if err != nil {
		return 0
	}
	t := &tType{fd: fd}
	n, err := t.EncodedSize(rv.UnsafePointer())
	if err != nil {
		panic(fmt.Sprintf("unexpected err: %s", err))
	}
	return n
}

func Encode(b []byte, v interface{}) (n int, err error) {
	rv := normalizeEncodeRV(reflect.ValueOf(v))
	fd, err := createOrGetFieldDesc(rv.Type())
	if err != nil {
		return 0, err
	}
	e := Encoder{}
	return e.Encode(b, rv.UnsafePointer(), fd)
}

func Decode(b []byte, v interface{}) (int, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return 0, errors.New("not a pointer")
	}
	if rv.IsNil() {
		return 0, errors.New("can't decode nil pointer")
	}
	if rv.Elem().Kind() != reflect.Struct {
		return 0, errors.New("not a pointer to a struct")
	}
	fd, err := createOrGetFieldDesc(rv.Type())
	if err != nil {
		return 0, err
	}
	d := decoderPool.Get().(*Decoder)
	n, err := d.Decode(b, rv.UnsafePointer(), fd)
	decoderPool.Put(d)
	return n, err
}
