// +build !go1.17

/*
 * Copyright 2022 ByteDance Inc.
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

var _S_getg = []byte {
    0x64, 0x48, 0x8b, 0x0c, 0x25, 0xf8, 0xff, 0xff, 0xff,   // MOVQ %fs:-8,%rcx
}