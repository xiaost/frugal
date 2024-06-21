package reflect

import (
	"encoding/binary"
	"unsafe"
)

func (p *TestTypesForBenchmark) Encode(b []byte) (int, error) {
	if p == nil {
		b[0] = 0
		return 1, nil
	}
	off := 0

	// Field#1
	if p.B0 != true {
		b[off] = 2
		binary.BigEndian.PutUint16(b[off+1:], 1)
		off += 3
		b[off] = *((*byte)(unsafe.Pointer(&p.B0)))
		off++
	}

	// Field#2
	if p.B1 != nil {
		b[off] = 2
		binary.BigEndian.PutUint16(b[off+1:], 2)
		off += 3
		b[off] = *((*byte)(unsafe.Pointer(p.B1)))
		off++
	}

	// Field#3
	b[off] = 2
	binary.BigEndian.PutUint16(b[off+1:], 3)
	off += 3
	b[off] = *((*byte)(unsafe.Pointer(&p.B2)))
	off++

	// Field#11
	if p.Str0 != "8" {
		b[off] = 11
		binary.BigEndian.PutUint16(b[off+1:], 11)
		off += 3
		binary.BigEndian.PutUint32(b[off:], uint32(len(p.Str0)))
		off += 4 + copy(b[off+4:], p.Str0)
	}

	// Field#12
	if p.Str1 != nil {
		b[off] = 11
		binary.BigEndian.PutUint16(b[off+1:], 12)
		off += 3
		binary.BigEndian.PutUint32(b[off:], uint32(len(*p.Str1)))
		off += 4 + copy(b[off+4:], *p.Str1)
	}

	// Field#13
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 13)
	off += 3
	binary.BigEndian.PutUint32(b[off:], uint32(len(p.Str2)))
	off += 4 + copy(b[off+4:], p.Str2)

	// Field#21
	if p.Msg0 != nil {
		b[off] = 12
		binary.BigEndian.PutUint16(b[off+1:], 21)
		off += 3
		if n, err := p.Msg0.Encode(b[off:]); err != nil {
			return off, err
		} else {
			off += n
		}
	}

	// Field#22
	b[off] = 12
	binary.BigEndian.PutUint16(b[off+1:], 22)
	off += 3
	if n, err := p.Msg1.Encode(b[off:]); err != nil {
		return off, err
	} else {
		off += n
	}

	// Field#31
	if p.M0 != nil {
		b[off] = 13
		binary.BigEndian.PutUint16(b[off+1:], 31)
		off += 3
		b[off] = 8
		b[off+1] = 8
		binary.BigEndian.PutUint32(b[off+2:], uint32(len(p.M0)))
		off += 6
		for k, v := range p.M0 {
			binary.BigEndian.PutUint32(b[off:], uint32(k))
			off += 4
			binary.BigEndian.PutUint32(b[off:], uint32(v))
			off += 4
		}
	}

	// Field#32
	b[off] = 13
	binary.BigEndian.PutUint16(b[off+1:], 32)
	off += 3
	b[off] = 11
	b[off+1] = 12
	binary.BigEndian.PutUint32(b[off+2:], uint32(len(p.M1)))
	off += 6
	for k, v := range p.M1 {
		binary.BigEndian.PutUint32(b[off:], uint32(len(k)))
		off += 4 + copy(b[off+4:], k)
		if n, err := v.Encode(b[off:]); err != nil {
			return off, err
		} else {
			off += n
		}
	}

	// Field#41
	if p.L0 != nil {
		b[off] = 15
		binary.BigEndian.PutUint16(b[off+1:], 41)
		off += 3
		b[off] = 8
		binary.BigEndian.PutUint32(b[off+1:], uint32(len(p.L0)))
		off += 5
		for _, v := range p.L0 {
			binary.BigEndian.PutUint32(b[off:], uint32(v))
			off += 4
		}
	}

	// Field#42
	b[off] = 15
	binary.BigEndian.PutUint16(b[off+1:], 42)
	off += 3
	b[off] = 12
	binary.BigEndian.PutUint32(b[off+1:], uint32(len(p.L1)))
	off += 5
	for _, v := range p.L1 {
		if n, err := v.Encode(b[off:]); err != nil {
			return off, err
		} else {
			off += n
		}
	}

	// Field#51
	if p.Set0 != nil {
		b[off] = 14
		binary.BigEndian.PutUint16(b[off+1:], 51)
		off += 3
		b[off] = 8
		binary.BigEndian.PutUint32(b[off+1:], uint32(len(p.Set0)))
		off += 5
		for _, v := range p.Set0 {
			binary.BigEndian.PutUint32(b[off:], uint32(v))
			off += 4
		}
	}

	// Field#52
	b[off] = 14
	binary.BigEndian.PutUint16(b[off+1:], 52)
	off += 3
	b[off] = 11
	binary.BigEndian.PutUint32(b[off+1:], uint32(len(p.Set1)))
	off += 5
	for _, v := range p.Set1 {
		binary.BigEndian.PutUint32(b[off:], uint32(len(v)))
		off += 4 + copy(b[off+4:], v)
	}
	b[off] = 0 // STOP
	return off + 1, nil
}
func (p *Msg) Encode(b []byte) (int, error) {
	if p == nil {
		b[0] = 0
		return 1, nil
	}
	off := 0

	// Field#1
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 1)
	off += 3
	binary.BigEndian.PutUint32(b[off:], uint32(len(p.Message)))
	off += 4 + copy(b[off+4:], p.Message)

	// Field#2
	b[off] = 8
	binary.BigEndian.PutUint16(b[off+1:], 2)
	off += 3
	binary.BigEndian.PutUint32(b[off:], uint32(p.Type))
	off += 4
	b[off] = 0 // STOP
	return off + 1, nil
}
