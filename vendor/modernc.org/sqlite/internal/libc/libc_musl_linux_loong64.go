// Copyright 2023 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"golang.org/x/sys/unix"
)

type long = int64

type ulong = uint64

type RawMem [1<<50 - 1]byte

func Xrenameat2(t *TLS, olddirfd int32, oldpath uintptr, newdirfd int32, newpath uintptr, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v olddirfd=%v oldpath=%v newdirfd=%v newpath=%v flags=%v, (%v:)", t, olddirfd, oldpath, newdirfd, newpath, flags, origin(2))
	}
	if _, _, err := unix.Syscall6(unix.SYS_RENAMEAT2, uintptr(olddirfd), oldpath, uintptr(newdirfd), newpath, uintptr(flags), 0); err != 0 {
		t.setErrno(int32(err))
		return -1
	}

	return 0
}
