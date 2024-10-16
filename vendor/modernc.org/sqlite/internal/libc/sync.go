// Copyright 2021 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(linux && (amd64 || arm64 || loong64))

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"sync/atomic"
)

var __sync_synchronize_dummy int32

func X__sync_synchronize(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}

	atomic.StoreInt32(&__sync_synchronize_dummy, atomic.LoadInt32(&__sync_synchronize_dummy)+1)
}
