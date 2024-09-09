// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	// "encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	// "runtime/debug"
	"syscall"
	"time"
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
	// "modernc.org/sqlite/internal/libc/signal"
	"modernc.org/sqlite/internal/libc/stdio"
	"modernc.org/sqlite/internal/libc/sys/socket"
	"modernc.org/sqlite/internal/libc/sys/stat"
	"modernc.org/sqlite/internal/libc/sys/types"
	"modernc.org/sqlite/internal/libc/termios"
	ctime "modernc.org/sqlite/internal/libc/time"
	"modernc.org/sqlite/internal/libc/unistd"
	"modernc.org/sqlite/internal/libc/uuid/uuid"
)

const (
	m_F_EOF = 16
)

var (
	in6_addr_any in.In6_addr
	_            = X__ctype_b_loc
)

type (
	long  = int
	ulong = uint
)

type file uintptr

func (f file) fd() int32 {
	panic(todo(""))

}
func (f file) setFd(fd int32) {
	panic(todo(""))

}
func (f file) err() bool {
	panic(todo(""))

}
func (f file) setErr() {
	panic(todo(""))

}
func (f file) flags() int32 {
	panic(todo(""))

}
func (f file) orFlags(n int32) {
	panic(todo(""))

}
func (f file) xorFlags(n int32) {
	panic(todo(""))

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

	return unix.Write(int(fd), b)
}

func Xfprintf(t *TLS, stream, format, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v args=%v, (%v:)", t, args, origin(2))
	}
	panic(todo(""))

}

func Xusleep(t *TLS, usec uint) int32 {
	if __ccgo_strace {
		trc("t=%v usec=%v, (%v:)", t, usec, origin(2))
	}
	panic(todo(""))

}

func Xgetrusage(t *TLS, who int32, usage uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v who=%v usage=%v, (%v:)", t, who, usage, origin(2))
	}
	panic(todo(""))

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
	panic(todo(""))

}

var localtime ctime.Tm

func Xlocaltime(_ *TLS, timep uintptr) uintptr {
	panic(todo(""))

}

func Xlocaltime_r(_ *TLS, timep, result uintptr) uintptr {
	panic(todo(""))

}

func Xopen(t *TLS, pathname uintptr, flags int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v args=%v, (%v:)", t, pathname, flags, args, origin(2))
	}
	return Xopen64(t, pathname, flags, args)
}

func Xopen64(t *TLS, pathname uintptr, flags int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v args=%v, (%v:)", t, pathname, flags, args, origin(2))
	}
	panic(todo(""))

}

func Xopenat(t *TLS, dirfd int32, pathname uintptr, flags int32, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v flags=%v mode=%v, (%v:)", t, dirfd, pathname, flags, mode, origin(2))
	}
	panic(todo(""))

}

func Xlseek(t *TLS, fd int32, offset types.Off_t, whence int32) types.Off_t {
	if __ccgo_strace {
		trc("t=%v fd=%v offset=%v whence=%v, (%v:)", t, fd, offset, whence, origin(2))
	}
	return types.Off_t(Xlseek64(t, fd, offset, whence))
}

func whenceStr(whence int32) string {
	panic(todo(""))

}

var fsyncStatbuf stat.Stat

func Xfsync(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}
	if noFsync {
		return Xfstat(t, fd, uintptr(unsafe.Pointer(&fsyncStatbuf)))
	}

	panic(todo(""))

}

func Xsysconf(t *TLS, name int32) long {
	if __ccgo_strace {
		trc("t=%v name=%v, (%v:)", t, name, origin(2))
	}
	switch name {
	case unistd.X_SC_PAGESIZE:
		return long(unix.Getpagesize())
	case unistd.X_SC_GETPW_R_SIZE_MAX:
		return -1
	case unistd.X_SC_GETGR_R_SIZE_MAX:
		return -1
	case unistd.X_SC_NPROCESSORS_ONLN:
		return long(runtime.NumCPU())
	}

	panic(todo("", name))
}

func Xclose(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}
	panic(todo(""))

}

func Xgetcwd(t *TLS, buf uintptr, size types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v buf=%v size=%v, (%v:)", t, buf, size, origin(2))
	}
	panic(todo(""))

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
	return Xftruncate64(t, fd, length)
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
	panic(todo(""))

}

func Xwrite(t *TLS, fd int32, buf uintptr, count types.Size_t) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	panic(todo(""))

}

func Xfchmod(t *TLS, fd int32, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v mode=%v, (%v:)", t, fd, mode, origin(2))
	}
	panic(todo(""))

}

func Xfchmodat(t *TLS, dirfd int32, pathname uintptr, mode types.Mode_t, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v mode=%v flags=%v, (%v:)", t, dirfd, pathname, mode, flags, origin(2))
	}
	panic(todo(""))

}

func Xfchown(t *TLS, fd int32, owner types.Uid_t, group types.Gid_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v owner=%v group=%v, (%v:)", t, fd, owner, group, origin(2))
	}
	panic(todo(""))

}

func Xgeteuid(t *TLS) types.Uid_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))

}

func Xmunmap(t *TLS, addr uintptr, length types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v addr=%v length=%v, (%v:)", t, addr, length, origin(2))
	}
	panic(todo(""))

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
	panic(todo(""))

}

func Xsetsockopt(t *TLS, sockfd, level, optname int32, optval uintptr, optlen socket.Socklen_t) int32 {
	if __ccgo_strace {
		trc("t=%v optname=%v optval=%v optlen=%v, (%v:)", t, optname, optval, optlen, origin(2))
	}
	panic(todo(""))

}

func Xioctl(t *TLS, fd int32, request ulong, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v request=%v va=%v, (%v:)", t, fd, request, va, origin(2))
	}
	panic(todo(""))

}

func Xgetsockname(t *TLS, sockfd int32, addr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addrlen=%v, (%v:)", t, sockfd, addrlen, origin(2))
	}
	panic(todo(""))

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
		t.setErrno(err)
		return -1
	}

	return int32(n)
}

func Xmkfifo(t *TLS, pathname uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	if err := unix.Mkfifo(GoString(pathname), mode); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xumask(t *TLS, mask types.Mode_t) types.Mode_t {
	if __ccgo_strace {
		trc("t=%v mask=%v, (%v:)", t, mask, origin(2))
	}
	panic(todo(""))

}

func Xexecvp(t *TLS, file, argv uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v argv=%v, (%v:)", t, argv, origin(2))
	}
	panic(todo(""))

}

func Xwaitpid(t *TLS, pid types.Pid_t, wstatus uintptr, optname int32) types.Pid_t {
	if __ccgo_strace {
		trc("t=%v pid=%v wstatus=%v optname=%v, (%v:)", t, pid, wstatus, optname, origin(2))
	}
	panic(todo(""))

}

func Xuname(t *TLS, buf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v buf=%v, (%v:)", t, buf, origin(2))
	}
	panic(todo(""))

}

func Xrecv(t *TLS, sockfd int32, buf uintptr, len types.Size_t, flags int32) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v sockfd=%v buf=%v len=%v flags=%v, (%v:)", t, sockfd, buf, len, flags, origin(2))
	}
	panic(todo(""))

}

func Xsend(t *TLS, sockfd int32, buf uintptr, len types.Size_t, flags int32) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v sockfd=%v buf=%v len=%v flags=%v, (%v:)", t, sockfd, buf, len, flags, origin(2))
	}
	panic(todo(""))

}

func Xshutdown(t *TLS, sockfd, how int32) int32 {
	if __ccgo_strace {
		trc("t=%v how=%v, (%v:)", t, how, origin(2))
	}
	panic(todo(""))

}

func Xgetpeername(t *TLS, sockfd int32, addr uintptr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))

}

func Xsocket(t *TLS, domain, type1, protocol int32) int32 {
	if __ccgo_strace {
		trc("t=%v protocol=%v, (%v:)", t, protocol, origin(2))
	}
	panic(todo(""))

}

func Xbind(t *TLS, sockfd int32, addr uintptr, addrlen uint32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))

}

func Xconnect(t *TLS, sockfd int32, addr uintptr, addrlen uint32) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))

}

func Xlisten(t *TLS, sockfd, backlog int32) int32 {
	if __ccgo_strace {
		trc("t=%v backlog=%v, (%v:)", t, backlog, origin(2))
	}
	panic(todo(""))

}

func Xaccept(t *TLS, sockfd int32, addr uintptr, addrlen uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v sockfd=%v addr=%v addrlen=%v, (%v:)", t, sockfd, addr, addrlen, origin(2))
	}
	panic(todo(""))

}

func Xgetrlimit(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	return Xgetrlimit64(t, resource, rlim)
}

func Xsetrlimit(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	return Xsetrlimit64(t, resource, rlim)
}

func Xgetuid(t *TLS) types.Uid_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return types.Uid_t(os.Getuid())
}

func Xgetpid(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return int32(os.Getpid())
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

func Xbacktrace(t *TLS, buf uintptr, size int32) int32 {
	if __ccgo_strace {
		trc("t=%v buf=%v size=%v, (%v:)", t, buf, size, origin(2))
	}
	panic(todo(""))
}

func Xbacktrace_symbols_fd(t *TLS, buffer uintptr, size, fd int32) {
	if __ccgo_strace {
		trc("t=%v buffer=%v fd=%v, (%v:)", t, buffer, fd, origin(2))
	}
	panic(todo(""))
}

func Xfileno(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	panic(todo(""))

}

func newFtsent(t *TLS, info int, path string, stat *unix.Stat_t, err syscall.Errno) (r *fts.FTSENT) {
	panic(todo(""))

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
	return Xfts64_open(t, path_argv, options, compar)
}

func Xfts64_open(t *TLS, path_argv uintptr, options int32, compar uintptr) uintptr {
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
	return Xfts64_read(t, ftsp)
}

func Xfts64_read(t *TLS, ftsp uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v ftsp=%v, (%v:)", t, ftsp, origin(2))
	}
	f := getObject(ftsp).(*ftstream)
	if f.x == len(f.s) {
		t.setErrno(0)
		return 0
	}

	r := f.s[f.x]
	if e := (*fts.FTSENT)(unsafe.Pointer(r)).Ffts_errno; e != 0 {
		t.setErrno(e)
	}
	f.x++
	return r
}

func Xfts_close(t *TLS, ftsp uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ftsp=%v, (%v:)", t, ftsp, origin(2))
	}
	return Xfts64_close(t, ftsp)
}

func Xfts64_close(t *TLS, ftsp uintptr) int32 {
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
	panic(todo("%q", GoString(filename)))
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

var gai_strerrorBuf [100]byte

func Xgai_strerror(t *TLS, errcode int32) uintptr {
	if __ccgo_strace {
		trc("t=%v errcode=%v, (%v:)", t, errcode, origin(2))
	}
	copy(gai_strerrorBuf[:], fmt.Sprintf("gai error %d\x00", errcode))
	return uintptr(unsafe.Pointer(&gai_strerrorBuf))
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

func Xcfsetospeed(t *TLS, termios_p uintptr, speed uint32) int32 {
	if __ccgo_strace {
		trc("t=%v termios_p=%v speed=%v, (%v:)", t, termios_p, speed, origin(2))
	}
	panic(todo(""))
}

func Xcfsetispeed(t *TLS, termios_p uintptr, speed uint32) int32 {
	if __ccgo_strace {
		trc("t=%v termios_p=%v speed=%v, (%v:)", t, termios_p, speed, origin(2))
	}
	panic(todo(""))
}

func Xfork(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
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
	return uintptr(unsafe.Pointer(&in6_addr_any))
}

func Xabort(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))

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
	panic(todo(""))

}

func Xfwrite(t *TLS, ptr uintptr, size, nmemb types.Size_t, stream uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v ptr=%v nmemb=%v stream=%v, (%v:)", t, ptr, nmemb, stream, origin(2))
	}
	panic(todo(""))

}

func Xfclose(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	return file(stream).close(t)
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
		file(stream).setErr()
		return -1
	}

	file(stream).xorFlags(m_F_EOF)
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

	return long(n)
}

func Xferror(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	return Bool32(file(stream).err())
}

func Xungetc(t *TLS, c int32, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v stream=%v, (%v:)", t, c, stream, origin(2))
	}
	if c == stdio.EOF {
		return c
	}

	ungetcMu.Lock()
	ungetc[stream] = byte(c)
	ungetcMu.Unlock()
	return int32(byte(c))
}

func Xfscanf(t *TLS, stream, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v va=%v, (%v:)", t, va, origin(2))
	}
	return scanf(&byteScanner{t: t, stream: stream}, format, va)
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

func Xreaddir64(t *TLS, dir uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v dir=%v, (%v:)", t, dir, origin(2))
	}
	return Xreaddir(t, dir)
}

func __syscall(r, _ uintptr, errno syscall.Errno) long {
	if errno != 0 {
		return long(-errno)
	}

	return long(r)
}

func X__syscall1(t *TLS, trap, p1 long) long {
	if __ccgo_strace {
		trc("t=%v p1=%v, (%v:)", t, p1, origin(2))
	}
	return __syscall(unix.Syscall(uintptr(trap), uintptr(p1), 0, 0))
}

func X__syscall3(t *TLS, trap, p1, p2, p3 long) long {
	if __ccgo_strace {
		trc("t=%v p3=%v, (%v:)", t, p3, origin(2))
	}
	return __syscall(unix.Syscall(uintptr(trap), uintptr(p1), uintptr(p2), uintptr(p3)))
}

func X__syscall4(t *TLS, trap, p1, p2, p3, p4 long) long {
	if __ccgo_strace {
		trc("t=%v p4=%v, (%v:)", t, p4, origin(2))
	}
	return __syscall(unix.Syscall6(uintptr(trap), uintptr(p1), uintptr(p2), uintptr(p3), uintptr(p4), 0, 0))
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

func Xsetenv(t *TLS, name, value uintptr, overwrite int32) int32 {
	if __ccgo_strace {
		trc("t=%v value=%v overwrite=%v, (%v:)", t, value, overwrite, origin(2))
	}
	panic(todo(""))
}

func Xunsetenv(t *TLS, name uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v name=%v, (%v:)", t, name, origin(2))
	}
	panic(todo(""))
}

func Xpause(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	err := unix.Pause()
	if err != nil {
		t.setErrno(err)
	}

	return -1
}

func Xwritev(t *TLS, fd int32, iov uintptr, iovcnt int32) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v fd=%v iov=%v iovcnt=%v, (%v:)", t, fd, iov, iovcnt, origin(2))
	}

	if iovcnt == 0 {
		panic(todo(""))
	}

	iovs := make([][]byte, iovcnt)
	for ; iovcnt != 0; iovcnt-- {
		base := (*unix.Iovec)(unsafe.Pointer(iov)).Base
		len := (*unix.Iovec)(unsafe.Pointer(iov)).Len

		if base != nil && len != 0 {
			iovs = append(iovs, (*RawMem)(unsafe.Pointer(base))[:len:len])
			iov += unsafe.Sizeof(unix.Iovec{})
		}
	}
	n, err := unix.Writev(int(fd), iovs)
	if err != nil {
		panic(todo(""))
	}

	return types.Ssize_t(n)
}

func X__isoc99_sscanf(t *TLS, str, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v va=%v, (%v:)", t, va, origin(2))
	}
	r := Xsscanf(t, str, format, va)

	return r
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

	}
	if err != nil {
		t.setErrno(err)
		return -1
	}

	return types.Ssize_t(n)
}

func Xfstatfs(t *TLS, fd int32, buf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v, (%v:)", t, fd, buf, origin(2))
	}
	panic(todo(""))

}

func Xgetrandom(t *TLS, buf uintptr, buflen size_t, flags uint32) ssize_t {
	if __ccgo_strace {
		trc("t=%v buf=%v buflen=%v flags=%v, (%v:)", t, buf, buflen, flags, origin(2))
	}
	panic(todo(""))

}

func Xposix_fadvise(t *TLS, fd int32, offset, len types.Off_t, advice int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v len=%v advice=%v, (%v:)", t, fd, len, advice, origin(2))
	}
	panic(todo(""))

}

func Xfgetc(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	panic(todo(""))

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

func Xmkdirat(t *TLS, dirfd int32, pathname uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v mode=%v, (%v:)", t, dirfd, pathname, mode, origin(2))
	}

	panic(todo(""))

}

func Xsymlinkat(t *TLS, target uintptr, newdirfd int32, linkpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v target=%v newdirfd=%v linkpath=%v, (%v:)", t, target, newdirfd, linkpath, origin(2))
	}
	panic(todo(""))

}

func Xutimensat(t *TLS, dirfd int32, pathname, times uintptr, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v times=%v flags=%v, (%v:)", t, dirfd, times, flags, origin(2))
	}
	panic(todo(""))

}

func Xunlinkat(t *TLS, dirfd int32, pathname uintptr, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v flags=%v, (%v:)", t, dirfd, pathname, flags, origin(2))
	}
	panic(todo(""))

}

func Xfaccessat(t *TLS, dirfd int32, pathname uintptr, mode, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v flags=%v, (%v:)", t, dirfd, pathname, flags, origin(2))
	}
	panic(todo(""))

}

func Xrenameat2(t *TLS, olddirfd int32, oldpath uintptr, newdirfd int32, newpath uintptr, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v olddirfd=%v oldpath=%v newdirfd=%v newpath=%v flags=%v, (%v:)", t, olddirfd, oldpath, newdirfd, newpath, flags, origin(2))
	}
	panic(todo(""))

}

func Xmknodat(t *TLS, dirfd int32, pathname uintptr, mode types.Mode_t, dev types.Dev_t) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v mode=%v dev=%v, (%v:)", t, dirfd, pathname, mode, dev, origin(2))
	}
	panic(todo(""))

}

func Xfchownat(t *TLS, dirfd int32, pathname uintptr, uid types.Uid_t, gid types.Gid_t, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v dirfd=%v pathname=%v uid=%v gid=%v flags=%v, (%v:)", t, dirfd, pathname, uid, gid, flags, origin(2))
	}
	panic(todo(""))

}

func Xlinkat(t *TLS, olddirfd int32, oldpath uintptr, newdirfd int32, newpath uintptr, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v olddirfd=%v oldpath=%v newdirfd=%v newpath=%v flags=%v, (%v:)", t, olddirfd, oldpath, newdirfd, newpath, flags, origin(2))
	}
	panic(todo(""))

}

func Xpipe2(t *TLS, pipefd uintptr, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v pipefd=%v flags=%v, (%v:)", t, pipefd, flags, origin(2))
	}

	panic(todo(""))

}

func Xdup3(t *TLS, oldfd int32, newfd int32, flags int32) int32 {
	if __ccgo_strace {
		trc("t=%v oldfd=%v newfd=%v flags=%v, (%v:)", t, oldfd, newfd, flags, origin(2))
	}

	panic(todo(""))

}

func Xreadlinkat(t *TLS, dirfd int32, pathname, buf uintptr, bufsiz types.Size_t) types.Ssize_t {
	if __ccgo_strace {
		trc("t=%v dirfd=%v buf=%v bufsiz=%v, (%v:)", t, dirfd, buf, bufsiz, origin(2))
	}

	panic(todo(""))

}

func Xnanosleep(t *TLS, req, rem uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v rem=%v, (%v:)", t, rem, origin(2))
	}
	v := *(*ctime.Timespec)(unsafe.Pointer(req))
	time.Sleep(time.Second*time.Duration(v.Ftv_sec) + time.Duration(v.Ftv_nsec))
	return 0
}

func Xfeof(t *TLS, f uintptr) (r int32) {
	if __ccgo_strace {
		trc("t=%v f=%v, (%v:)", t, f, origin(2))
		defer func() { trc("-> %v", r) }()
	}
	panic(todo(""))

}

type byteScanner struct {
	t      *TLS
	stream uintptr

	last byte
}

func (s *byteScanner) ReadByte() (byte, error) {
	c := Xfgetc(s.t, s.stream)
	if c < 0 {
		return 0, io.EOF
	}

	s.last = byte(c)
	return byte(c), nil
}

func (s *byteScanner) UnreadByte() error {
	Xungetc(s.t, int32(s.last), s.stream)
	return nil
}
