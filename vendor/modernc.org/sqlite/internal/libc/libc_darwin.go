// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	gosignal "os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	gotime "time"
	"unicode"
	"unsafe"

	guuid "github.com/google/uuid"
	"golang.org/x/sys/unix"
	"modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/fcntl"
	"modernc.org/sqlite/internal/libc/fts"
	gonetdb "modernc.org/sqlite/internal/libc/honnef.co/go/netdb"
	"modernc.org/sqlite/internal/libc/langinfo"
	"modernc.org/sqlite/internal/libc/limits"
	"modernc.org/sqlite/internal/libc/netdb"
	"modernc.org/sqlite/internal/libc/netinet/in"
	"modernc.org/sqlite/internal/libc/signal"
	"modernc.org/sqlite/internal/libc/stdio"
	"modernc.org/sqlite/internal/libc/sys/socket"
	"modernc.org/sqlite/internal/libc/sys/stat"
	"modernc.org/sqlite/internal/libc/sys/types"
	"modernc.org/sqlite/internal/libc/termios"
	"modernc.org/sqlite/internal/libc/time"
	"modernc.org/sqlite/internal/libc/unistd"
	"modernc.org/sqlite/internal/libc/uuid/uuid"
	"modernc.org/sqlite/internal/libc/wctype"
)

const (
	maxPathLen = 1024
)

type (
	long  = types.User_long_t
	ulong = types.User_ulong_t
)

type pthreadAttr struct {
	detachState int32
}

var X__stderrp = Xstdout
var X__stdinp = Xstdin
var X__stdoutp = Xstdout

var X__mb_cur_max int32 = 1

var startTime = gotime.Now()

type file uintptr

func (f file) fd() int32      { return int32((*stdio.FILE)(unsafe.Pointer(f)).F_file) }
func (f file) setFd(fd int32) { (*stdio.FILE)(unsafe.Pointer(f)).F_file = int16(fd) }

func (f file) err() bool {
	return (*stdio.FILE)(unsafe.Pointer(f)).F_flags&1 != 0
}

func (f file) setErr() {
	(*stdio.FILE)(unsafe.Pointer(f)).F_flags |= 1
}

func (f file) close(t *TLS) int32 {
	r := Xclose(t, f.fd())
	Xfree(t, uintptr(f))
	if r < 0 {
		return stdio.EOF
	}

	return 0
}

func newFile(t *TLS, fd int32) uintptr {
	p := Xcalloc(t, 1, types.Size_t(unsafe.Sizeof(stdio.FILE{})))
	if p == 0 {
		return 0
	}

	file(p).setFd(fd)
	return p
}

func fwrite(fd int32, b []byte) (int, error) {
	if fd == unistd.STDOUT_FILENO {
		return write(b)
	}

	if dmesgs {
		dmesg("%v: fd %v: %s", origin(1), fd, hex.Dump(b))
	}
	return unix.Write(int(fd), b)
}

func X__inline_isnand(t *TLS, x float64) int32 {
	if __ccgo_strace {
		trc("t=%v x=%v, (%v:)", t, x, origin(2))
	}
	return Xisnan(t, x)
}

func X__inline_isnanf(t *TLS, x float32) int32 {
	if __ccgo_strace {
		trc("t=%v x=%v, (%v:)", t, x, origin(2))
	}
	return Xisnanf(t, x)
}

func X__inline_isnanl(t *TLS, x float64) int32 {
	if __ccgo_strace {
		trc("t=%v x=%v, (%v:)", t, x, origin(2))
	}
	return Xisnan(t, x)
}

func Xfprintf(t *TLS, stream, format, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v args=%v, (%v:)", t, args, origin(2))
	}
	n, _ := fwrite(int32((*stdio.FILE)(unsafe.Pointer(stream)).F_file), printf(format, args))
	return int32(n)
}

func Xusleep(t *TLS, usec types.Useconds_t) int32 {
	if __ccgo_strace {
		trc("t=%v usec=%v, (%v:)", t, usec, origin(2))
	}
	gotime.Sleep(gotime.Microsecond * gotime.Duration(usec))
	return 0
}

func Xfutimes(t *TLS, fd int32, tv uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v tv=%v, (%v:)", t, fd, tv, origin(2))
	}
	var a []unix.Timeval
	if tv != 0 {
		a = make([]unix.Timeval, 2)
		a[0] = *(*unix.Timeval)(unsafe.Pointer(tv))
		a[1] = *(*unix.Timeval)(unsafe.Pointer(tv + unsafe.Sizeof(unix.Timeval{})))
	}
	if err := unix.Futimes(int(fd), a); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xsrandomdev(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func Xgethostuuid(t *TLS, id uintptr, wait uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v id=%v wait=%v, (%v:)", t, id, wait, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_GETHOSTUUID, id, wait, 0); err != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xflock(t *TLS, fd, operation int32) int32 {
	if __ccgo_strace {
		trc("t=%v operation=%v, (%v:)", t, operation, origin(2))
	}
	if err := unix.Flock(int(fd), int(operation)); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xfsctl(t *TLS, path uintptr, request ulong, data uintptr, options uint32) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v request=%v data=%v options=%v, (%v:)", t, path, request, data, options, origin(2))
	}
	panic(todo(""))

}

func X__error(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return t.errnop
}

func Xisspace(t *TLS, c int32) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v, (%v:)", t, c, origin(2))
	}
	return __isspace(t, c)
}

func X__assert_rtn(t *TLS, function, file uintptr, line int32, assertion uintptr) {
	if __ccgo_strace {
		trc("t=%v file=%v line=%v assertion=%v, (%v:)", t, file, line, assertion, origin(2))
	}
	panic(todo(""))

}

func Xgetrusage(t *TLS, who int32, usage uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v who=%v usage=%v, (%v:)", t, who, usage, origin(2))
	}
	panic(todo(""))

}

func Xfgetc(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	fd := int((*stdio.FILE)(unsafe.Pointer(stream)).F_file)
	var buf [1]byte
	if n, _ := unix.Read(fd, buf[:]); n != 0 {
		return int32(buf[0])
	}

	return stdio.EOF
}

func Xlstat(t *TLS, pathname, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v statbuf=%v, (%v:)", t, statbuf, origin(2))
	}
	return Xlstat64(t, pathname, statbuf)
}

func Xstat(t *TLS, pathname, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v statbuf=%v, (%v:)", t, statbuf, origin(2))
	}
	return Xstat64(t, pathname, statbuf)
}

func Xchdir(t *TLS, path uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v, (%v:)", t, path, origin(2))
	}
	if err := unix.Chdir(GoString(path)); err != nil {
		if dmesgs {
			dmesg("%v: %q: %v FAIL", origin(1), GoString(path), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %q: ok", origin(1), GoString(path))
	}
	return 0
}

var localtime time.Tm

func Xlocaltime(_ *TLS, timep uintptr) uintptr {
	loc := getLocalLocation()
	ut := *(*time.Time_t)(unsafe.Pointer(timep))
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

func Xlocaltime_r(_ *TLS, timep, result uintptr) uintptr {
	loc := getLocalLocation()
	ut := *(*time_t)(unsafe.Pointer(timep))
	t := gotime.Unix(int64(ut), 0).In(loc)
	(*time.Tm)(unsafe.Pointer(result)).Ftm_sec = int32(t.Second())
	(*time.Tm)(unsafe.Pointer(result)).Ftm_min = int32(t.Minute())
	(*time.Tm)(unsafe.Pointer(result)).Ftm_hour = int32(t.Hour())
	(*time.Tm)(unsafe.Pointer(result)).Ftm_mday = int32(t.Day())
	(*time.Tm)(unsafe.Pointer(result)).Ftm_mon = int32(t.Month() - 1)
	(*time.Tm)(unsafe.Pointer(result)).Ftm_year = int32(t.Year() - 1900)
	(*time.Tm)(unsafe.Pointer(result)).Ftm_wday = int32(t.Weekday())
	(*time.Tm)(unsafe.Pointer(result)).Ftm_yday = int32(t.YearDay())
	(*time.Tm)(unsafe.Pointer(result)).Ftm_isdst = Bool32(isTimeDST(t))
	return result
}

func Xopen(t *TLS, pathname uintptr, flags int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v args=%v, (%v:)", t, pathname, flags, args, origin(2))
	}
	var mode types.Mode_t
	if args != 0 {
		mode = (types.Mode_t)(VaUint32(&args))
	}
	fd, err := unix.Open(GoString(pathname), int(flags), uint32(mode))
	if err != nil {
		if dmesgs {
			dmesg("%v: %q %#x %#o: %v FAIL", origin(1), GoString(pathname), flags, mode, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %q flags %#x mode %#o: fd %v", origin(1), GoString(pathname), flags, mode, fd)
	}
	return int32(fd)
}

func Xlseek(t *TLS, fd int32, offset types.Off_t, whence int32) types.Off_t {
	if __ccgo_strace {
		trc("t=%v fd=%v offset=%v whence=%v, (%v:)", t, fd, offset, whence, origin(2))
	}
	return types.Off_t(Xlseek64(t, fd, offset, whence))
}

func whenceStr(whence int32) string {
	switch whence {
	case fcntl.SEEK_CUR:
		return "SEEK_CUR"
	case fcntl.SEEK_END:
		return "SEEK_END"
	case fcntl.SEEK_SET:
		return "SEEK_SET"
	default:
		return fmt.Sprintf("whence(%d)", whence)
	}
}

var fsyncStatbuf stat.Stat

func Xfsync(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}
	if noFsync {
		return Xfstat(t, fd, uintptr(unsafe.Pointer(&fsyncStatbuf)))
	}

	if err := unix.Fsync(int(fd)); err != nil {
		if dmesgs {
			dmesg("%v: %v: %v FAIL", origin(1), fd, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %d: ok", origin(1), fd)
	}
	return 0
}

func Xsysconf(t *TLS, name int32) long {
	if __ccgo_strace {
		trc("t=%v name=%v, (%v:)", t, name, origin(2))
	}
	switch name {
	case unistd.X_SC_PAGESIZE:
		return long(unix.Getpagesize())
	case unistd.X_SC_NPROCESSORS_ONLN:
		return long(runtime.NumCPU())
	case unistd.X_SC_GETPW_R_SIZE_MAX:
		return 128
	}

	panic(todo("", name))
}

func Xclose(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}
	if err := unix.Close(int(fd)); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %d: ok", origin(1), fd)
	}
	return 0
}

func Xgetcwd(t *TLS, buf uintptr, size types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v buf=%v size=%v, (%v:)", t, buf, size, origin(2))
	}
	if _, err := unix.Getcwd((*RawMem)(unsafe.Pointer(buf))[:size:size]); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return 0
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return buf
}

func Xfstat(t *TLS, fd int32, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v statbuf=%v, (%v:)", t, fd, statbuf, origin(2))
	}
	return Xfstat64(t, fd, statbuf)
}

func Xftruncate(t *TLS, fd int32, length types.Off_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v length=%v, (%v:)", t, fd, length, origin(2))
	}
	if err := unix.Ftruncate(int(fd), int64(length)); err != nil {
		if dmesgs {
			dmesg("%v: fd %d: %v FAIL", origin(1), fd, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %d %#x: ok", origin(1), fd, length)
	}
	return 0
}

func Xfcntl(t *TLS, fd, cmd int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v cmd=%v args=%v, (%v:)", t, cmd, args, origin(2))
	}
	return Xfcntl64(t, fd, cmd, args)
}

func Xread(t *TLS, fd int32, buf uintptr, count types.Size_t) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	var n int
	var err error
	switch {
	case count == 0:
		n, err = unix.Read(int(fd), nil)
	default:
		n, err = unix.Read(int(fd), (*RawMem)(unsafe.Pointer(buf))[:count:count])
		if dmesgs && err == nil {
			dmesg("%v: fd %v, count %#x, n %#x\n%s", origin(1), fd, count, n, hex.Dump((*RawMem)(unsafe.Pointer(buf))[:n:n]))
		}
	}
	if err != nil {
		if dmesgs {
			dmesg("%v: fd %v, %v FAIL", origin(1), fd, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return types.Ssize_t(n)
}

func Xwrite(t *TLS, fd int32, buf uintptr, count types.Size_t) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	var n int
	var err error
	switch {
	case count == 0:
		n, err = unix.Write(int(fd), nil)
	default:
		n, err = unix.Write(int(fd), (*RawMem)(unsafe.Pointer(buf))[:count:count])
		if dmesgs {
			dmesg("%v: fd %v, count %#x\n%s", origin(1), fd, count, hex.Dump((*RawMem)(unsafe.Pointer(buf))[:count:count]))
		}
	}
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return types.Ssize_t(n)
}

func Xfchmod(t *TLS, fd int32, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v mode=%v, (%v:)", t, fd, mode, origin(2))
	}
	if err := unix.Fchmod(int(fd), uint32(mode)); err != nil {
		if dmesgs {
			dmesg("%v: %d %#o: %v FAIL", origin(1), fd, mode, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %d %#o: ok", origin(1), fd, mode)
	}
	return 0
}

func Xfchown(t *TLS, fd int32, owner types.Uid_t, group types.Gid_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v owner=%v group=%v, (%v:)", t, fd, owner, group, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_FCHOWN, uintptr(fd), uintptr(owner), uintptr(group)); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xgeteuid(t *TLS) types.Uid_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r := types.Uid_t(unix.Geteuid())
	if dmesgs {
		dmesg("%v: %v", origin(1), r)
	}
	return r
}

func Xmmap(t *TLS, addr uintptr, length types.Size_t, prot, flags, fd int32, offset types.Off_t) uintptr {
	if __ccgo_strace {
		trc("t=%v addr=%v length=%v fd=%v offset=%v, (%v:)", t, addr, length, fd, offset, origin(2))
	}

	data, _, err := unix.Syscall6(unix.SYS_MMAP, addr, uintptr(length), uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset))
	if err != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return ^uintptr(0)
	}

	if dmesgs {
		dmesg("%v: %#x", origin(1), data)
	}
	return data
}

func Xmunmap(t *TLS, addr uintptr, length types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v addr=%v length=%v, (%v:)", t, addr, length, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_MUNMAP, addr, uintptr(length), 0); err != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xgettimeofday(t *TLS, tv, tz uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v tz=%v, (%v:)", t, tz, origin(2))
	}
	if tz != 0 {
		panic(todo(""))
	}

	var tvs unix.Timeval
	err := unix.Gettimeofday(&tvs)
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	*(*unix.Timeval)(unsafe.Pointer(tv)) = tvs
	return 0
}

func Xgetsockopt(t *TLS, sockfd, level, optname int32, optval, optlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v optname=%v optlen=%v, (%v:)", t, optname, optlen, origin(2))
	}
	if _, _, err := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(sockfd), uintptr(level), uintptr(optname), optval, optlen, 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xsetsockopt(t *TLS, sockfd, level, optname int32, optval uintptr, optlen socket.Socklen_t) int32 {
	if __ccgo_strace {
		trc("t=%v optname=%v optval=%v optlen=%v, (%v:)", t, optname, optval, optlen, origin(2))
	}
	if _, _, err := unix.Syscall6(unix.SYS_SETSOCKOPT, uintptr(sockfd), uintptr(level), uintptr(optname), optval, uintptr(optlen), 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xioctl(t *TLS, fd int32, request ulong, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v request=%v va=%v, (%v:)", t, fd, request, va, origin(2))
	}
	var argp uintptr
	if va != 0 {
		argp = VaUintptr(&va)
	}
	n, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(request), argp)
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return int32(n)
}

func Xgetsockname(t *TLS, sockfd int32, addr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addrlen=%v, (%v:)", t, sockfd, addrlen, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_GETSOCKNAME, uintptr(sockfd), addr, addrlen); err != 0 {
		if dmesgs {
			dmesg("%v: fd %v: %v FAIL", origin(1), sockfd, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: fd %v: ok", origin(1), sockfd)
	}
	return 0
}

func Xselect(t *TLS, nfds int32, readfds, writefds, exceptfds, timeout uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v nfds=%v timeout=%v, (%v:)", t, nfds, timeout, origin(2))
	}
	n, err := unix.Select(
		int(nfds),
		(*unix.FdSet)(unsafe.Pointer(readfds)),
		(*unix.FdSet)(unsafe.Pointer(writefds)),
		(*unix.FdSet)(unsafe.Pointer(exceptfds)),
		(*unix.Timeval)(unsafe.Pointer(timeout)),
	)
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return int32(n)
}

func Xmkfifo(t *TLS, pathname uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	if err := unix.Mkfifo(GoString(pathname), uint32(mode)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xumask(t *TLS, mask types.Mode_t) types.Mode_t {
	if __ccgo_strace {
		trc("t=%v mask=%v, (%v:)", t, mask, origin(2))
	}
	return types.Mode_t(unix.Umask(int(mask)))
}

func Xwaitpid(t *TLS, pid types.Pid_t, wstatus uintptr, optname int32) types.Pid_t {
	if __ccgo_strace {
		trc("t=%v pid=%v wstatus=%v optname=%v, (%v:)", t, pid, wstatus, optname, origin(2))
	}
	n, err := unix.Wait4(int(pid), (*unix.WaitStatus)(unsafe.Pointer(wstatus)), int(optname), nil)
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return types.Pid_t(n)
}

func Xuname(t *TLS, buf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v buf=%v, (%v:)", t, buf, origin(2))
	}
	if err := unix.Uname((*unix.Utsname)(unsafe.Pointer(buf))); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xrecv(t *TLS, sockfd int32, buf uintptr, len types.Size_t, flags int32) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v sockfd=%v buf=%v len=%v flags=%v, (%v:)", t, sockfd, buf, len, flags, origin(2))
	}
	n, _, err := unix.Syscall6(unix.SYS_RECVFROM, uintptr(sockfd), buf, uintptr(len), uintptr(flags), 0, 0)
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return types.Ssize_t(n)
}

func Xsend(t *TLS, sockfd int32, buf uintptr, len types.Size_t, flags int32) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v sockfd=%v buf=%v len=%v flags=%v, (%v:)", t, sockfd, buf, len, flags, origin(2))
	}
	n, _, err := unix.Syscall6(unix.SYS_SENDTO, uintptr(sockfd), buf, uintptr(len), uintptr(flags), 0, 0)
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return types.Ssize_t(n)
}

func Xshutdown(t *TLS, sockfd, how int32) int32 {
	if __ccgo_strace {
		trc("t=%v how=%v, (%v:)", t, how, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_SHUTDOWN, uintptr(sockfd), uintptr(how), 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xgetpeername(t *TLS, sockfd int32, addr uintptr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_GETPEERNAME, uintptr(sockfd), addr, uintptr(addrlen)); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xsocket(t *TLS, domain, type1, protocol int32) int32 {
	if __ccgo_strace {
		trc("t=%v protocol=%v, (%v:)", t, protocol, origin(2))
	}
	n, _, err := unix.Syscall(unix.SYS_SOCKET, uintptr(domain), uintptr(type1), uintptr(protocol))
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return int32(n)
}

func Xbind(t *TLS, sockfd int32, addr uintptr, addrlen uint32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	n, _, err := unix.Syscall(unix.SYS_BIND, uintptr(sockfd), addr, uintptr(addrlen))
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return int32(n)
}

func Xconnect(t *TLS, sockfd int32, addr uintptr, addrlen uint32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_CONNECT, uintptr(sockfd), addr, uintptr(addrlen)); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xlisten(t *TLS, sockfd, backlog int32) int32 {
	if __ccgo_strace {
		trc("t=%v backlog=%v, (%v:)", t, backlog, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_LISTEN, uintptr(sockfd), uintptr(backlog), 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xaccept(t *TLS, sockfd int32, addr uintptr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))

}

func Xgetuid(t *TLS) types.Uid_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r := types.Uid_t(os.Getuid())
	if dmesgs {
		dmesg("%v: %v", origin(1), r)
	}
	return r
}

func Xgetpid(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r := int32(os.Getpid())
	if dmesgs {
		dmesg("%v: %v", origin(1), r)
	}
	return r
}

func Xsystem(t *TLS, command uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v command=%v, (%v:)", t, command, origin(2))
	}
	s := GoString(command)
	if command == 0 {
		panic(todo(""))
	}

	cmd := exec.Command("sh", "-c", s)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ps := err.(*exec.ExitError)
		return int32(ps.ExitCode())
	}

	return 0
}

func Xsetvbuf(t *TLS, stream, buf uintptr, mode int32, size types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v buf=%v mode=%v size=%v, (%v:)", t, buf, mode, size, origin(2))
	}
	return 0
}

func Xraise(t *TLS, sig int32) int32 {
	if __ccgo_strace {
		trc("t=%v sig=%v, (%v:)", t, sig, origin(2))
	}
	panic(todo(""))
}

func Xfileno(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	if stream == 0 {
		if dmesgs {
			dmesg("%v: FAIL", origin(1))
		}
		t.setErrno(errno.EBADF)
		return -1
	}

	if fd := int32((*stdio.FILE)(unsafe.Pointer(stream)).F_file); fd >= 0 {
		return fd
	}

	if dmesgs {
		dmesg("%v: FAIL", origin(1))
	}
	t.setErrno(errno.EBADF)
	return -1
}

func newFtsent(t *TLS, info int, path string, stat *unix.Stat_t, err syscall.Errno) (r *fts.FTSENT) {
	var statp uintptr
	if stat != nil {
		statp = Xmalloc(t, types.Size_t(unsafe.Sizeof(unix.Stat_t{})))
		if statp == 0 {
			panic("OOM")
		}

		*(*unix.Stat_t)(unsafe.Pointer(statp)) = *stat
	}
	csp, errx := CString(path)
	if errx != nil {
		panic("OOM")
	}

	return &fts.FTSENT{
		Ffts_info:    uint16(info),
		Ffts_path:    csp,
		Ffts_pathlen: uint16(len(path)),
		Ffts_statp:   statp,
		Ffts_errno:   int32(err),
	}
}

func newCFtsent(t *TLS, info int, path string, stat *unix.Stat_t, err syscall.Errno) uintptr {
	p := Xcalloc(t, 1, types.Size_t(unsafe.Sizeof(fts.FTSENT{})))
	if p == 0 {
		panic("OOM")
	}

	*(*fts.FTSENT)(unsafe.Pointer(p)) = *newFtsent(t, info, path, stat, err)
	return p
}

func ftsentClose(t *TLS, p uintptr) {
	Xfree(t, (*fts.FTSENT)(unsafe.Pointer(p)).Ffts_path)
	Xfree(t, (*fts.FTSENT)(unsafe.Pointer(p)).Ffts_statp)
}

type ftstream struct {
	s []uintptr
	x int
}

func (f *ftstream) close(t *TLS) {
	for _, p := range f.s {
		ftsentClose(t, p)
		Xfree(t, p)
	}
	*f = ftstream{}
}

func Xfts_open(t *TLS, path_argv uintptr, options int32, compar uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v path_argv=%v options=%v compar=%v, (%v:)", t, path_argv, options, compar, origin(2))
	}
	f := &ftstream{}

	var walk func(string)
	walk = func(path string) {
		var fi os.FileInfo
		var err error
		switch {
		case options&fts.FTS_LOGICAL != 0:
			fi, err = os.Stat(path)
		case options&fts.FTS_PHYSICAL != 0:
			fi, err = os.Lstat(path)
		default:
			panic(todo(""))
		}

		if err != nil {
			return
		}

		var statp *unix.Stat_t
		if options&fts.FTS_NOSTAT == 0 {
			var stat unix.Stat_t
			switch {
			case options&fts.FTS_LOGICAL != 0:
				if err := unix.Stat(path, &stat); err != nil {
					panic(todo(""))
				}
			case options&fts.FTS_PHYSICAL != 0:
				if err := unix.Lstat(path, &stat); err != nil {
					panic(todo(""))
				}
			default:
				panic(todo(""))
			}

			statp = &stat
		}

	out:
		switch {
		case fi.IsDir():
			f.s = append(f.s, newCFtsent(t, fts.FTS_D, path, statp, 0))
			g, err := os.Open(path)
			switch x := err.(type) {
			case nil:
			case *os.PathError:
				f.s = append(f.s, newCFtsent(t, fts.FTS_DNR, path, statp, errno.EACCES))
				break out
			default:
				panic(todo("%q: %v %T", path, x, x))
			}

			names, err := g.Readdirnames(-1)
			g.Close()
			if err != nil {
				panic(todo(""))
			}

			for _, name := range names {
				walk(path + "/" + name)
				if f == nil {
					break out
				}
			}

			f.s = append(f.s, newCFtsent(t, fts.FTS_DP, path, statp, 0))
		default:
			info := fts.FTS_F
			if fi.Mode()&os.ModeSymlink != 0 {
				info = fts.FTS_SL
			}
			switch {
			case statp != nil:
				f.s = append(f.s, newCFtsent(t, info, path, statp, 0))
			case options&fts.FTS_NOSTAT != 0:
				f.s = append(f.s, newCFtsent(t, fts.FTS_NSOK, path, nil, 0))
			default:
				panic(todo(""))
			}
		}
	}

	for {
		p := *(*uintptr)(unsafe.Pointer(path_argv))
		if p == 0 {
			if f == nil {
				return 0
			}

			if compar != 0 {
				panic(todo(""))
			}

			return addObject(f)
		}

		walk(GoString(p))
		path_argv += unsafe.Sizeof(uintptr(0))
	}
}

func Xfts_read(t *TLS, ftsp uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v ftsp=%v, (%v:)", t, ftsp, origin(2))
	}
	f := getObject(ftsp).(*ftstream)
	if f.x == len(f.s) {
		if dmesgs {
			dmesg("%v: FAIL", origin(1))
		}
		t.setErrno(0)
		return 0
	}

	r := f.s[f.x]
	if e := (*fts.FTSENT)(unsafe.Pointer(r)).Ffts_errno; e != 0 {
		if dmesgs {
			dmesg("%v: FAIL", origin(1))
		}
		t.setErrno(e)
	}
	f.x++
	return r
}

func Xfts_close(t *TLS, ftsp uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ftsp=%v, (%v:)", t, ftsp, origin(2))
	}
	getObject(ftsp).(*ftstream).close(t)
	removeObject(ftsp)
	return 0
}

func Xtzset(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}

}

var strerrorBuf [100]byte

func Xstrerror(t *TLS, errnum int32) uintptr {
	if __ccgo_strace {
		trc("t=%v errnum=%v, (%v:)", t, errnum, origin(2))
	}

	copy(strerrorBuf[:], fmt.Sprintf("strerror(%d)\x00", errnum))
	return uintptr(unsafe.Pointer(&strerrorBuf[0]))
}

func Xdlopen(t *TLS, filename uintptr, flags int32) uintptr {
	if __ccgo_strace {
		trc("t=%v filename=%v flags=%v, (%v:)", t, filename, flags, origin(2))
	}
	panic(todo(""))
}

func Xdlerror(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func Xdlclose(t *TLS, handle uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v handle=%v, (%v:)", t, handle, origin(2))
	}
	panic(todo(""))
}

func Xdlsym(t *TLS, handle, symbol uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v symbol=%v, (%v:)", t, symbol, origin(2))
	}
	panic(todo(""))
}

func Xperror(t *TLS, s uintptr) {
	if __ccgo_strace {
		trc("t=%v s=%v, (%v:)", t, s, origin(2))
	}
	panic(todo(""))
}

func Xpclose(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	panic(todo(""))
}

func Xgai_strerror(t *TLS, errcode int32) uintptr {
	if __ccgo_strace {
		trc("t=%v errcode=%v, (%v:)", t, errcode, origin(2))
	}
	panic(todo(""))

}

func Xtcgetattr(t *TLS, fd int32, termios_p uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v termios_p=%v, (%v:)", t, fd, termios_p, origin(2))
	}
	panic(todo(""))
}

func Xtcsetattr(t *TLS, fd, optional_actions int32, termios_p uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v optional_actions=%v termios_p=%v, (%v:)", t, optional_actions, termios_p, origin(2))
	}
	panic(todo(""))
}

func Xcfgetospeed(t *TLS, termios_p uintptr) termios.Speed_t {
	if __ccgo_strace {
		trc("t=%v termios_p=%v, (%v:)", t, termios_p, origin(2))
	}
	panic(todo(""))
}

func Xcfsetospeed(...interface{}) int32 {
	panic(todo(""))
}

func Xcfsetispeed(...interface{}) int32 {
	panic(todo(""))
}

func Xfork(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	if dmesgs {
		dmesg("%v: FAIL", origin(1))
	}
	t.setErrno(errno.ENOSYS)
	return -1
}

var emptyStr = [1]byte{}

func Xsetlocale(t *TLS, category int32, locale uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v category=%v locale=%v, (%v:)", t, category, locale, origin(2))
	}
	return uintptr(unsafe.Pointer(&emptyStr))
}

func Xnl_langinfo(t *TLS, item langinfo.Nl_item) uintptr {
	if __ccgo_strace {
		trc("t=%v item=%v, (%v:)", t, item, origin(2))
	}
	return uintptr(unsafe.Pointer(&emptyStr))
}

func Xpopen(t *TLS, command, type1 uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v type1=%v, (%v:)", t, type1, origin(2))
	}
	panic(todo(""))
}

func Xrealpath(t *TLS, path, resolved_path uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v resolved_path=%v, (%v:)", t, resolved_path, origin(2))
	}
	s, err := filepath.EvalSymlinks(GoString(path))
	if err != nil {
		if os.IsNotExist(err) {
			if dmesgs {
				dmesg("%v: %q: %v FAIL", origin(1), GoString(path), err)
			}
			t.setErrno(errno.ENOENT)
			return 0
		}

		panic(todo("", err))
	}

	if resolved_path == 0 {
		panic(todo(""))
	}

	if len(s) >= limits.PATH_MAX {
		s = s[:limits.PATH_MAX-1]
	}

	copy((*RawMem)(unsafe.Pointer(resolved_path))[:len(s):len(s)], s)
	(*RawMem)(unsafe.Pointer(resolved_path))[len(s)] = 0
	if dmesgs {
		dmesg("%v: %q: ok", origin(1), GoString(path))
	}
	return resolved_path
}

func Xinet_ntoa(t *TLS, in1 in.In_addr) uintptr {
	if __ccgo_strace {
		trc("t=%v in1=%v, (%v:)", t, in1, origin(2))
	}
	panic(todo(""))
}

func X__ccgo_in6addr_anyp(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))

}

func Xabort(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	if dmesgs {
		dmesg("%v:", origin(1))
	}
	p := Xcalloc(t, 1, types.Size_t(unsafe.Sizeof(signal.Sigaction{})))
	if p == 0 {
		panic("OOM")
	}

	(*signal.Sigaction)(unsafe.Pointer(p)).F__sigaction_u.F__sa_handler = signal.SIG_DFL
	Xsigaction(t, signal.SIGABRT, p, 0)
	Xfree(t, p)
	unix.Kill(unix.Getpid(), syscall.Signal(signal.SIGABRT))
	panic(todo("unrechable"))
}

func Xfflush(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	return 0
}

func Xfread(t *TLS, ptr uintptr, size, nmemb types.Size_t, stream uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v ptr=%v nmemb=%v stream=%v, (%v:)", t, ptr, nmemb, stream, origin(2))
	}
	fd := uintptr(file(stream).fd())
	count := size * nmemb
	var n int
	var err error
	switch {
	case count == 0:
		n, err = unix.Read(int(fd), nil)
	default:
		n, err = unix.Read(int(fd), (*RawMem)(unsafe.Pointer(ptr))[:count:count])
		if dmesgs && err == nil {
			dmesg("%v: fd %v, n %#x\n%s", origin(1), fd, n, hex.Dump((*RawMem)(unsafe.Pointer(ptr))[:n:n]))
		}
	}
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		file(stream).setErr()
		return types.Size_t(n) / size
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return types.Size_t(n) / size
}

func Xfwrite(t *TLS, ptr uintptr, size, nmemb types.Size_t, stream uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v ptr=%v nmemb=%v stream=%v, (%v:)", t, ptr, nmemb, stream, origin(2))
	}
	fd := uintptr(file(stream).fd())
	count := size * nmemb
	var n int
	var err error
	switch {
	case count == 0:
		n, err = unix.Write(int(fd), nil)
	default:
		n, err = unix.Write(int(fd), (*RawMem)(unsafe.Pointer(ptr))[:count:count])
		if dmesgs {
			dmesg("%v: fd %v, count %#x\n%s", origin(1), fd, count, hex.Dump((*RawMem)(unsafe.Pointer(ptr))[:count:count]))
		}
	}
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		file(stream).setErr()
		return types.Size_t(n) / size
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return types.Size_t(n) / size
}

func Xfclose(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	r := file(stream).close(t)
	if r != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), r)
		}
		t.setErrno(r)
		return stdio.EOF
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xfputc(t *TLS, c int32, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v stream=%v, (%v:)", t, c, stream, origin(2))
	}
	if _, err := fwrite(file(stream).fd(), []byte{byte(c)}); err != nil {
		return stdio.EOF
	}

	return int32(byte(c))
}

func Xfseek(t *TLS, stream uintptr, offset long, whence int32) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v offset=%v whence=%v, (%v:)", t, stream, offset, whence, origin(2))
	}
	if n := Xlseek(t, int32(file(stream).fd()), types.Off_t(offset), whence); n < 0 {
		if dmesgs {
			dmesg("%v: fd %v, off %#x, whence %v: %v", origin(1), file(stream).fd(), offset, whenceStr(whence), n)
		}
		file(stream).setErr()
		return -1
	}

	if dmesgs {
		dmesg("%v: fd %v, off %#x, whence %v: ok", origin(1), file(stream).fd(), offset, whenceStr(whence))
	}
	return 0
}

func Xftell(t *TLS, stream uintptr) long {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	n := Xlseek(t, file(stream).fd(), 0, stdio.SEEK_CUR)
	if n < 0 {
		file(stream).setErr()
		return -1
	}

	if dmesgs {
		dmesg("%v: fd %v, n %#x: ok %#x", origin(1), file(stream).fd(), n, long(n))
	}
	return long(n)
}

func Xferror(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	return Bool32(file(stream).err())
}

func Xfputs(t *TLS, s, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	panic(todo(""))

}

var getservbynameStaticResult netdb.Servent

func Xgetservbyname(t *TLS, name, proto uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v proto=%v, (%v:)", t, proto, origin(2))
	}
	var protoent *gonetdb.Protoent
	if proto != 0 {
		protoent = gonetdb.GetProtoByName(GoString(proto))
	}
	servent := gonetdb.GetServByName(GoString(name), protoent)
	if servent == nil {
		if dmesgs {
			dmesg("%q %q: nil (protoent %+v)", GoString(name), GoString(proto), protoent)
		}
		return 0
	}

	Xfree(t, (*netdb.Servent)(unsafe.Pointer(&getservbynameStaticResult)).Fs_name)
	if v := (*netdb.Servent)(unsafe.Pointer(&getservbynameStaticResult)).Fs_aliases; v != 0 {
		for {
			p := *(*uintptr)(unsafe.Pointer(v))
			if p == 0 {
				break
			}

			Xfree(t, p)
			v += unsafe.Sizeof(uintptr(0))
		}
		Xfree(t, v)
	}
	Xfree(t, (*netdb.Servent)(unsafe.Pointer(&getservbynameStaticResult)).Fs_proto)
	cname, err := CString(servent.Name)
	if err != nil {
		getservbynameStaticResult = netdb.Servent{}
		return 0
	}

	var protoname uintptr
	if protoent != nil {
		if protoname, err = CString(protoent.Name); err != nil {
			Xfree(t, cname)
			getservbynameStaticResult = netdb.Servent{}
			return 0
		}
	}
	var a []uintptr
	for _, v := range servent.Aliases {
		cs, err := CString(v)
		if err != nil {
			for _, v := range a {
				Xfree(t, v)
			}
			return 0
		}

		a = append(a, cs)
	}
	v := Xcalloc(t, types.Size_t(len(a)+1), types.Size_t(unsafe.Sizeof(uintptr(0))))
	if v == 0 {
		Xfree(t, cname)
		Xfree(t, protoname)
		for _, v := range a {
			Xfree(t, v)
		}
		getservbynameStaticResult = netdb.Servent{}
		return 0
	}
	for _, p := range a {
		*(*uintptr)(unsafe.Pointer(v)) = p
		v += unsafe.Sizeof(uintptr(0))
	}

	getservbynameStaticResult = netdb.Servent{
		Fs_name:    cname,
		Fs_aliases: v,
		Fs_port:    int32(servent.Port),
		Fs_proto:   protoname,
	}
	return uintptr(unsafe.Pointer(&getservbynameStaticResult))
}

func fcntlCmdStr(cmd int32) string {
	switch cmd {
	case fcntl.F_GETOWN:
		return "F_GETOWN"
	case fcntl.F_SETLK:
		return "F_SETLK"
	case fcntl.F_GETLK:
		return "F_GETLK"
	case fcntl.F_SETFD:
		return "F_SETFD"
	case fcntl.F_GETFD:
		return "F_GETFD"
	default:
		return fmt.Sprintf("cmd(%d)", cmd)
	}
}

func Xpwrite(t *TLS, fd int32, buf uintptr, count types.Size_t, offset types.Off_t) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v offset=%v, (%v:)", t, fd, buf, count, offset, origin(2))
	}
	var n int
	var err error
	switch {
	case count == 0:
		n, err = unix.Pwrite(int(fd), nil, int64(offset))
	default:
		n, err = unix.Pwrite(int(fd), (*RawMem)(unsafe.Pointer(buf))[:count:count], int64(offset))
		if dmesgs {
			dmesg("%v: fd %v, off %#x, count %#x\n%s", origin(1), fd, offset, count, hex.Dump((*RawMem)(unsafe.Pointer(buf))[:count:count]))
		}
	}
	if err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return types.Ssize_t(n)
}

func X_NSGetEnviron(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return EnvironP()
}

func Xchflags(t *TLS, path uintptr, flags uint32) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v flags=%v, (%v:)", t, path, flags, origin(2))
	}
	if err := unix.Chflags(GoString(path), int(flags)); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xrmdir(t *TLS, pathname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v, (%v:)", t, pathname, origin(2))
	}
	if err := unix.Rmdir(GoString(pathname)); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xmach_absolute_time(t *TLS) uint64 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return uint64(gotime.Now().UnixNano())
}

type machTimebaseInfo = struct {
	Fnumer uint32
	Fdenom uint32
}

func Xmach_timebase_info(t *TLS, info uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v info=%v, (%v:)", t, info, origin(2))
	}
	*(*machTimebaseInfo)(unsafe.Pointer(info)) = machTimebaseInfo{Fnumer: 1, Fdenom: 1}
	return 0
}

func Xgetattrlist(t *TLS, path, attrList, attrBuf uintptr, attrBufSize types.Size_t, options uint32) int32 {
	if __ccgo_strace {
		trc("t=%v attrBuf=%v attrBufSize=%v options=%v, (%v:)", t, attrBuf, attrBufSize, options, origin(2))
	}
	if _, _, err := unix.Syscall6(unix.SYS_GETATTRLIST, path, attrList, attrBuf, uintptr(attrBufSize), uintptr(options), 0); err != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xsetattrlist(t *TLS, path, attrList, attrBuf uintptr, attrBufSize types.Size_t, options uint32) int32 {
	if __ccgo_strace {
		trc("t=%v attrBuf=%v attrBufSize=%v options=%v, (%v:)", t, attrBuf, attrBufSize, options, origin(2))
	}
	if _, _, err := unix.Syscall6(unix.SYS_SETATTRLIST, path, attrList, attrBuf, uintptr(attrBufSize), uintptr(options), 0); err != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	return 0
}

func Xcopyfile(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xtruncate(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

type darwinDir struct {
	buf [4096]byte
	fd  int
	h   int
	l   int

	eof bool
}

func Xopendir(t *TLS, name uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v name=%v, (%v:)", t, name, origin(2))
	}
	p := Xmalloc(t, uint64(unsafe.Sizeof(darwinDir{})))
	if p == 0 {
		panic("OOM")
	}

	fd := int(Xopen(t, name, fcntl.O_RDONLY|fcntl.O_DIRECTORY|fcntl.O_CLOEXEC, 0))
	if fd < 0 {
		if dmesgs {
			dmesg("%v: FAIL %v", origin(1), (*darwinDir)(unsafe.Pointer(p)).fd)
		}
		Xfree(t, p)

		return 0
	}

	if dmesgs {
		dmesg("%v: ok", origin(1))
	}
	(*darwinDir)(unsafe.Pointer(p)).fd = fd
	(*darwinDir)(unsafe.Pointer(p)).h = 0
	(*darwinDir)(unsafe.Pointer(p)).l = 0
	(*darwinDir)(unsafe.Pointer(p)).eof = false

	return p
}

func Xreaddir(t *TLS, dir uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v dir=%v, (%v:)", t, dir, origin(2))
	}
	if (*darwinDir)(unsafe.Pointer(dir)).eof {
		return 0
	}

	if (*darwinDir)(unsafe.Pointer(dir)).l == (*darwinDir)(unsafe.Pointer(dir)).h {
		n, err := unix.Getdirentries((*darwinDir)(unsafe.Pointer(dir)).fd, (*darwinDir)(unsafe.Pointer(dir)).buf[:], nil)

		if n == 0 {
			if err != nil && err != io.EOF {
				if dmesgs {
					dmesg("%v: %v FAIL", origin(1), err)
				}
				t.setErrno(err)
			}
			(*darwinDir)(unsafe.Pointer(dir)).eof = true
			return 0
		}

		(*darwinDir)(unsafe.Pointer(dir)).l = 0
		(*darwinDir)(unsafe.Pointer(dir)).h = n

	}
	de := dir + unsafe.Offsetof(darwinDir{}.buf) + uintptr((*darwinDir)(unsafe.Pointer(dir)).l)

	(*darwinDir)(unsafe.Pointer(dir)).l += int((*unix.Dirent)(unsafe.Pointer(de)).Reclen)

	return de
}

func Xclosedir(t *TLS, dir uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v dir=%v, (%v:)", t, dir, origin(2))
	}

	r := Xclose(t, int32((*darwinDir)(unsafe.Pointer(dir)).fd))
	Xfree(t, dir)
	return r
}

func Xpipe(t *TLS, pipefd uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pipefd=%v, (%v:)", t, pipefd, origin(2))
	}
	var a [2]int
	if err := syscall.Pipe(a[:]); err != nil {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	*(*[2]int32)(unsafe.Pointer(pipefd)) = [2]int32{int32(a[0]), int32(a[1])}
	if dmesgs {
		dmesg("%v: %v ok", origin(1), a)
	}
	return 0
}

func X__isoc99_sscanf(t *TLS, str, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v va=%v, (%v:)", t, va, origin(2))
	}
	r := scanf(strings.NewReader(GoString(str)), format, va)

	return r
}

func Xsscanf(t *TLS, str, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v va=%v, (%v:)", t, va, origin(2))
	}
	r := scanf(strings.NewReader(GoString(str)), format, va)

	return r
}

func Xposix_fadvise(t *TLS, fd int32, offset, len types.Off_t, advice int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v len=%v advice=%v, (%v:)", t, fd, len, advice, origin(2))
	}
	panic(todo(""))
}

func Xclock(t *TLS) time.Clock_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return time.Clock_t(gotime.Since(startTime) * gotime.Duration(time.CLOCKS_PER_SEC) / gotime.Second)
}

func Xiswspace(t *TLS, wc wctype.Wint_t) int32 {
	if __ccgo_strace {
		trc("t=%v wc=%v, (%v:)", t, wc, origin(2))
	}
	return Bool32(unicode.IsSpace(rune(wc)))
}

func Xiswalnum(t *TLS, wc wctype.Wint_t) int32 {
	if __ccgo_strace {
		trc("t=%v wc=%v, (%v:)", t, wc, origin(2))
	}
	return Bool32(unicode.IsLetter(rune(wc)) || unicode.IsNumber(rune(wc)))
}

func Xarc4random_buf(t *TLS, buf uintptr, buflen size_t) {
	if __ccgo_strace {
		trc("t=%v buf=%v buflen=%v, (%v:)", t, buf, buflen, origin(2))
	}
	if _, err := crand.Read((*RawMem)(unsafe.Pointer(buf))[:buflen]); err != nil {
		panic(todo(""))
	}
}

type darwin_mutexattr_t struct {
	sig int64
	x   [8]byte
}

type darwin_mutex_t struct {
	sig int64
	x   [65]byte
}

func X__ccgo_pthreadMutexattrGettype(tls *TLS, a uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v, (%v:)", tls, a, origin(2))
	}
	return (int32((*darwin_mutexattr_t)(unsafe.Pointer(a)).x[4] >> 2 & 3))
}

func X__ccgo_getMutexType(tls *TLS, m uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v m=%v, (%v:)", tls, m, origin(2))
	}
	return (int32((*darwin_mutex_t)(unsafe.Pointer(m)).x[4] >> 2 & 3))
}

func X__ccgo_pthreadAttrGetDetachState(tls *TLS, a uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v, (%v:)", tls, a, origin(2))
	}
	return (*pthreadAttr)(unsafe.Pointer(a)).detachState
}

func Xpthread_attr_getdetachstate(tls *TLS, a uintptr, state uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v state=%v, (%v:)", tls, a, state, origin(2))
	}
	panic(todo(""))
}

func Xpthread_attr_setdetachstate(tls *TLS, a uintptr, state int32) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v state=%v, (%v:)", tls, a, state, origin(2))
	}
	panic(todo(""))
}

func Xpthread_mutexattr_destroy(tls *TLS, a uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v, (%v:)", tls, a, origin(2))
	}
	return 0
}

func Xpthread_mutexattr_init(tls *TLS, a uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v, (%v:)", tls, a, origin(2))
	}
	*(*darwin_mutexattr_t)(unsafe.Pointer(a)) = darwin_mutexattr_t{}
	return 0
}

func Xpthread_mutexattr_settype(tls *TLS, a uintptr, type1 int32) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v type1=%v, (%v:)", tls, a, type1, origin(2))
	}
	if uint32(type1) > uint32(2) {
		return errno.EINVAL
	}
	(*darwin_mutexattr_t)(unsafe.Pointer(a)).x[4] = byte(type1 << 2)
	return 0
}

func Xwritev(t *TLS, fd int32, iov uintptr, iovcnt int32) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v fd=%v iov=%v iovcnt=%v, (%v:)", t, fd, iov, iovcnt, origin(2))
	}

	r, _, err := unix.Syscall(unix.SYS_WRITEV, uintptr(fd), iov, uintptr(iovcnt))
	if err != 0 {
		if dmesgs {
			dmesg("%v: %v FAIL", origin(1), err)
		}
		t.setErrno(err)
		return -1
	}

	return types.Ssize_t(r)
}

func Xpause(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	c := make(chan os.Signal)
	gosignal.Notify(c,
		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGBUS,
		syscall.SIGCHLD,
		syscall.SIGCONT,
		syscall.SIGFPE,
		syscall.SIGHUP,
		syscall.SIGILL,
		syscall.SIGIO,
		syscall.SIGIOT,
		syscall.SIGKILL,
		syscall.SIGPIPE,
		syscall.SIGPROF,
		syscall.SIGQUIT,
		syscall.SIGSEGV,
		syscall.SIGSTOP,
		syscall.SIGSYS,
		syscall.SIGTERM,
		syscall.SIGTRAP,
		syscall.SIGTSTP,
		syscall.SIGTTIN,
		syscall.SIGTTOU,
		syscall.SIGURG,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGVTALRM,
		syscall.SIGWINCH,
		syscall.SIGXCPU,
		syscall.SIGXFSZ,
	)
	switch <-c {
	case syscall.SIGINT:
		panic(todo(""))
	default:
		t.setErrno(errno.EINTR)
		return -1
	}
}

func X__darwin_fd_set(tls *TLS, _fd int32, _p uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v _fd=%v _p=%v, (%v:)", tls, _fd, _p, origin(2))
	}
	*(*int32)(unsafe.Pointer(_p + uintptr(uint64(_fd)/(uint64(unsafe.Sizeof(int32(0)))*uint64(8)))*4)) |= int32(uint64(uint64(1)) << (uint64(_fd) % (uint64(unsafe.Sizeof(int32(0))) * uint64(8))))
	return int32(0)
}

func X__darwin_fd_isset(tls *TLS, _fd int32, _p uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v _fd=%v _p=%v, (%v:)", tls, _fd, _p, origin(2))
	}
	return *(*int32)(unsafe.Pointer(_p + uintptr(uint64(_fd)/(uint64(unsafe.Sizeof(int32(0)))*uint64(8)))*4)) & int32(uint64(uint64(1))<<(uint64(_fd)%(uint64(unsafe.Sizeof(int32(0)))*uint64(8))))
}

func X__darwin_fd_clr(tls *TLS, _fd int32, _p uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v _fd=%v _p=%v, (%v:)", tls, _fd, _p, origin(2))
	}
	*(*int32)(unsafe.Pointer(_p + uintptr(uint64(_fd)/(uint64(unsafe.Sizeof(int32(0)))*uint64(8)))*4)) &= ^int32(uint64(uint64(1)) << (uint64(_fd) % (uint64(unsafe.Sizeof(int32(0))) * uint64(8))))
	return int32(0)
}

func Xungetc(t *TLS, c int32, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v stream=%v, (%v:)", t, c, stream, origin(2))
	}
	panic(todo(""))
}

func Xissetugid(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

var progname uintptr

func Xgetprogname(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	if progname != 0 {
		return progname
	}

	var err error
	progname, err = CString(filepath.Base(os.Args[0]))
	if err != nil {
		t.setErrno(err)
		return 0
	}

	return progname
}

func Xuuid_copy(t *TLS, dst, src uintptr) {
	if __ccgo_strace {
		trc("t=%v src=%v, (%v:)", t, src, origin(2))
	}
	*(*uuid.Uuid_t)(unsafe.Pointer(dst)) = *(*uuid.Uuid_t)(unsafe.Pointer(src))
}

func Xuuid_parse(t *TLS, in uintptr, uu uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v in=%v uu=%v, (%v:)", t, in, uu, origin(2))
	}
	r, err := guuid.Parse(GoString(in))
	if err != nil {
		return -1
	}

	copy((*RawMem)(unsafe.Pointer(uu))[:unsafe.Sizeof(uuid.Uuid_t{})], r[:])
	return 0
}

func X__sincosf_stret(t *TLS, f float32) struct{ F__sinval, F__cosval float32 } {
	panic(todo(""))
}

func X__sincos_stret(t *TLS, f float64) struct{ F__sinval, F__cosval float64 } {
	panic(todo(""))
}

func X__sincospif_stret(t *TLS, f float32) struct{ F__sinval, F__cosval float32 } {
	panic(todo(""))
}

func X__sincospi_stret(t *TLS, f float64) struct{ F__sinval, F__cosval float64 } {
	panic(todo(""))
}

func X__srget(t *TLS, f uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v f=%v, (%v:)", t, f, origin(2))
	}
	panic(todo(""))
}

func X__svfscanf(t *TLS, f uintptr, p, q uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v f=%v q=%v, (%v:)", t, f, q, origin(2))
	}
	panic(todo(""))
}

func X__swbuf(t *TLS, i int32, f uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v i=%v f=%v, (%v:)", t, i, f, origin(2))
	}
	panic(todo(""))
}

func Xnanosleep(t *TLS, req, rem uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v rem=%v, (%v:)", t, rem, origin(2))
	}
	v := *(*time.Timespec)(unsafe.Pointer(req))
	gotime.Sleep(gotime.Second*gotime.Duration(v.Ftv_sec) + gotime.Duration(v.Ftv_nsec))
	return 0
}

func Xmalloc_size(t *TLS, ptr uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v ptr=%v, (%v:)", t, ptr, origin(2))
	}
	panic(todo(""))
}

func Xopen64(t *TLS, pathname uintptr, flags int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v args=%v, (%v:)", t, pathname, flags, args, origin(2))
	}
	var mode types.Mode_t
	if args != 0 {
		mode = (types.Mode_t)(VaUint32(&args))
	}
	fdcwd := fcntl.AT_FDCWD
	n, _, err := unix.Syscall6(unix.SYS_OPENAT, uintptr(fdcwd), pathname, uintptr(flags), uintptr(mode), 0, 0)
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return int32(n)
}

func Xgetrlimit(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_GETRLIMIT, uintptr(resource), uintptr(rlim), 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xsetrlimit(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_SETRLIMIT, uintptr(resource), uintptr(rlim), 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func X__fpclassifyd(tls *TLS, x float64) (r int32) {
	return X__fpclassify(tls, x)
}

var Xin6addr_any = in6_addr{}
