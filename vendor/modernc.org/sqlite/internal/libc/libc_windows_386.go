// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"os"
	"strings"
	"syscall"
	gotime "time"
	"unsafe"

	"modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/sys/stat"
	"modernc.org/sqlite/internal/libc/sys/types"
	"modernc.org/sqlite/internal/libc/time"
)

func Xsigaction(t *TLS, signum int32, act, oldact uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v signum=%v oldact=%v, (%v:)", t, signum, oldact, origin(2))
	}
	panic(todo(""))

}

func Xfcntl64(t *TLS, fd, cmd int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v cmd=%v args=%v, (%v:)", t, cmd, args, origin(2))
	}
	panic(todo(""))

}

func Xlstat64(t *TLS, pathname, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v statbuf=%v, (%v:)", t, statbuf, origin(2))
	}
	panic(todo(""))

}

func Xstat64(t *TLS, pathname, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v statbuf=%v, (%v:)", t, statbuf, origin(2))
	}
	panic(todo(""))

}

func Xfstat64(t *TLS, fd int32, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v statbuf=%v, (%v:)", t, fd, statbuf, origin(2))
	}
	panic(todo(""))

}

func Xmremap(t *TLS, old_address uintptr, old_size, new_size types.Size_t, flags int32, args uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v old_address=%v new_size=%v flags=%v args=%v, (%v:)", t, old_address, new_size, flags, args, origin(2))
	}
	panic(todo(""))

}

func Xmmap(t *TLS, addr uintptr, length types.Size_t, prot, flags, fd int32, offset types.Off_t) uintptr {
	if __ccgo_strace {
		trc("t=%v addr=%v length=%v fd=%v offset=%v, (%v:)", t, addr, length, fd, offset, origin(2))
	}
	return Xmmap64(t, addr, length, prot, flags, fd, offset)
}

func Xmmap64(t *TLS, addr uintptr, length types.Size_t, prot, flags, fd int32, offset types.Off_t) uintptr {
	if __ccgo_strace {
		trc("t=%v addr=%v length=%v fd=%v offset=%v, (%v:)", t, addr, length, fd, offset, origin(2))
	}
	panic(todo(""))

}

func Xftruncate64(t *TLS, fd int32, length types.Off_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v length=%v, (%v:)", t, fd, length, origin(2))
	}
	panic(todo(""))

}

func Xlseek64(t *TLS, fd int32, offset types.Off_t, whence int32) types.Off_t {
	if __ccgo_strace {
		trc("t=%v fd=%v offset=%v whence=%v, (%v:)", t, fd, offset, whence, origin(2))
	}

	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	n, err := syscall.Seek(f.Handle, offset, int(whence))
	if err != nil {
		if dmesgs {
			dmesg("%v: fd %v, off %#x, whence %v: %v", origin(1), f._fd, offset, whenceStr(whence), n)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: fd %v, off %#x, whence %v: ok", origin(1), f._fd, offset, whenceStr(whence))
	}
	return n
}

func Xutime(t *TLS, filename, times uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v times=%v, (%v:)", t, times, origin(2))
	}
	panic(todo(""))

}

func Xalarm(t *TLS, seconds uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v seconds=%v, (%v:)", t, seconds, origin(2))
	}
	panic(todo(""))

}

func Xgetrlimit64(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	panic(todo(""))

}

func Xtime(t *TLS, tloc uintptr) types.Time_t {
	if __ccgo_strace {
		trc("t=%v tloc=%v, (%v:)", t, tloc, origin(2))
	}
	panic(todo(""))

}

func Xmkdir(t *TLS, path uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v mode=%v, (%v:)", t, path, mode, origin(2))
	}
	panic(todo(""))

}

func Xsymlink(t *TLS, target, linkpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v linkpath=%v, (%v:)", t, linkpath, origin(2))
	}
	panic(todo(""))

}

func Xutimes(t *TLS, filename, times uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v times=%v, (%v:)", t, times, origin(2))
	}
	panic(todo(""))

}

func Xunlink(t *TLS, pathname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v, (%v:)", t, pathname, origin(2))
	}
	err := syscall.DeleteFile((*uint16)(unsafe.Pointer(pathname)))
	if err != nil {
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %q: ok", origin(1), GoString(pathname))
	}

	return 0

}

func Xrmdir(t *TLS, pathname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v, (%v:)", t, pathname, origin(2))
	}
	panic(todo(""))

}

func Xmknod(t *TLS, pathname uintptr, mode types.Mode_t, dev types.Dev_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v dev=%v, (%v:)", t, pathname, mode, dev, origin(2))
	}
	panic(todo(""))

}

func Xlink(t *TLS, oldpath, newpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v newpath=%v, (%v:)", t, newpath, origin(2))
	}
	panic(todo(""))

}

func Xpipe(t *TLS, pipefd uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pipefd=%v, (%v:)", t, pipefd, origin(2))
	}
	panic(todo(""))

}

func Xdup2(t *TLS, oldfd, newfd int32) int32 {
	if __ccgo_strace {
		trc("t=%v newfd=%v, (%v:)", t, newfd, origin(2))
	}
	panic(todo(""))

}

func Xreadlink(t *TLS, path, buf uintptr, bufsize types.Size_t) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v buf=%v bufsize=%v, (%v:)", t, buf, bufsize, origin(2))
	}
	panic(todo(""))

}

func Xfopen64(t *TLS, pathname, mode uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v mode=%v, (%v:)", t, mode, origin(2))
	}

	m := strings.ReplaceAll(GoString(mode), "b", "")
	var flags int
	switch m {
	case "r":
		flags = os.O_RDONLY
	case "r+":
		flags = os.O_RDWR
	case "w":
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "w+":
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a":
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "a+":
		flags = os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		panic(m)
	}

	h, err := syscall.Open(GoString(pathname), int(flags), uint32(0666))
	if err != nil {
		t.setErrno(err)
		return 0
	}

	p, _ := wrapFdHandle(h)
	if p != 0 {
		return p
	}
	_ = syscall.Close(h)
	t.setErrno(errno.ENOMEM)
	return 0
}

func Xrecv(t *TLS, sockfd uint32, buf uintptr, len, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v buf=%v flags=%v, (%v:)", t, sockfd, buf, flags, origin(2))
	}
	panic(todo(""))
}

func Xsend(t *TLS, sockfd uint32, buf uintptr, len, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v buf=%v flags=%v, (%v:)", t, sockfd, buf, flags, origin(2))
	}
	panic(todo(""))
}

func Xshutdown(t *TLS, sockfd uint32, how int32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v how=%v, (%v:)", t, sockfd, how, origin(2))
	}
	panic(todo(""))

}

func Xgetpeername(t *TLS, sockfd uint32, addr uintptr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))
}

func Xgetsockname(t *TLS, sockfd uint32, addr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addrlen=%v, (%v:)", t, sockfd, addrlen, origin(2))
	}
	panic(todo(""))
}

func Xsocket(t *TLS, domain, type1, protocol int32) uint32 {
	if __ccgo_strace {
		trc("t=%v protocol=%v, (%v:)", t, protocol, origin(2))
	}
	panic(todo(""))
}

func Xbind(t *TLS, sockfd uint32, addr uintptr, addrlen int32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))
}

func Xconnect(t *TLS, sockfd uint32, addr uintptr, addrlen int32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))
}

func Xlisten(t *TLS, sockfd uint32, backlog int32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v backlog=%v, (%v:)", t, sockfd, backlog, origin(2))
	}
	panic(todo(""))
}

func Xaccept(t *TLS, sockfd uint32, addr uintptr, addrlen uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))
}

func X_localtime32(_ *TLS, sourceTime uintptr) uintptr {
	loc := getLocalLocation()
	ut := *(*time.Time_t)(unsafe.Pointer(sourceTime))
	t := gotime.Unix(int64(ut), 0).In(loc)
	localtime.Ftm_sec = int32(t.Second())
	localtime.Ftm_min = int32(t.Minute())
	localtime.Ftm_hour = int32(t.Hour())
	localtime.Ftm_mday = int32(t.Day())
	localtime.Ftm_mon = int32(t.Month() - 1)
	localtime.Ftm_year = int32(t.Year() - 1900)
	localtime.Ftm_wday = int32(t.Weekday())
	localtime.Ftm_yday = int32(t.YearDay())
	localtime.Ftm_isdst = Bool32(isTimeDST(t))
	return uintptr(unsafe.Pointer(&localtime))
}

func X_gmtime32(t *TLS, sourceTime uintptr) uintptr {
	r0, _, err := syscall.SyscallN(procGmtime32.Addr(), uintptr(sourceTime))
	if err != 0 {
		t.setErrno(err)
	}
	return uintptr(r0)
}

func XSetWindowLongW(t *TLS, hwnd uintptr, nIndex int32, dwNewLong long) long {
	if __ccgo_strace {
		trc("t=%v hwnd=%v nIndex=%v dwNewLong=%v, (%v:)", t, hwnd, nIndex, dwNewLong, origin(2))
	}
	panic(todo(""))
}

func XGetWindowLongW(t *TLS, hwnd uintptr, nIndex int32) long {
	if __ccgo_strace {
		trc("t=%v hwnd=%v nIndex=%v, (%v:)", t, hwnd, nIndex, origin(2))
	}
	panic(todo(""))
}

func XDefWindowProcW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSendMessageTimeoutW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X_fstat(t *TLS, fd int32, buffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v buffer=%v, (%v:)", t, fd, buffer, origin(2))
	}
	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(EBADF)
		return -1
	}

	var d syscall.ByHandleFileInformation
	err := syscall.GetFileInformationByHandle(f.Handle, &d)
	if err != nil {
		t.setErrno(EBADF)
		return -1
	}

	var bStat32 = (*stat.X_stat32)(unsafe.Pointer(buffer))
	var accessTime = int64(d.LastAccessTime.HighDateTime)<<32 + int64(d.LastAccessTime.LowDateTime)
	bStat32.Fst_atime = int32(WindowsTickToUnixSeconds(accessTime))
	var modTime = int64(d.LastWriteTime.HighDateTime)<<32 + int64(d.LastWriteTime.LowDateTime)
	bStat32.Fst_mtime = int32(WindowsTickToUnixSeconds(modTime))
	var crTime = int64(d.CreationTime.HighDateTime)<<32 + int64(d.CreationTime.LowDateTime)
	bStat32.Fst_ctime = int32(WindowsTickToUnixSeconds(crTime))
	var fSz = int64(d.FileSizeHigh)<<32 + int64(d.FileSizeLow)
	bStat32.Fst_size = int32(fSz)
	bStat32.Fst_mode = WindowsAttrbiutesToStat(d.FileAttributes)

	return 0
}

func Xstrspn(tls *TLS, s uintptr, c uintptr) size_t {
	if __ccgo_strace {
		trc("tls=%v s=%v c=%v, (%v:)", tls, s, c, origin(2))
	}
	bp := tls.Alloc(32)
	defer tls.Free(32)

	var a uintptr = s
	*(*[8]size_t)(unsafe.Pointer(bp)) = [8]size_t{0: size_t(0)}

	if !(int32(*(*int8)(unsafe.Pointer(c))) != 0) {
		return size_t(0)
	}
	if !(int32(*(*int8)(unsafe.Pointer(c + 1))) != 0) {
		for ; int32(*(*int8)(unsafe.Pointer(s))) == int32(*(*int8)(unsafe.Pointer(c))); s++ {
		}
		return size_t((int32(s) - int32(a)) / 1)
	}

	for ; *(*int8)(unsafe.Pointer(c)) != 0 && AssignOrPtrUint32(bp+uintptr(size_t(*(*uint8)(unsafe.Pointer(c)))/(uint32(8)*uint32(unsafe.Sizeof(size_t(0)))))*4, size_t(size_t(1))<<(size_t(*(*uint8)(unsafe.Pointer(c)))%(uint32(8)*uint32(unsafe.Sizeof(size_t(0)))))) != 0; c++ {
	}
	for ; *(*int8)(unsafe.Pointer(s)) != 0 && *(*size_t)(unsafe.Pointer(bp + uintptr(size_t(*(*uint8)(unsafe.Pointer(s)))/(uint32(8)*uint32(unsafe.Sizeof(size_t(0)))))*4))&(size_t(size_t(1))<<(size_t(*(*uint8)(unsafe.Pointer(s)))%(uint32(8)*uint32(unsafe.Sizeof(size_t(0)))))) != 0; s++ {
	}
	return size_t((int32(s) - int32(a)) / 1)
}
