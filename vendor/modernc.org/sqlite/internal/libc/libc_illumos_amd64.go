// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	// "os"
	// "strings"
	gotime "time"
	"unicode"
	"unsafe"

	"golang.org/x/sys/unix"
	// "modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/fcntl"
	// "modernc.org/sqlite/internal/libc/signal"
	"modernc.org/sqlite/internal/libc/stdio"
	"modernc.org/sqlite/internal/libc/sys/types"
	"modernc.org/sqlite/internal/libc/time"
	"modernc.org/sqlite/internal/libc/wctype"
)

var (
	startTime = gotime.Now()
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
	var arg uintptr
	if args != 0 {
		arg = *(*uintptr)(unsafe.Pointer(args))
	}
	if cmd == fcntl.F_SETFL {
		arg |= unix.O_LARGEFILE
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

func Xmremap(t *TLS, old_address uintptr, old_size, new_size types.Size_t, flags int32, args uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v old_address=%v new_size=%v flags=%v args=%v, (%v:)", t, old_address, new_size, flags, args, origin(2))
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
	panic(todo(""))

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

func Xtime(t *TLS, tloc uintptr) types.Time_t {
	if __ccgo_strace {
		trc("t=%v tloc=%v, (%v:)", t, tloc, origin(2))
	}
	panic(todo(""))

}

func Xgetrlimit64(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
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

func Xchmod(t *TLS, pathname uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
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
	panic(todo(""))

}

func Xaccess(t *TLS, pathname uintptr, mode int32) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	panic(todo(""))

}

func Xrmdir(t *TLS, pathname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v, (%v:)", t, pathname, origin(2))
	}
	panic(todo(""))

}

func Xrename(t *TLS, oldpath, newpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v newpath=%v, (%v:)", t, newpath, origin(2))
	}
	panic(todo(""))

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
	panic(todo(""))

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

func Xsetrlimit64(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	panic(todo(""))

}

func AtomicLoadPInt8(addr uintptr) (val int8) {
	return int8(a_load_8(addr))
}

func AtomicLoadPInt16(addr uintptr) (val int16) {
	return int16(a_load_16(addr))
}

func AtomicLoadPUint8(addr uintptr) byte {
	return byte(a_load_8(addr))
}

func AtomicLoadPUint16(addr uintptr) uint16 {
	return uint16(a_load_16(addr))
}

func AtomicLoadNUint8(ptr uintptr, memorder int32) uint8 {
	return byte(a_load_8(ptr))
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

type Tin6_addr = struct {
	F__in6_union struct {
		F__s6_addr16 [0][8]uint16
		F__s6_addr32 [0][4]uint32
		F__s6_addr   [16]uint8
	}
}

var Xin6addr_any = Tin6_addr{}

func Xrewinddir(tls *TLS, f uintptr) {
	if __ccgo_strace {
		trc("tls=%v f=%v, (%v:)", tls, f, origin(2))
	}
	Xfseek(tls, f, 0, stdio.SEEK_SET)
}

func Xclock(t *TLS) time.Clock_t {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return time.Clock_t(gotime.Since(startTime) * gotime.Duration(time.CLOCKS_PER_SEC) / gotime.Second)
}
