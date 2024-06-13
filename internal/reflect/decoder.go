package reflect

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"unsafe"

	"github.com/cloudwego/frugal/internal/binary/defs"
)

// defaultDecoderMemSize controls the min block mem used to malloc,
// DO NOT increase it mindlessly which would cause mem issue,
// coz objects even use one byte of the mem, it won't be released.
const defaultDecoderMemSize = 512

var decoderPool = sync.Pool{
	New: func() interface{} {
		d := &Decoder{}
		d.p = 0
		d.b = make([]byte, defaultDecoderMemSize)
		return d
	},
}

type Decoder struct {
	p int
	b []byte
}

func (d *Decoder) Malloc(n, align int) unsafe.Pointer {
	mask := align - 1
	if need := n + mask; d.p+need > len(d.b) {
		sz := defaultDecoderMemSize
		if need > sz {
			sz = need
		}
		d.p = 0
		d.b = make([]byte, sz)
	}
	p0 := unsafe.Pointer(&d.b[d.p])
	p1 := (uintptr(p0) + uintptr(mask)) & ^(uintptr(mask)) // memory addr alignment
	d.p += (n + int(p1-uintptr(p0)))
	return unsafe.Pointer(p1)
}

func (d *Decoder) mallocIfPointer(t *tType, p unsafe.Pointer) unsafe.Pointer {
	if t.IsPointer {
		// we need to malloc the type first before assigning a value to it
		ret := d.Malloc(t.V.Size, t.V.Align)
		*((*uintptr)(p)) = (uintptr)(ret)
		return ret
	}
	return p
}

func (d *Decoder) Decode(b []byte, base unsafe.Pointer, fd *FieldDesc) (int, error) {
	var bitset *fieldBitset
	if len(fd.requiredFields) > 0 {
		bitset = bitsetPool.Get().(*fieldBitset)
		defer bitsetPool.Put(bitset)
		for _, f := range fd.requiredFields {
			bitset.unset(f.ID)
		}
	}

	i := 0
	for {
		tp := ttype(b[i])
		i++
		if tp == tSTOP {
			break
		}
		fid := binary.BigEndian.Uint16(b[i:])
		i += 2

		f := fd.GetField(fid)
		if f == nil {
			n, err := skipType(tp, b[i:])
			if err != nil {
				return i, fmt.Errorf("skip unknown field %d of struct %s err: %w", fid, fd.rt.String(), err)
			}
			i += n
			continue
		}
		t := &f.Type
		if t.WT != tp {
			return i, errors.New("type mismatch")
		}
		p := unsafe.Pointer(uintptr(base) + f.Offset) // pointer to the field
		p = d.mallocIfPointer(t, p)
		n, err := d.decodeType(t, b[i:], p)
		if err != nil {
			return i, fmt.Errorf("decode field %d of struct %s err: %w", fid, fd.rt.String(), err)
		}
		if bitset != nil {
			bitset.set(f.ID)
		}
		i += n
	}
	for _, f := range fd.requiredFields {
		if !bitset.test(f.ID) {
			return i, newRequiredFieldNotSetException(lookupFieldName(fd.rt, f.Offset))
		}
	}
	return i, nil
}

func (d *Decoder) decodeType(t *tType, b []byte, p unsafe.Pointer) (int, error) {
	switch t.T {
	case tBOOL:
		*((*bool)(p)) = (b[0] > 0)
		return 1, nil
	case tBYTE:
		*((*byte)(p)) = b[0]
		return 1, nil
	case tDOUBLE:
		n := binary.BigEndian.Uint64(b)
		*((*float64)(p)) = math.Float64frombits(n)
		return 8, nil
	case tI16:
		*((*int16)(p)) = int16(binary.BigEndian.Uint16(b))
		return 2, nil
	case tI32:
		*((*int32)(p)) = int32(binary.BigEndian.Uint32(b))
		return 4, nil
	case tENUM:
		*((*int64)(p)) = int64(int32(binary.BigEndian.Uint32(b)))
		return 4, nil
	case tI64:
		*((*int64)(p)) = int64(binary.BigEndian.Uint64(b))
		return 8, nil
	case tSTRING:
		i := 0
		l := int(binary.BigEndian.Uint32(b))
		i += 4
		x := d.Malloc(l, 1)
		if t.Tag == defs.T_binary {
			h := (*reflect.SliceHeader)(p)
			h.Data = uintptr(x)
			h.Len = l
			h.Cap = l
		} else { //  convert to str
			h := (*reflect.StringHeader)(p)
			h.Data = uintptr(x)
			h.Len = l
		}
		copyn(x, b[i:], l)
		i += l
		return i, nil
	case tMAP:
		// map header
		t0, t1, l := ttype(b[0]), ttype(b[1]), int(binary.BigEndian.Uint32(b[2:]))
		i := 6

		// check types
		kt := t.K
		vt := t.V
		if t0 != kt.T || t1 != vt.T {
			return 0, errors.New("type mismatch")
		}

		// decode map

		tmp := t.mapTmpVarsPool.Get().(*tmpMapVars)
		defer t.mapTmpVarsPool.Put(tmp)
		k := tmp.k
		v := tmp.v
		kp := tmp.kp
		vp := tmp.vp
		m := reflect.MakeMapWithSize(t.RT, l)
		*((*uintptr)(p)) = m.Pointer() // p = make(t.RT, l)
		for j := 0; j < l; j++ {
			if n, err := d.decodeType(kt, b[i:], kp); err != nil {
				return i, err
			} else {
				i += n
			}
			// v can be pointer, k not
			// for v, we need to malloc space for the pointer
			p := d.mallocIfPointer(vt, vp)
			if n, err := d.decodeType(vt, b[i:], p); err != nil {
				return i, err
			} else {
				i += n
			}
			m.SetMapIndex(k, v)
		}
		return i, nil
	case tLIST, tSET: // NOTE: for tSET, it may be map in the future
		// list header
		tp, l := ttype(b[0]), int(binary.BigEndian.Uint32(b[1:]))
		i := 5

		// check types
		et := t.V
		if et.T != tp {
			return 0, errors.New("type mismatch")
		}

		// decode list
		x := d.Malloc(l*et.Size, et.Align) // malloc for slice. make([]Type, l, l)
		h := (*reflect.SliceHeader)(p)     // update the slice field
		h.Data = uintptr(x)
		h.Len = l
		h.Cap = l
		p = unsafe.Pointer(h.Data)
		for j := 0; j < l; j++ {
			n, err := d.decodeType(et, b[i:], d.mallocIfPointer(et, p))
			if err != nil {
				return i, err
			}
			i += n
			p = unsafe.Pointer(uintptr(p) + uintptr(et.Size)) // next element
		}
		return i, nil
	case tSTRUCT:
		return d.Decode(b, p, t.fd)
	}
	return 0, fmt.Errorf("unknown type: %d", t.T)
}

func skipType(t ttype, b []byte) (int, error) {
	if n := typeToSize[t]; n > 0 {
		return int(n), nil
	}
	switch t {
	case tSTRING:
		l := int(binary.BigEndian.Uint32(b))
		return 4 + l, nil
	case tMAP:
		i := 6
		t0, t1, l := ttype(b[0]), ttype(b[1]), int(binary.BigEndian.Uint32(b[2:]))
		for j := 0; j < l; j++ {
			if n, err := skipType(t0, b[i:]); err != nil {
				return i, err
			} else {
				i += n
			}
			if n, err := skipType(t1, b[i:]); err != nil {
				return i, err
			} else {
				i += n
			}
		}
		return i, nil
	case tLIST, tSET:
		i := 5
		et, l := ttype(b[0]), int(binary.BigEndian.Uint32(b[1:]))
		for j := 0; j < l; j++ {
			n, err := skipType(et, b[i:])
			if err != nil {
				return i, err
			}
			i += n
		}
		return i, nil
	case tSTRUCT:
		i := 0
		for {
			t := ttype(b[i])
			i += 1
			if t == tSTOP {
				return i, nil
			}
			i += 2 // field id
			n, err := skipType(t, b[i:])
			if err != nil {
				return i, err
			}
			i += n
		}
		return i, nil
	}
	return 0, errors.New("unknown type")
}
