package reflect

import (
	"errors"
	"reflect"
	"sync"
	"unsafe"

	"github.com/cloudwego/frugal/internal/binary/defs"
)

var fdsMu sync.RWMutex
var fds = map[reflect.Type]*FieldDesc{}

type ttype uint8

const (
	tSTOP   ttype = 0
	tVOID   ttype = 1
	tBOOL   ttype = 2
	tBYTE   ttype = 3
	tI08    ttype = 3
	tDOUBLE ttype = 4
	tI16    ttype = 6
	tI32    ttype = 8
	tI64    ttype = 10
	tSTRING ttype = 11
	tUTF7   ttype = 11
	tSTRUCT ttype = 12
	tMAP    ttype = 13
	tSET    ttype = 14
	tLIST   ttype = 15
	tUTF8   ttype = 16
	tUTF16  ttype = 17

	// internal use only
	tENUM ttype = 0xfe // XXX: kitex issue, int64, but encode as int32 ...
)

func createOrGetFieldDesc(t reflect.Type) (*FieldDesc, error) {
	fdsMu.RLock()
	d := fds[t]
	fdsMu.RUnlock()
	if d != nil {
		return d, nil
	}

	// slow path

	fdsMu.Lock()
	defer fdsMu.Unlock()
	if d := fds[t]; d != nil {
		return d, nil
	}
	fd, err := newFieldDesc(t)
	if err != nil {
		return nil, err
	}
	fds[t] = fd
	if err := prefetchSubFieldDesc(fd); err != nil {
		delete(fds, t)
		return nil, err
	}
	return fd, nil
}

func prefetchSubFieldDesc(d *FieldDesc) error {
	for i := range d.fields {
		var t *tType
		f := d.fields[i]
		if f.Type.T == tSTRUCT {
			t = &f.Type
		} else if f.Type.T == tMAP && f.Type.V.T == tSTRUCT {
			t = f.Type.V
		} else if f.Type.T == tLIST && f.Type.V.T == tSTRUCT {
			t = f.Type.V
		} else {
			continue
		}
		fd, ok := fds[t.RT]
		if ok {
			t.fd = fd
			continue
		}
		fd, err := newFieldDesc(t.RT)
		if err != nil {
			return err
		}
		fds[t.RT] = fd
		if err := prefetchSubFieldDesc(fd); err != nil {
			delete(fds, t.RT)
			return err
		}
		t.fd = fd
	}
	return nil
}

type FieldDesc struct {
	rt reflect.Type // Kind() == reflect.Struct

	maxID    uint16
	fieldIdx []int // directly maps field id to Field for performance
	fields   []*tField

	fixedLenFieldSize int       // sum of f.EncodedSize() > 0
	varLenFields      []*tField // list of fields that f.EncodedSize() <= 0
	requiredFields    []*tField
}

func newFieldDesc(t reflect.Type) (*FieldDesc, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("not struct")
	}
	ff, err := defs.ResolveFields(t)
	if err != nil {
		return nil, err
	}
	d := &FieldDesc{rt: t}
	d.fromDefsFields(ff)
	return d, nil
}

func (d *FieldDesc) GetField(fid uint16) *tField {
	if fid > d.maxID {
		return nil
	}
	i := d.fieldIdx[fid]
	if i < 0 {
		return nil
	}
	return d.fields[i]
}

func (d *FieldDesc) fromDefsFields(ff []defs.Field) {
	maxFieldID := uint16(0)
	for _, f := range ff {
		if f.ID > maxFieldID {
			maxFieldID = f.ID
		}
	}
	d.maxID = maxFieldID
	d.fieldIdx = make([]int, maxFieldID+1)
	for i := range d.fieldIdx {
		d.fieldIdx[i] = -1
	}
	d.fields = make([]*tField, len(ff))
	for i, f := range ff {
		d.fields[i] = &tField{}
		d.fields[i].fromDefsField(f)
		d.fieldIdx[f.ID] = i
	}
	d.varLenFields = make([]*tField, 0, len(ff))
	for _, f := range d.fields {
		if n := f.EncodedSize(); n > 0 {
			d.fixedLenFieldSize += n
		} else {
			d.varLenFields = append(d.varLenFields, f)
		}
		if f.IsRequired() {
			d.requiredFields = append(d.requiredFields, f)
		}
	}
}

type tField struct {
	ID     uint16
	Offset uintptr
	Type   tType

	Opts    defs.Options
	Spec    defs.Requiredness
	Default unsafe.Pointer

	CanSkipEncodeIfNil bool
	CanSkipIfDefault   bool
}

func (f *tField) IsRequired() bool { return f.Spec == defs.Required }
func (f *tField) IsOptional() bool { return f.Spec == defs.Optional }

var containerTypes = [256]bool{
	tMAP:  true,
	tLIST: true,
	tSET:  true,
}

var typeToSize = [256]int8{
	tBOOL:   1,
	tBYTE:   1,
	tDOUBLE: 8,
	tI16:    2,
	tI32:    4,
	tI64:    8,
	tENUM:   4,
}

// EncodedSize returns encoded size of the field, -1 if can not be determined.
func (f *tField) EncodedSize() int {
	if f.Type.IsPointer { // may be nil, then skip encoding
		return -1
	}
	if f.IsOptional() {
		return -1 // may have default value, then skip encoding
	}
	if f.Type.FixedSize > 0 {
		return fieldHeaderLen + f.Type.FixedSize // type + id + len
	}
	return -1
}

func (f *tField) fromDefsField(x defs.Field) {
	f.ID = x.ID
	f.Offset = uintptr(x.F)
	f.Type.fromDefsType(x.Type)
	f.Opts = x.Opts
	f.Spec = x.Spec

	t := &f.Type

	// for map or slice, t.IsPointer() is false,
	// but we can consider the types as pointer as per lang spec
	// for defs.T_binary, actually it's []byte, like tLIST
	f.CanSkipEncodeIfNil = f.Spec == defs.Optional &&
		(t.Tag == defs.T_pointer || t.Tag == defs.T_binary || containerTypes[t.T])

	// for SkipEncodeDefault
	v := x.Default
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if !v.IsValid() {
		return
	}
	f.Default = unsafe.Pointer(v.UnsafeAddr())
	f.CanSkipIfDefault = (f.Spec == defs.Optional) &&
		t.Tag != defs.T_pointer && // normally if fields with default values, it's non-pointer
		f.Default != nil // the field must have default value
}
