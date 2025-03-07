/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package reflect

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const appendListFileName = "append_list_gen.go"

func TestGenAppendListCode(t *testing.T) {
	if *gencode {
		genAppendListCode(t, appendListFileName)
		return
	}

	type EnumType int64

	type Msg struct {
		X int64 `frugal:"1,default,i64"`
		Y int64 `frugal:"2,default,i64"`
	}

	type TestStruct struct {
		L1 []int8     `frugal:"1,optional,list<i8>"`
		L2 []int16    `frugal:"2,optional,list<i16>"`
		L3 []int32    `frugal:"3,optional,list<i32>"`
		L4 []int64    `frugal:"4,optional,list<i64>"`
		L5 []EnumType `frugal:"5,optional,list<EnumType>"`
		L6 []string   `frugal:"6,optional,list<string>"`
		L7 []*Msg     `frugal:"7,optional,list<Msg>"`
	}

	var p0, p1 *TestStruct
	var b []byte
	var err error

	p0 = &TestStruct{
		L1: []int8{11, 12, 13},
		L2: []int16{21, 22},
		L3: []int32{31, 32, 33},
		L4: []int64{41, 42},
		L5: []EnumType{51, 52, 53},
		L6: []string{"61", "62"},
		L7: []*Msg{{X: 71, Y: 72}, {X: 73, Y: 74}, {X: 0, Y: 0}},
	}
	b, err = Append(nil, p0)
	require.NoError(t, err)

	p1 = &TestStruct{}
	_, err = Decode(b, p1)
	require.NoError(t, err)
	require.Equal(t, p0, p1)

	// Empty list
	p0 = &TestStruct{
		L1: []int8{},
		L2: []int16{},
		L3: []int32{},
		L4: []int64{},
		L5: []EnumType{},
		L6: []string{},
		L7: []*Msg{},
	}
	b, err = Append(nil, p0)
	require.NoError(t, err)

	p1 = &TestStruct{}
	_, err = Decode(b, p1)
	require.NoError(t, err)
	require.Equal(t, p0, p1)

}

func genAppendListCode(t *testing.T, filename string) {
	f := &bytes.Buffer{}
	f.WriteString(appendListGenFileHeader)
	fm := func(format string, args ...any) {
		fmt.Fprintf(f, format, args...)
		fmt.Fprintln(f)
	}

	var listAppendElementCode = map[ttype]func(){
		tBYTE:   func() { fm("b = append(b, *(*byte)(vp))") },
		tI16:    func() { fm("b = appendUint16(b, *(*uint16)(vp))") },
		tI32:    func() { fm("b = appendUint32(b, *(*uint32)(vp))") },
		tI64:    func() { fm("b = appendUint64(b, *(*uint64)(vp))") },
		tDOUBLE: func() { fm("b = appendUint64(b, *(*uint64)(vp))") },
		tENUM:   func() { fm("b = appendUint32(b, uint32(*(*uint64)(vp)))") },
		tSTRING: func() {
			fm("s := *(*string)(vp)")
			fm("b = appendUint32(b, uint32(len(s)))")
			fm("b = append(b, s...)")
		},
		// tSTRUCT, tMAP, tSET, tLIST -> tOTHER
		tOTHER: func() {
			fm("var err error")
			fm("if t.IsPointer {")
			{
				fm("b, err = t.AppendFunc(t, b, *(*unsafe.Pointer)(vp))")
			}
			fm("} else {")
			{
				fm("b, err = t.AppendFunc(t, b, vp)")
			}
			fm("}")
			fm("if err != nil { return b, err }")
		},
	}

	// func init
	fm("func init() {")
	supportTypes := []ttype{
		tBYTE, tI16, tI32, tI64, tDOUBLE,
		tENUM, tSTRING, tSTRUCT, tMAP, tSET, tLIST,
	}
	t2var := map[ttype]string{
		tBYTE: "tBYTE", tI16: "tI16", tI32: "tI32", tI64: "tI64", tDOUBLE: "tDOUBLE",
		tENUM: "tENUM", tSTRING: "tSTRING",
		tSTRUCT: "tSTRUCT", tMAP: "tMAP", tSET: "tSET", tLIST: "tLIST",
	}
	for _, v := range supportTypes {
		fm("registerListAppendFunc(%s, %s)", t2var[v], appendListFuncName(v))
	}
	fm("}")

	// func appendList_XXX
	for _, v := range []ttype{tBYTE, tI16, tI32, tI64, tENUM, tSTRING, tOTHER} {
		fm("\nfunc %s(t *tType, b []byte, p unsafe.Pointer) ([]byte, error) {",
			appendListFuncName(v))
		fm("t = t.V")
		fm("b, n, vp := appendListHeader(t, b, p)")
		fmt.Fprintln(f, "if n == 0 { return b, nil }")

		fm("for i := uint32(0); i < n; i++ {")
		{
			fm("if i != 0 { vp = unsafe.Add(vp, t.Size) }")
			listAppendElementCode[v]()
		}
		fm("}")
		fm("return b, nil")
		fm("}")
	}

	fileb, err := format.Source(f.Bytes())
	if err != nil {
		t.Log(codeWithLine(f.Bytes()))
		t.Fatal(err)
	}
	err = os.WriteFile(filename, fileb, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("generated: %s", filename)
}

func appendListFuncName(t ttype) string {
	return fmt.Sprintf("appendList_%s", ttype2FuncType(t))
}

const appendListGenFileHeader = `/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package reflect

import "unsafe"

// This File is generated by append_gen.sh. DO NOT EDIT.
// Template and code can be found in append_list_gen_test.go.

`
