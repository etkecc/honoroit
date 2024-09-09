// Copyright 2021 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !freebsd && !openbsd && !(linux && (amd64 || arm64 || loong64))

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"unsafe"

	"modernc.org/sqlite/internal/libc/pthread"
)

func Xpthread_attr_init(t *TLS, pAttr uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pAttr=%v, (%v:)", t, pAttr, origin(2))
	}
	*(*pthread.Pthread_attr_t)(unsafe.Pointer(pAttr)) = pthread.Pthread_attr_t{}
	return 0
}

func Xpthread_mutex_init(t *TLS, pMutex, pAttr uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pAttr=%v, (%v:)", t, pAttr, origin(2))
	}
	typ := pthread.PTHREAD_MUTEX_DEFAULT
	if pAttr != 0 {
		typ = int(X__ccgo_pthreadMutexattrGettype(t, pAttr))
	}
	mutexesMu.Lock()

	defer mutexesMu.Unlock()

	mutexes[pMutex] = newMutex(typ)
	return 0
}

func Xpthread_atfork(tls *TLS, prepare, parent, child uintptr) int32 {
	return 0
}

func Xpthread_sigmask(tls *TLS, now int32, set, old uintptr) int32 {
	return 0
}
