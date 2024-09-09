// Copyright 2021 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build libc.membrk && !libc.memgrind && !(linux && (amd64 || arm64 || loong64))

// This is a debug-only version of the memory handling functions. When a
// program is built with -tags=libc.membrk a simple but safe version of malloc
// and friends is used that works like sbrk(2). Additionally free becomes a
// nop.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"unsafe"

	"modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/sys/types"
)

const (
	heapAlign = 16
	memgrind  = false
)

var (
	heap     = make([]byte, heapSize)
	heapP    = uintptr(unsafe.Pointer(&heap[heapAlign]))
	heapLast = uintptr(unsafe.Pointer(&heap[heapSize-1]))
)

func Xmalloc(t *TLS, n types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v n=%v, (%v:)", t, n, origin(2))
	}
	if n == 0 {
		return 0
	}

	allocMu.Lock()

	defer allocMu.Unlock()

	n2 := uintptr(n) + uintptrSize
	p := roundup(heapP, 16)
	if p+uintptr(n2) >= heapLast {
		t.setErrno(errno.ENOMEM)
		return 0
	}

	heapP = p + uintptr(n2)
	*(*uintptr)(unsafe.Pointer(p - uintptrSize)) = uintptr(n)
	return p
}

func Xcalloc(t *TLS, n, size types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v size=%v, (%v:)", t, size, origin(2))
	}
	return Xmalloc(t, n*size)
}

func Xrealloc(t *TLS, ptr uintptr, size types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v ptr=%v size=%v, (%v:)", t, ptr, size, origin(2))
	}
	switch {
	case ptr != 0 && size != 0:
		p := Xmalloc(t, size)
		sz0 := UsableSize(ptr)
		if p != 0 {
			copy((*RawMem)(unsafe.Pointer(p))[:size:size], (*RawMem)(unsafe.Pointer(ptr))[:sz0:sz0])
		}
		return p
	case ptr == 0 && size != 0:
		return Xmalloc(t, size)
	}
	return 0
}

func Xfree(t *TLS, p uintptr) {
	if __ccgo_strace {
		trc("t=%v p=%v, (%v:)", t, p, origin(2))
	}
}

func UsableSize(p uintptr) types.Size_t {
	return types.Size_t(*(*uintptr)(unsafe.Pointer(p - uintptrSize)))
}

func MemAuditStart() {}

func MemAuditReport() error { return nil }
