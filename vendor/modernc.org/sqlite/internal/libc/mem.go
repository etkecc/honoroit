// Copyright 2021 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !libc.membrk && !libc.memgrind && !(linux && (amd64 || arm64 || loong64))

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"modernc.org/memory"
	"modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/sys/types"
)

const memgrind = false

var (
	allocator memory.Allocator
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

	p, err := allocator.UintptrMalloc(int(n))
	if err != nil {
		t.setErrno(errno.ENOMEM)
		return 0
	}

	return p
}

func Xcalloc(t *TLS, n, size types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v size=%v, (%v:)", t, size, origin(2))
	}
	rq := int(n * size)
	if rq == 0 {
		return 0
	}

	allocMu.Lock()

	defer allocMu.Unlock()

	p, err := allocator.UintptrCalloc(int(n * size))
	if err != nil {
		t.setErrno(errno.ENOMEM)
		return 0
	}

	return p
}

func Xrealloc(t *TLS, ptr uintptr, size types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v ptr=%v size=%v, (%v:)", t, ptr, size, origin(2))
	}
	allocMu.Lock()

	defer allocMu.Unlock()

	p, err := allocator.UintptrRealloc(ptr, int(size))
	if err != nil {
		t.setErrno(errno.ENOMEM)
		return 0
	}

	return p
}

func Xfree(t *TLS, p uintptr) {
	if __ccgo_strace {
		trc("t=%v p=%v, (%v:)", t, p, origin(2))
	}
	if p == 0 {
		return
	}

	allocMu.Lock()

	defer allocMu.Unlock()

	allocator.UintptrFree(p)
}

func Xmalloc_usable_size(tls *TLS, p uintptr) (r types.Size_t) {
	if __ccgo_strace {
		trc("tls=%v p=%v, (%v:)", tls, p, origin(2))
		defer func() { trc("-> %v", r) }()
	}
	if p == 0 {
		return 0
	}

	allocMu.Lock()

	defer allocMu.Unlock()

	return types.Size_t(memory.UintptrUsableSize(p))
}

func UsableSize(p uintptr) types.Size_t {
	allocMu.Lock()

	defer allocMu.Unlock()

	return types.Size_t(memory.UintptrUsableSize(p))
}

func MemAuditStart() {}

func MemAuditReport() error { return nil }
