/*
 * Copyright 2021 ByteDance Inc.
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

package atm

import (
    `fmt`
)

type Register interface {
    A() Argument
}

type (
    Argument        uint8
    GenericRegister uint8
    PointerRegister uint8
)

const (
    ArgMask    = 0x7f
    ArgGeneric = 0x00
    ArgPointer = 0x80
)

const (
    R0 GenericRegister = iota
    R1
    R2
    R3
    R4
    R5
    R6
    R7
    Rz
)

const (
    P0 PointerRegister = iota
    P1
    P2
    P3
    P4
    P5
    P6
    P7
    Pn
)

var _GR_Names = [...]string {
    R0: "r0",
    R1: "r1",
    R2: "r2",
    R3: "r3",
    R4: "r4",
    R5: "r5",
    R6: "r6",
    R7: "r7",
    Rz: "z",
}

var _PR_Names = [...]string {
    P0: "p0",
    P1: "p1",
    P2: "p2",
    P3: "p3",
    P4: "p4",
    P5: "p5",
    P6: "p6",
    P7: "p7",
    Pn: "nil",
}

func (self GenericRegister) A() Argument { return Argument(self) | ArgGeneric }
func (self PointerRegister) A() Argument { return Argument(self) | ArgPointer }

func (self GenericRegister) String() string {
    if v := _GR_Names[self]; v == "" {
        panic(fmt.Sprintf("invalid generic register: 0x%02x", uint8(self)))
    } else {
        return v
    }
}

func (self PointerRegister) String() string {
    if v := _PR_Names[self]; v == "" {
        panic(fmt.Sprintf("invalid pointer register: 0x%02x", uint8(self)))
    } else {
        return v
    }
}
