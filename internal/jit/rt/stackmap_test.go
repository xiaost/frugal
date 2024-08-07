/*
 * Copyright 2022 CloudWeGo Authors
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

package rt

import (
	"unsafe"
)

const (
	_FUNCDATA_ArgsPointerMaps   = 0
	_FUNCDATA_LocalsPointerMaps = 1
)

type funcInfo struct {
	fn    unsafe.Pointer
	datap unsafe.Pointer
}

type bitvector struct {
	n        int32 // # of bits
	bytedata *uint8
}

//go:nosplit
func addb(p *byte, n uintptr) *byte {
	return (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + n))
}

func (bv *bitvector) ptrbit(i uintptr) uint8 {
	b := *(addb(bv.bytedata, i/8))
	return (b >> (i % 8)) & 1
}

var keepalive struct {
	s  string
	i  int
	vp unsafe.Pointer
	sb interface{}
	fv uint64
}

func stackmaptestfunc(s string, i int, vp unsafe.Pointer, sb interface{}, fv uint64) (x *uint64, y string, z *int) {
	z = new(int)
	x = new(uint64)
	y = s + "asdf"
	keepalive.s = s
	keepalive.i = i
	keepalive.vp = vp
	keepalive.sb = sb
	keepalive.fv = fv
	return
}
