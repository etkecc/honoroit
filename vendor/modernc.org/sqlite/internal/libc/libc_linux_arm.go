// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"os"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
	"modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/fcntl"
	"modernc.org/sqlite/internal/libc/signal"
	"modernc.org/sqlite/internal/libc/stdio"
	// "modernc.org/sqlite/internal/libc/sys/stat"
	"modernc.org/sqlite/internal/libc/sys/types"
	ctime "modernc.org/sqlite/internal/libc/time"
)

var (
	startTime = time.Now()
)

func Xsigaction(t *TLS, signum int32, act, oldact uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v signum=%v oldact=%v, (%v:)", t, signum, oldact, origin(2))
	}

	type k_sigaction struct {
		handler  uintptr
		flags    ulong
		restorer uintptr
		mask     [2]uint32
	}

	var kact, koldact uintptr
	if act != 0 {
		sz := int(unsafe.Sizeof(k_sigaction{}))
		kact = t.Alloc(sz)
		defer t.Free(sz)
		*(*k_sigaction)(unsafe.Pointer(kact)) = k_sigaction{
			handler:  (*signal.Sigaction)(unsafe.Pointer(act)).F__sigaction_handler.Fsa_handler,
			flags:    ulong((*signal.Sigaction)(unsafe.Pointer(act)).Fsa_flags),
			restorer: (*signal.Sigaction)(unsafe.Pointer(act)).Fsa_restorer,
		}
		Xmemcpy(t, kact+unsafe.Offsetof(k_sigaction{}.mask), act+unsafe.Offsetof(signal.Sigaction{}.Fsa_mask), types.Size_t(unsafe.Sizeof(k_sigaction{}.mask)))
	}
	if oldact != 0 {
		panic(todo(""))
	}

	if _, _, err := unix.Syscall6(unix.SYS_RT_SIGACTION, uintptr(signum), kact, koldact, unsafe.Sizeof(k_sigaction{}.mask), 0, 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	if oldact != 0 {
		panic(todo(""))
	}

	return 0
}

func Xfcntl64(t *TLS, fd, cmd int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v cmd=%v args=%v, (%v:)", t, cmd, args, origin(2))
	}
	var arg uintptr
	if args != 0 {
		arg = *(*uintptr)(unsafe.Pointer(args))
	}
	if cmd == fcntl.F_SETFL {
		arg |= unix.O_LARGEFILE
	}
	n, _, err := unix.Syscall(unix.SYS_FCNTL64, uintptr(fd), uintptr(cmd), arg)
	if err != 0 {
		t.setErrno(err)
		return -1
	}

	return int32(n)
}

func Xlstat64(t *TLS, pathname, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v statbuf=%v, (%v:)", t, statbuf, origin(2))
	}
	if err := unix.Lstat(GoString(pathname), (*unix.Stat_t)(unsafe.Pointer(statbuf))); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xstat64(t *TLS, pathname, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v statbuf=%v, (%v:)", t, statbuf, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_STAT64, pathname, statbuf, 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xfstat64(t *TLS, fd int32, statbuf uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v statbuf=%v, (%v:)", t, fd, statbuf, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_FSTAT64, uintptr(fd), statbuf, 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xmremap(t *TLS, old_address uintptr, old_size, new_size types.Size_t, flags int32, args uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v old_address=%v new_size=%v flags=%v args=%v, (%v:)", t, old_address, new_size, flags, args, origin(2))
	}
	var arg uintptr
	if args != 0 {
		arg = *(*uintptr)(unsafe.Pointer(args))
	}
	data, _, err := unix.Syscall6(unix.SYS_MREMAP, old_address, uintptr(old_size), uintptr(new_size), uintptr(flags), arg, 0)
	if err != 0 {
		t.setErrno(err)
		return ^uintptr(0)
	}

	return data
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
	data, _, err := unix.Syscall6(unix.SYS_MMAP2, addr, uintptr(length), uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset>>12))
	if err != 0 {
		t.setErrno(err)
		return ^uintptr(0)
	}

	return data
}

func Xsymlink(t *TLS, target, linkpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v linkpath=%v, (%v:)", t, linkpath, origin(2))
	}
	if err := unix.Symlink(GoString(target), GoString(linkpath)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xchmod(t *TLS, pathname uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	if err := unix.Chmod(GoString(pathname), uint32(mode)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xftruncate64(t *TLS, fd int32, length types.Off_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v length=%v, (%v:)", t, fd, length, origin(2))
	}
	if _, _, err := unix.Syscall6(unix.SYS_FTRUNCATE64, uintptr(fd), 0, uintptr(length), uintptr(length>>32), 0, 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xlseek64(t *TLS, fd int32, offset types.Off_t, whence int32) types.Off_t {
	if __ccgo_strace {
		trc("t=%v fd=%v offset=%v whence=%v, (%v:)", t, fd, offset, whence, origin(2))
	}
	n, err := unix.Seek(int(fd), int64(offset), int(whence))
	if err != nil {
		t.setErrno(err)
		return -1
	}

	return types.Off_t(n)
}

func Xutime(t *TLS, filename, times uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v times=%v, (%v:)", t, times, origin(2))
	}
	if dmesgs {
		dmesg("t=%v filename=%q times=%v, (%v:)", t, GoString(filename), times, origin(2))
	}
	switch {
	case times == 0:
		var buf unix.Utimbuf
		tm := time.Now().Unix()
		buf.Actime = int32(tm)
		buf.Modtime = int32(tm)
		if err := unix.Utime(GoString(filename), &buf); err != nil {
			if dmesgs {
				dmesg("t=%v err=%v", t, err)
			}
			t.setErrno(err)
			return -1
		}
	default:
		if err := unix.Utime(GoString(filename), (*unix.Utimbuf)(unsafe.Pointer(times))); err != nil {
			if dmesgs {
				dmesg("t=%v err=%v", t, err)
			}
			t.setErrno(err)
			return -1
		}
	}

	if dmesgs {
		dmesg("t=%v OK", t)
	}
	return 0
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
	if err := unix.Getrlimit(int(resource), (*unix.Rlimit)(unsafe.Pointer(rlim))); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xtime(t *TLS, tloc uintptr) types.Time_t {
	if __ccgo_strace {
		trc("t=%v tloc=%v, (%v:)", t, tloc, origin(2))
	}
	n := time.Now().UTC().Unix()
	if tloc != 0 {
		*(*types.Time_t)(unsafe.Pointer(tloc)) = types.Time_t(n)
	}
	return types.Time_t(n)
}

func Xutimes(t *TLS, filename, times uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v times=%v, (%v:)", t, times, origin(2))
	}
	var tv []unix.Timeval
	if times != 0 {
		tv = make([]unix.Timeval, 2)
		*(*[2]unix.Timeval)(unsafe.Pointer(&tv[0])) = *(*[2]unix.Timeval)(unsafe.Pointer(times))
	}
	if err := unix.Utimes(GoString(filename), tv); err != nil {
		t.setErrno(err)
		return -1
	}

	if times != 0 {
		*(*[2]unix.Timeval)(unsafe.Pointer(times)) = *(*[2]unix.Timeval)(unsafe.Pointer(&tv[0]))
	}

	return 0
}

func Xunlink(t *TLS, pathname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v, (%v:)", t, pathname, origin(2))
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, GoString(pathname), 0); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xaccess(t *TLS, pathname uintptr, mode int32) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	if err := unix.Faccessat(unix.AT_FDCWD, GoString(pathname), uint32(mode), 0); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xrmdir(t *TLS, pathname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v, (%v:)", t, pathname, origin(2))
	}
	if err := unix.Rmdir(GoString(pathname)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xrename(t *TLS, oldpath, newpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v newpath=%v, (%v:)", t, newpath, origin(2))
	}
	if err := unix.Rename(GoString(oldpath), GoString(newpath)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xmknod(t *TLS, pathname uintptr, mode types.Mode_t, dev types.Dev_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v dev=%v, (%v:)", t, pathname, mode, dev, origin(2))
	}
	panic(todo(""))
}

func Xchown(t *TLS, pathname uintptr, owner types.Uid_t, group types.Gid_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v owner=%v group=%v, (%v:)", t, pathname, owner, group, origin(2))
	}
	if err := unix.Chown(GoString(pathname), int(owner), int(group)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
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
	if _, _, err := unix.Syscall(unix.SYS_PIPE2, pipefd, 0, 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
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
	n, err := unix.Readlink(GoString(path), GoBytes(buf, int(bufsize)))
	if err != nil {
		t.setErrno(err)
		return -1
	}

	return types.Ssize_t(n)
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

	fd, _, err := unix.Syscall(unix.SYS_OPEN, pathname, uintptr(flags|unix.O_LARGEFILE), 0666)
	if err != 0 {
		t.setErrno(err)
		return 0
	}

	if p := newFile(t, int32(fd)); p != 0 {
		return p
	}

	Xclose(t, int32(fd))
	t.setErrno(errno.ENOMEM)
	return 0
}

func Xmkdir(t *TLS, path uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v mode=%v, (%v:)", t, path, mode, origin(2))
	}
	if err := unix.Mkdir(GoString(path), uint32(mode)); err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func Xsetrlimit64(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	if _, _, err := unix.Syscall(unix.SYS_SETRLIMIT, uintptr(resource), uintptr(rlim), 0); err != 0 {
		t.setErrno(err)
		return -1
	}

	return 0
}

var _table1 = [384]int32{
	129: int32(1),
	130: int32(2),
	131: int32(3),
	132: int32(4),
	133: int32(5),
	134: int32(6),
	135: int32(7),
	136: int32(8),
	137: int32(9),
	138: int32(10),
	139: int32(11),
	140: int32(12),
	141: int32(13),
	142: int32(14),
	143: int32(15),
	144: int32(16),
	145: int32(17),
	146: int32(18),
	147: int32(19),
	148: int32(20),
	149: int32(21),
	150: int32(22),
	151: int32(23),
	152: int32(24),
	153: int32(25),
	154: int32(26),
	155: int32(27),
	156: int32(28),
	157: int32(29),
	158: int32(30),
	159: int32(31),
	160: int32(32),
	161: int32(33),
	162: int32(34),
	163: int32(35),
	164: int32(36),
	165: int32(37),
	166: int32(38),
	167: int32(39),
	168: int32(40),
	169: int32(41),
	170: int32(42),
	171: int32(43),
	172: int32(44),
	173: int32(45),
	174: int32(46),
	175: int32(47),
	176: int32(48),
	177: int32(49),
	178: int32(50),
	179: int32(51),
	180: int32(52),
	181: int32(53),
	182: int32(54),
	183: int32(55),
	184: int32(56),
	185: int32(57),
	186: int32(58),
	187: int32(59),
	188: int32(60),
	189: int32(61),
	190: int32(62),
	191: int32(63),
	192: int32(64),
	193: int32('a'),
	194: int32('b'),
	195: int32('c'),
	196: int32('d'),
	197: int32('e'),
	198: int32('f'),
	199: int32('g'),
	200: int32('h'),
	201: int32('i'),
	202: int32('j'),
	203: int32('k'),
	204: int32('l'),
	205: int32('m'),
	206: int32('n'),
	207: int32('o'),
	208: int32('p'),
	209: int32('q'),
	210: int32('r'),
	211: int32('s'),
	212: int32('t'),
	213: int32('u'),
	214: int32('v'),
	215: int32('w'),
	216: int32('x'),
	217: int32('y'),
	218: int32('z'),
	219: int32(91),
	220: int32(92),
	221: int32(93),
	222: int32(94),
	223: int32(95),
	224: int32(96),
	225: int32('a'),
	226: int32('b'),
	227: int32('c'),
	228: int32('d'),
	229: int32('e'),
	230: int32('f'),
	231: int32('g'),
	232: int32('h'),
	233: int32('i'),
	234: int32('j'),
	235: int32('k'),
	236: int32('l'),
	237: int32('m'),
	238: int32('n'),
	239: int32('o'),
	240: int32('p'),
	241: int32('q'),
	242: int32('r'),
	243: int32('s'),
	244: int32('t'),
	245: int32('u'),
	246: int32('v'),
	247: int32('w'),
	248: int32('x'),
	249: int32('y'),
	250: int32('z'),
	251: int32(123),
	252: int32(124),
	253: int32(125),
	254: int32(126),
	255: int32(127),
}

var _ptable1 = uintptr(unsafe.Pointer(&_table1)) + uintptr(128)*4

func X__ctype_tolower_loc(tls *TLS) (r uintptr) {
	if __ccgo_strace {
		trc("tls=%v, (%v:)", tls, origin(2))
		defer func() { trc("-> %v", r) }()
	}
	return uintptr(unsafe.Pointer(&_ptable1))
}

var Xin6addr_any = Tin6_addr{}

type Tin6_addr = struct {
	F__in6_union struct {
		F__s6_addr16 [0][8]uint16
		F__s6_addr32 [0][4]uint32
		F__s6_addr   [16]uint8
	}
}

func Xrewinddir(tls *TLS, f uintptr) {
	if __ccgo_strace {
		trc("tls=%v f=%v, (%v:)", tls, f, origin(2))
	}
	Xfseek(tls, f, 0, stdio.SEEK_SET)
}

func AtomicLoadNUint8(ptr uintptr, memorder int32) uint8 {
	return byte(a_load_8(ptr))
}

func Xclock(t *TLS) ctime.Clock_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return ctime.Clock_t(time.Since(startTime) * time.Duration(ctime.CLOCKS_PER_SEC) / time.Second)
}
