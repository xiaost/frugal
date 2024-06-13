package reflect

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/cloudwego/frugal/internal/binary/defs"
)

type tType struct {
	T ttype
	K *tType
	V *tType

	WT ttype // wiretype tNUM -> tI32

	Tag defs.Tag

	RT    reflect.Type
	Size  int
	Align int

	rv reflect.Value

	IsPointer  bool // true if t.Tag == defs.T_pointer
	SimpleType bool // true if simpleTypes[t.T]
	FixedSize  int  // typeToSize[t.T]

	// for tSTRUCT
	fd *FieldDesc

	// for tLIST, tSET, tMAP, tSTRUCT
	encodedSizeFunc func(p unsafe.Pointer) (int, error)

	// tMAP only
	mapTmpVarsPool *sync.Pool // for decoder tmp vars
}

// Equal returns true if data of two pointers point to.
func (t *tType) Equal(p0, p1 unsafe.Pointer) bool {
	switch t.T {
	case tBOOL:
		return *(*bool)(p0) == *(*bool)(p1)
	case tBYTE:
		return *(*int8)(p0) == *(*int8)(p1)
	case tDOUBLE:
		return *(*float64)(p0) == *(*float64)(p1)
	case tI16:
		return *(*int16)(p0) == *(*int16)(p1)
	case tI32:
		return *(*int32)(p0) == *(*int32)(p1)
	case tI64, tENUM:
		return *(*int64)(p0) == *(*int64)(p1)
	case tSTRING:
		return *(*string)(p0) == *(*string)(p1)
	}
	return false
}

func (t *tType) fromDefsType(x *defs.Type) {
	t.T = ttype(x.Tag())
	t.WT = t.T
	t.Tag = x.T
	if t.Tag == defs.T_enum {
		t.T = tENUM
	}
	t.RT = x.S
	t.Size = int(x.S.Size())
	t.Align = x.S.Align()
	if x.K != nil {
		t.K = &tType{}
		t.K.fromDefsType(x.K)
	}
	if x.V != nil {
		t.V = &tType{}
		t.V.fromDefsType(x.V)
	}

	t.rv = reflect.New(t.RT).Elem()

	if t.T == tMAP {
		t.mapTmpVarsPool = initOrGetMapTmpVarsPool(t)
	}
	t.IsPointer = t.Tag == defs.T_pointer
	t.SimpleType = simpleTypes[t.T]
	t.FixedSize = int(typeToSize[t.T])
	switch t.T {
	case tMAP:
		t.encodedSizeFunc = t.encodedMapSize
	case tLIST, tSET:
		t.encodedSizeFunc = t.encodedListSize
	case tSTRUCT:
		t.encodedSizeFunc = t.EncodedSize
	}
}

const (
	fieldHeaderLen = 1 + 2     // type + id
	mapHeaderLen   = 1 + 1 + 4 // k type, v type, len
	listHeaderLen  = 1 + 4     // elem type, len
)

func encodedStringSize(p unsafe.Pointer) int {
	// string type in list or map, it's always non-pointer
	// so we no need to do the check of t.IsPointer
	return 4 + (*reflect.SliceHeader)(p).Len
}

func (t *tType) EncodedSize(base unsafe.Pointer) (int, error) {
	fd := t.fd
	if t.T != 0 { // not from reflect.EncodedSize
		// for field of a struct, value of a map, or elem of a list,
		// it's a poitner to struct pointer, then we have to convert it to struct pointer
		base = unsafe.Pointer(*(*uintptr)(base))
	}
	if base == nil {
		return 1, nil // tSTOP
	}
	ret := fd.fixedLenFieldSize
	for _, f := range fd.varLenFields {
		p := unsafe.Pointer(uintptr(base) + f.Offset)
		if f.CanSkipEncodeIfNil && unsafe.Pointer(*(*uintptr)(p)) == nil {
			continue
		}
		t := &f.Type
		if f.CanSkipIfDefault && t.Equal(f.Default, p) {
			continue
		}
		if n := t.FixedSize; n > 0 {
			ret += (fieldHeaderLen + int(n))
			// fast skip types like tBOOL, tBYTE, tDOUBLE, tI16, tI32, tI64
			continue
		}
		if t.T == tSTRING {
			if t.IsPointer {
				p = unsafe.Pointer(*(*uintptr)(p))
			}
			ret += fieldHeaderLen + encodedStringSize(p)
			continue
		}
		ret += fieldHeaderLen
		n, err := t.encodedSizeFunc(p) // tLIST, tSET, tMAP, tSTRUCT
		if err != nil {
			return ret, err
		}
		ret += n
	}
	ret += 1 // tSTOP
	return ret, nil
}

func (t *tType) encodedMapSize(p unsafe.Pointer) (int, error) {
	if unsafe.Pointer(*(*uintptr)(p)) == nil {
		// We always encode nil map for required or default requiredness
		return mapHeaderLen, nil // 0-len map
	}

	kt, doneK := t.K, false
	vt, doneV := t.V, false
	l := maplen(unsafe.Pointer(*(*uintptr)(p)))
	ret := mapHeaderLen
	if kt.FixedSize > 0 {
		ret += l * kt.FixedSize
		doneK = true
	}
	if vt.FixedSize > 0 {
		ret += l * vt.FixedSize
		doneV = true
	}
	if doneK && doneV {
		return ret, nil // fast path
	}

	mv := reflectValueWithPointer(t.rv, p)

	// we already skipped primitive types.
	// need to handle tSTRING, tMAP, tLIST, tSET or tSTRUCT
	for it := mv.MapRange(); it.Next(); { // XXX: use unexported runtime mapiterxxx?
		kp, vp := mapIterKeyValue(it)
		// Key
		// tSTRING
		if !doneK { // tSTRING
			if kt.T != tSTRING {
				panic("unexpected key type")
			}
			ret += encodedStringSize(kp)
		}
		if doneV {
			continue
		}
		// Value
		// tSTRING, tMAP, tLIST, tSET or tSTRUCT
		if vt.T == tSTRING {
			ret += encodedStringSize(vp)
			continue
		}
		n, err := vt.encodedSizeFunc(vp)
		if err != nil {
			return ret, err
		}
		ret += n
	}
	return ret, nil
}

func (t *tType) encodedListSize(p unsafe.Pointer) (int, error) {
	if unsafe.Pointer(*(*uintptr)(p)) == nil {
		return listHeaderLen, nil // 0-len list
	}
	vt := t.V
	h := (*reflect.SliceHeader)(p)
	if vt.FixedSize > 0 {
		return listHeaderLen + (h.Len * vt.FixedSize), nil
	}
	ret := listHeaderLen
	vp := unsafe.Pointer(h.Data)
	for i := 0; i < h.Len; i++ {
		if vt.T == tSTRING {
			ret += encodedStringSize(vp)
		} else {
			n, err := vt.encodedSizeFunc(vp)
			if err != nil {
				return ret, err
			}
			ret += n
		}
		vp = unsafe.Pointer(uintptr(vp) + uintptr(vt.Size)) // move to next element
	}
	return ret, nil
}
