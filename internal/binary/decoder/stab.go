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

package decoder

import (
    `unsafe`

    `github.com/cloudwego/frugal/internal/rt`
)

func mkstab(sw *int, iv int64) (p []int) {
    (*rt.GoSlice)(unsafe.Pointer(&p)).Cap = int(iv)
    (*rt.GoSlice)(unsafe.Pointer(&p)).Len = int(iv)
    (*rt.GoSlice)(unsafe.Pointer(&p)).Ptr = unsafe.Pointer(sw)
    return
}