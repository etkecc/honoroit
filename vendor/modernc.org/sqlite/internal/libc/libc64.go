// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ((amd64 || arm64 || loong64) && !linux) || ppc64le || riscv64 || s390x || mips64le

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"unsafe"

	"modernc.org/sqlite/internal/libc/limits"
	"modernc.org/sqlite/internal/libc/sys/types"
)

const (
	heapSize = 2 << 30
)

type (
	RawMem [1<<50 - 1]byte

	stackHeaderPadding struct {
		a uintptr
	}
)

type bits []int

func newBits(n int) (r bits)  { return make(bits, (n+63)>>6) }
func (b bits) has(n int) bool { return b != nil && b[n>>6]&(1<<uint(n&63)) != 0 }
func (b bits) set(n int)      { b[n>>6] |= 1 << uint(n&63) }

func Xstrchrnul(tls *TLS, s uintptr, c int32) (r uintptr) {
	if __ccgo_strace {
		trc("tls=%v s=%v c=%v, (%v:)", tls, s, c, origin(2))
		defer func() { trc("-> %v", r) }()
	}
	var k types.Size_t
	var w uintptr
	_, _ = k, w
	c = int32(uint8(uint8(c)))
	if !(c != 0) {
		return s + uintptr(Xstrlen(tls, s))
	}
	for {
		if !(uint64(uint64(s))%Uint64FromInt64(8) != 0) {
			break
		}
		if !(*(*int8)(unsafe.Pointer(s)) != 0) || int32(*(*uint8)(unsafe.Pointer(s))) == c {
			return s
		}
		goto _1
	_1:
		s++
	}
	k = uint64(-Int32FromInt32(1)) / Uint64FromInt32(limits.UCHAR_MAX) * uint64(uint64(c))
	w = s
	for {
		if !(!((*(*uint64)(unsafe.Pointer(w))-uint64(-Int32FromInt32(1))/Uint64FromInt32(limits.UCHAR_MAX)) & ^*(*uint64)(unsafe.Pointer(w)) & (uint64(-Int32FromInt32(1))/Uint64FromInt32(limits.UCHAR_MAX)*uint64(Int32FromInt32(limits.UCHAR_MAX)/Int32FromInt32(2)+Int32FromInt32(1))) != 0) && !((*(*uint64)(unsafe.Pointer(w))^k-uint64(-Int32FromInt32(1))/Uint64FromInt32(limits.UCHAR_MAX)) & ^(*(*uint64)(unsafe.Pointer(w))^k) & (uint64(-Int32FromInt32(1))/Uint64FromInt32(limits.UCHAR_MAX)*uint64(Int32FromInt32(limits.UCHAR_MAX)/Int32FromInt32(2)+Int32FromInt32(1))) != 0)) {
			break
		}
		goto _2
	_2:
		w += 8
	}
	s = w
	for {
		if !(*(*int8)(unsafe.Pointer(s)) != 0 && int32(*(*uint8)(unsafe.Pointer(s))) != c) {
			break
		}
		goto _3
	_3:
		s++
	}
	return s
}
