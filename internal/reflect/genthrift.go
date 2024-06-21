package reflect

import (
	"fmt"
	"io"
	"reflect"
)

type ThriftCodeGenerator struct {
}

// same as fmt.Fprintf with "\n" ...
func fmtFprintf(w io.Writer, fmtstr string, args ...interface{}) {
	fmt.Fprintf(w, fmtstr+"\n", args...)
}

func wiretype(t ttype) ttype {
	if t == tENUM { // XXX: ...
		return tI32
	}
	return t
}

func pointToVar(t *tType, varname string) string {
	if t.IsPointer {
		return "*" + varname
	}
	return varname
}

func pointerOfVar(t *tType, varname string) string {
	if t.IsPointer {
		return varname
	}
	return "&" + varname
}

func (g *ThriftCodeGenerator) GenEncode(w io.Writer, fd *FieldDesc) {
	// func(p *{TypeName}) Encode(b []byte) (int, error) { ... }
	fmtFprintf(w, "func(p *%s) Encode(b []byte) (int, error) {", fd.rt.Name())
	// check struct pointer nil
	// we write a single stop byte for the case
	fmtFprintf(w, "if p == nil { b[0] = 0; return 1, nil; }")
	fmtFprintf(w, "off := 0")
	for _, f := range fd.fields {
		g.genEncodeField(w, fd.rt, &f)
	}
	fmtFprintf(w, "b[off] = 0 // STOP")
	fmtFprintf(w, "return off + 1, nil")
	fmtFprintf(w, "}")
}

func (g *ThriftCodeGenerator) genEncodeField(w io.Writer, rt reflect.Type, f *tField) {
	t := &f.Type
	// skipcheck, for optional fields
	varname := "p." + lookupFieldName(rt, f.Offset)
	skipcheck := true
	fmtFprintf(w, "\n// Field#%d", f.ID)
	if f.CanSkipEncodeIfNil {
		fmtFprintf(w, "if %s != nil {", varname)
	} else if f.CanSkipIfDefault {
		// only for simple types, for containers only check nil
		fmtFprintf(w, "if %s != %#v {",
			varname, reflect.NewAt(t.RT, f.Default).Elem())
	} else {
		skipcheck = false
	}
	// field header
	fmtFprintf(w, "b[off] = %d", wiretype(t.T))
	fmtFprintf(w, "binary.BigEndian.PutUint16(b[off+1:], %d) ", f.ID)
	fmtFprintf(w, "off += 3")

	g.genEncodeType(w, f, t, varname, 0)

	// end field encoding
	if skipcheck {
		fmtFprintf(w, "}")
	}
}

func (g *ThriftCodeGenerator) genEncodeType(w io.Writer, f *tField, t *tType, varname string, depth int) {
	switch t.T {
	case tBOOL:
		g.genEncodeBool(w, t, varname)
	case tBYTE:
		g.genEncodeByte(w, t, varname)
	case tDOUBLE:
		g.genEncodeDouble(w, t, varname)
	case tI16:
		g.genEncodeInt16(w, t, varname)
	case tI32:
		g.genEncodeInt32(w, t, varname)
	case tI64:
		g.genEncodeInt64(w, t, varname)
	case tSTRING:
		g.genEncodeString(w, t, varname)
	case tENUM:
		g.genEncodeEnum(w, t, varname)
	case tLIST, tSET:
		g.genEncodeList(w, f, t, varname, depth)
	case tMAP:
		g.genEncodeMap(w, f, t, varname, depth)
	case tSTRUCT:
		g.genEncodeStruct(w, t, varname)
	default:
		panic(fmt.Sprintf("unexpected type: %d", t.T))
	}
}

func (g *ThriftCodeGenerator) genEncodeBool(w io.Writer, t *tType, varname string) {
	// for bool, the underlying byte of true is always 1, and 0 for false
	// which is same as thrift binary protocol
	fmtFprintf(w, "b[off] = *((*byte)(unsafe.Pointer(%s)))", pointerOfVar(t, varname))
	fmt.Fprintln(w, "off++")
}

func (g *ThriftCodeGenerator) genEncodeByte(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "b[off] = %s", pointToVar(t, varname))
	fmtFprintf(w, "off++")
}

func (g *ThriftCodeGenerator) genEncodeDouble(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "binary.BigEndian.PutUint64(b[off:], *(*uint64)(unsafe.Pointer(%s)))", pointerOfVar(t, varname))
	fmtFprintf(w, "off += 8")
}

func (g *ThriftCodeGenerator) genEncodeInt16(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "binary.BigEndian.PutUint16(b[off:], %s)", pointToVar(t, varname))
	fmtFprintf(w, "off += 2")
}

func (g *ThriftCodeGenerator) genEncodeInt32(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "binary.BigEndian.PutUint32(b[off:], uint32(%s))", pointToVar(t, varname))
	fmtFprintf(w, "off += 4")
}

func (g *ThriftCodeGenerator) genEncodeInt64(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "binary.BigEndian.PutUint64(b[off:], %s)", pointToVar(t, varname))
	fmtFprintf(w, "off += 8")
}

func (g *ThriftCodeGenerator) genEncodeString(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "binary.BigEndian.PutUint32(b[off:], uint32(len(%s)))", pointToVar(t, varname))
	fmtFprintf(w, "off += 4 + copy(b[off+4:], %s)", pointToVar(t, varname))
}

func (g *ThriftCodeGenerator) genEncodeEnum(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "binary.BigEndian.PutUint32(b[off:], uint32(%s))", pointToVar(t, varname))
	fmtFprintf(w, "off += 4")
}

func (g *ThriftCodeGenerator) genEncodeStruct(w io.Writer, t *tType, varname string) {
	fmtFprintf(w, "if n, err := %s.Encode(b[off:]); err != nil { return off, err } else { off += n }", varname)
}

func (g *ThriftCodeGenerator) genEncodeList(w io.Writer, f *tField, t *tType, varname string, depth int) {
	// list header
	fmtFprintf(w, "b[off] = %d", wiretype(t.V.T))
	fmtFprintf(w, "binary.BigEndian.PutUint32(b[off+1:], uint32(len(%s)))", pointToVar(t, varname))
	fmtFprintf(w, "off += 5")

	// iteration tmp var
	tmpvvar := "v"
	if depth > 0 {
		// avoid redeclared vars
		tmpvvar = fmt.Sprintf("v%d", depth-1)
	}
	fmtFprintf(w, "for _, %s := range %s {", tmpvvar, pointToVar(t, varname))
	g.genEncodeType(w, f, t.V, tmpvvar, depth+1)
	fmtFprintf(w, "}")
}

func (g *ThriftCodeGenerator) genEncodeMap(w io.Writer, f *tField, t *tType, varname string, depth int) {
	// map header
	fmtFprintf(w, "b[off] = %d", wiretype(t.K.T))
	fmtFprintf(w, "b[off+1] = %d", wiretype(t.V.T))
	fmtFprintf(w, "binary.BigEndian.PutUint32(b[off+2:], uint32(len(%s)))", pointToVar(t, varname))
	fmtFprintf(w, "off += 6")

	// iteration tmp vars
	tmpkvar := "k"
	tmpvvar := "v"
	if depth > 0 {
		// avoid redeclared vars
		tmpkvar = fmt.Sprintf("k%d", depth-1)
		tmpvvar = fmt.Sprintf("v%d", depth-1)
	}
	fmtFprintf(w, "for %s, %s := range %s {", tmpkvar, tmpvvar, pointToVar(t, varname))
	g.genEncodeType(w, f, t.K, tmpkvar, depth+1)
	g.genEncodeType(w, f, t.V, tmpvvar, depth+1)
	fmtFprintf(w, "}")
}
