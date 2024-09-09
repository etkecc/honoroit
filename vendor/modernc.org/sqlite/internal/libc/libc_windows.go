// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	gotime "time"
	"unicode"
	"unicode/utf16"
	"unsafe"

	"github.com/ncruces/go-strftime"
	"modernc.org/sqlite/internal/libc/errno"
	"modernc.org/sqlite/internal/libc/fcntl"
	"modernc.org/sqlite/internal/libc/limits"
	"modernc.org/sqlite/internal/libc/stdio"
	"modernc.org/sqlite/internal/libc/sys/stat"
	"modernc.org/sqlite/internal/libc/sys/types"
	"modernc.org/sqlite/internal/libc/time"
	"modernc.org/sqlite/internal/libc/unistd"
)

var X__imp__environ = EnvironP()
var X__imp__wenviron = uintptr(unsafe.Pointer(&wenviron))
var X_imp___environ = EnvironP()
var X_imp___wenviron = uintptr(unsafe.Pointer(&wenviron))
var X_iob [stdio.X_IOB_ENTRIES]stdio.FILE
var Xin6addr_any [16]byte
var Xtimezone long

var (
	iobMap     = map[uintptr]int32{}
	wenvValid  bool
	wenviron   uintptr
	winEnviron = []uintptr{0}
)

func init() {
	for i := range X_iob {
		iobMap[uintptr(unsafe.Pointer(&X_iob[i]))] = int32(i)
	}
}

func X__p__wenviron(t *TLS) uintptr {
	if !wenvValid {
		bootWinEnviron(t)
	}
	return uintptr(unsafe.Pointer(&wenviron))
}

func winGetObject(stream uintptr) interface{} {
	if fd, ok := iobMap[stream]; ok {
		f, _ := fdToFile(fd)
		return f
	}

	return getObject(stream)
}

type (
	long  = int32
	ulong = uint32
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procAreFileApisANSI            = modkernel32.NewProc("AreFileApisANSI")
	procCopyFileW                  = modkernel32.NewProc("CopyFileW")
	procCreateEventA               = modkernel32.NewProc("CreateEventA")
	procCreateEventW               = modkernel32.NewProc("CreateEventW")
	procCreateFileA                = modkernel32.NewProc("CreateFileA")
	procCreateFileMappingW         = modkernel32.NewProc("CreateFileMappingW")
	procCreateFileW                = modkernel32.NewProc("CreateFileW")
	procCreateHardLinkW            = modkernel32.NewProc("CreateHardLinkW")
	procCreatePipe                 = modkernel32.NewProc("CreatePipe")
	procCreateProcessA             = modkernel32.NewProc("CreateProcessA")
	procCreateProcessW             = modkernel32.NewProc("CreateProcessW")
	procCreateThread               = modkernel32.NewProc("CreateThread")
	procDeleteCriticalSection      = modkernel32.NewProc("DeleteCriticalSection")
	procDeviceIoControl            = modkernel32.NewProc("DeviceIoControl")
	procDuplicateHandle            = modkernel32.NewProc("DuplicateHandle")
	procEnterCriticalSection       = modkernel32.NewProc("EnterCriticalSection")
	procFindClose                  = modkernel32.NewProc("FindClose")
	procFindFirstFileExW           = modkernel32.NewProc("FindFirstFileExW")
	procFindFirstFileW             = modkernel32.NewProc("FindFirstFileW")
	procFindNextFileW              = modkernel32.NewProc("FindNextFileW")
	procFormatMessageW             = modkernel32.NewProc("FormatMessageW")
	procGetACP                     = modkernel32.NewProc("GetACP")
	procGetCommState               = modkernel32.NewProc("GetCommState")
	procGetComputerNameExW         = modkernel32.NewProc("GetComputerNameExW")
	procGetConsoleCP               = modkernel32.NewProc("GetConsoleCP")
	procGetConsoleScreenBufferInfo = modkernel32.NewProc("GetConsoleScreenBufferInfo")
	procGetCurrentProcess          = modkernel32.NewProc("GetCurrentProcess")
	procGetCurrentProcessId        = modkernel32.NewProc("GetCurrentProcessId")
	procGetCurrentThread           = modkernel32.NewProc("GetCurrentThread")
	procGetCurrentThreadId         = modkernel32.NewProc("GetCurrentThreadId")
	procGetEnvironmentVariableA    = modkernel32.NewProc("GetEnvironmentVariableA")
	procGetEnvironmentVariableW    = modkernel32.NewProc("GetEnvironmentVariableW")
	procGetExitCodeProcess         = modkernel32.NewProc("GetExitCodeProcess")
	procGetExitCodeThread          = modkernel32.NewProc("GetExitCodeThread")
	procGetFileAttributesA         = modkernel32.NewProc("GetFileAttributesA")
	procGetFileAttributesExA       = modkernel32.NewProc("GetFileAttributesExA")
	procGetFileAttributesExW       = modkernel32.NewProc("GetFileAttributesExW")
	procGetFileInformationByHandle = modkernel32.NewProc("GetFileInformationByHandle")
	procGetFileSize                = modkernel32.NewProc("GetFileSize")
	procGetFullPathNameW           = modkernel32.NewProc("GetFullPathNameW")
	procGetLastError               = modkernel32.NewProc("GetLastError")
	procGetLogicalDriveStringsA    = modkernel32.NewProc("GetLogicalDriveStringsA")
	procGetModuleFileNameW         = modkernel32.NewProc("GetModuleFileNameW")
	procGetModuleHandleA           = modkernel32.NewProc("GetModuleHandleA")
	procGetModuleHandleW           = modkernel32.NewProc("GetModuleHandleW")
	procGetPrivateProfileStringA   = modkernel32.NewProc("GetPrivateProfileStringA")
	procGetProcAddress             = modkernel32.NewProc("GetProcAddress")
	procGetProcessHeap             = modkernel32.NewProc("GetProcessHeap")
	procGetSystemInfo              = modkernel32.NewProc("GetSystemInfo")
	procGetSystemTime              = modkernel32.NewProc("GetSystemTime")
	procGetSystemTimeAsFileTime    = modkernel32.NewProc("GetSystemTimeAsFileTime")
	procGetTempFileNameW           = modkernel32.NewProc("GetTempFileNameW")
	procGetTickCount               = modkernel32.NewProc("GetTickCount")
	procGetVersionExA              = modkernel32.NewProc("GetVersionExA")
	procGetVersionExW              = modkernel32.NewProc("GetVersionExW")
	procGetVolumeInformationA      = modkernel32.NewProc("GetVolumeInformationA")
	procGetVolumeInformationW      = modkernel32.NewProc("GetVolumeInformationW")
	procHeapAlloc                  = modkernel32.NewProc("HeapAlloc")
	procHeapFree                   = modkernel32.NewProc("HeapFree")
	procInitializeCriticalSection  = modkernel32.NewProc("InitializeCriticalSection")
	procLeaveCriticalSection       = modkernel32.NewProc("LeaveCriticalSection")
	procLockFile                   = modkernel32.NewProc("LockFile")
	procLockFileEx                 = modkernel32.NewProc("LockFileEx")
	procLstrlenW                   = modkernel32.NewProc("lstrlenW")
	procMapViewOfFile              = modkernel32.NewProc("MapViewOfFile")
	procMoveFileW                  = modkernel32.NewProc("MoveFileW")
	procMultiByteToWideChar        = modkernel32.NewProc("MultiByteToWideChar")
	procOpenEventA                 = modkernel32.NewProc("OpenEventA")
	procOpenProcessToken           = modkernel32.NewProc("OpenProcessToken")
	procPeekConsoleInputW          = modkernel32.NewProc("PeekConsoleInputW")
	procPeekNamedPipe              = modkernel32.NewProc("PeekNamedPipe")
	procQueryPerformanceCounter    = modkernel32.NewProc("QueryPerformanceCounter")
	procQueryPerformanceFrequency  = modkernel32.NewProc("QueryPerformanceFrequency")
	procReadConsoleW               = modkernel32.NewProc("ReadConsoleW")
	procReadFile                   = modkernel32.NewProc("ReadFile")
	procResetEvent                 = modkernel32.NewProc("ResetEvent")
	procSearchPathW                = modkernel32.NewProc("SearchPathW")
	procSetConsoleCtrlHandler      = modkernel32.NewProc("SetConsoleCtrlHandler")
	procSetConsoleMode             = modkernel32.NewProc("SetConsoleMode")
	procSetConsoleTextAttribute    = modkernel32.NewProc("SetConsoleTextAttribute")
	procSetEvent                   = modkernel32.NewProc("SetEvent")
	procSetFilePointer             = modkernel32.NewProc("SetFilePointer")
	procSetFileTime                = modkernel32.NewProc("SetFileTime")
	procSleepEx                    = modkernel32.NewProc("SleepEx")
	procSystemTimeToFileTime       = modkernel32.NewProc("SystemTimeToFileTime")
	procTerminateThread            = modkernel32.NewProc("TerminateThread")
	procTryEnterCriticalSection    = modkernel32.NewProc("TryEnterCriticalSection")
	procUnlockFile                 = modkernel32.NewProc("UnlockFile")
	procUnlockFileEx               = modkernel32.NewProc("UnlockFileEx")
	procWaitForSingleObjectEx      = modkernel32.NewProc("WaitForSingleObjectEx")
	procWideCharToMultiByte        = modkernel32.NewProc("WideCharToMultiByte")
	procWriteConsoleA              = modkernel32.NewProc("WriteConsoleA")
	procWriteConsoleW              = modkernel32.NewProc("WriteConsoleW")
	procWriteFile                  = modkernel32.NewProc("WriteFile")

	modadvapi = syscall.NewLazyDLL("advapi32.dll")

	procAccessCheck                = modadvapi.NewProc("AccessCheck")
	procAddAce                     = modadvapi.NewProc("AddAce")
	procEqualSid                   = modadvapi.NewProc("EqualSid")
	procGetAce                     = modadvapi.NewProc("GetAce")
	procGetAclInformation          = modadvapi.NewProc("GetAclInformation")
	procGetFileSecurityA           = modadvapi.NewProc("GetFileSecurityA")
	procGetFileSecurityW           = modadvapi.NewProc("GetFileSecurityW")
	procGetLengthSid               = modadvapi.NewProc("GetLengthSid")
	procGetNamedSecurityInfoW      = modadvapi.NewProc("GetNamedSecurityInfoW")
	procGetSecurityDescriptorDacl  = modadvapi.NewProc("GetSecurityDescriptorDacl")
	procGetSecurityDescriptorOwner = modadvapi.NewProc("GetSecurityDescriptorOwner")
	procGetSidIdentifierAuthority  = modadvapi.NewProc("GetSidIdentifierAuthority")
	procGetSidLengthRequired       = modadvapi.NewProc("GetSidLengthRequired")
	procGetSidSubAuthority         = modadvapi.NewProc("GetSidSubAuthority")
	procGetTokenInformation        = modadvapi.NewProc("GetTokenInformation")
	procImpersonateSelf            = modadvapi.NewProc("ImpersonateSelf")
	procInitializeAcl              = modadvapi.NewProc("InitializeAcl")
	procInitializeSid              = modadvapi.NewProc("InitializeSid")
	procOpenThreadToken            = modadvapi.NewProc("OpenThreadToken")
	procRevertToSelf               = modadvapi.NewProc("RevertToSelf")

	modws2_32 = syscall.NewLazyDLL("ws2_32.dll")

	procWSAStartup = modws2_32.NewProc("WSAStartup")

	moduser32 = syscall.NewLazyDLL("user32.dll")

	procCharLowerW                  = moduser32.NewProc("CharLowerW")
	procCreateWindowExW             = moduser32.NewProc("CreateWindowExW")
	procMsgWaitForMultipleObjectsEx = moduser32.NewProc("MsgWaitForMultipleObjectsEx")
	procPeekMessageW                = moduser32.NewProc("PeekMessageW")
	procRegisterClassW              = moduser32.NewProc("RegisterClassW")
	procUnregisterClassW            = moduser32.NewProc("UnregisterClassW")
	procWaitForInputIdle            = moduser32.NewProc("WaitForInputIdle")

	netapi             = syscall.NewLazyDLL("netapi32.dll")
	procNetGetDCName   = netapi.NewProc("NetGetDCName")
	procNetUserGetInfo = netapi.NewProc("NetUserGetInfo")

	userenvapi                = syscall.NewLazyDLL("userenv.dll")
	procGetProfilesDirectoryW = userenvapi.NewProc("GetProfilesDirectoryW")

	modcrt        = syscall.NewLazyDLL("msvcrt.dll")
	procAccess    = modcrt.NewProc("_access")
	procChmod     = modcrt.NewProc("_chmod")
	procGmtime    = modcrt.NewProc("gmtime")
	procGmtime32  = modcrt.NewProc("_gmtime32")
	procGmtime64  = modcrt.NewProc("_gmtime64")
	procStat64i32 = modcrt.NewProc("_stat64i32")
	procStati64   = modcrt.NewProc("_stati64")
	procStrftime  = modcrt.NewProc("strftime")
	procStrtod    = modcrt.NewProc("strtod")

	moducrt         = syscall.NewLazyDLL("ucrtbase.dll")
	procFindfirst32 = moducrt.NewProc("_findfirst32")
	procFindnext32  = moducrt.NewProc("_findnext32")
)

var (
	threadCallback uintptr
)

func init() {
	isWindows = true
	threadCallback = syscall.NewCallback(ThreadProc)
}

var EBADF = errors.New("EBADF")

var w_nextFd int32 = 42
var w_fdLock sync.Mutex
var w_fd_to_file = map[int32]*file{}

type file struct {
	_fd    int32
	hadErr bool
	t      uintptr
	syscall.Handle
}

func addFile(hdl syscall.Handle, fd int32) uintptr {
	var f = file{_fd: fd, Handle: hdl}
	w_fdLock.Lock()
	defer w_fdLock.Unlock()
	w_fd_to_file[fd] = &f
	f.t = addObject(&f)
	return f.t
}

func remFile(f *file) {
	removeObject(f.t)
	w_fdLock.Lock()
	defer w_fdLock.Unlock()
	delete(w_fd_to_file, f._fd)
}

func fdToFile(fd int32) (*file, bool) {
	w_fdLock.Lock()
	defer w_fdLock.Unlock()
	f, ok := w_fd_to_file[fd]
	return f, ok
}

func wrapFdHandle(hdl syscall.Handle) (uintptr, int32) {
	newFd := atomic.AddInt32(&w_nextFd, 1)
	return addFile(hdl, newFd), newFd
}

func (f *file) err() bool {
	return f.hadErr
}

func (f *file) setErr() {
	f.hadErr = true
}

func (tls *TLS) SetLastError(_dwErrCode uint32) {
	if tls != nil {
		tls.lastError = _dwErrCode
	}
}

func (tls *TLS) GetLastError() (r uint32) {
	if tls == nil {
		return 0
	}

	return tls.lastError
}

func newFile(t *TLS, fd int32) uintptr {
	if fd == unistd.STDIN_FILENO {
		h, err := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
		if err != nil {
			panic("no console")
		}
		return addFile(h, fd)
	}
	if fd == unistd.STDOUT_FILENO {
		h, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
		if err != nil {
			panic("no console")
		}
		return addFile(h, fd)
	}
	if fd == unistd.STDERR_FILENO {
		h, err := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
		if err != nil {
			panic("no console")
		}
		return addFile(h, fd)
	}

	panic("unknown fd source")
	return 0
}

func (f *file) close(t *TLS) int32 {
	remFile(f)
	err := syscall.Close(f.Handle)
	if err != nil {
		return (-1)
	}
	return 0
}

func fwrite(fd int32, b []byte) (int, error) {
	if fd == unistd.STDOUT_FILENO {
		return write(b)
	}

	f, ok := fdToFile(fd)
	if !ok {
		return -1, EBADF
	}

	if dmesgs {
		dmesg("%v: fd %v: %s", origin(1), fd, b)
	}
	return syscall.Write(f.Handle, b)
}

func Xfprintf(t *TLS, stream, format, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v args=%v, (%v:)", t, args, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	n, _ := fwrite(f._fd, printf(format, args))
	return int32(n)
}

func Xusleep(t *TLS, usec types.Useconds_t) int32 {
	if __ccgo_strace {
		trc("t=%v usec=%v, (%v:)", t, usec, origin(2))
	}
	gotime.Sleep(gotime.Microsecond * gotime.Duration(usec))
	return 0
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
	err := syscall.Chdir(GoString(path))
	if err != nil {
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

func X_localtime64(_ *TLS, timep uintptr) uintptr {
	return Xlocaltime(nil, timep)
}

func Xlocaltime_r(_ *TLS, timep, result uintptr) uintptr {
	panic(todo(""))

}

func X_wopen(t *TLS, pathname uintptr, flags int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v args=%v, (%v:)", t, pathname, flags, args, origin(2))
	}
	var mode types.Mode_t
	if args != 0 {
		mode = *(*types.Mode_t)(unsafe.Pointer(args))
	}
	s := goWideString(pathname)
	h, err := syscall.Open(GoString(pathname), int(flags), uint32(mode))
	if err != nil {
		if dmesgs {
			dmesg("%v: %q %#x: %v", origin(1), s, flags, err)
		}

		t.setErrno(err)
		return 0
	}

	_, n := wrapFdHandle(h)
	if dmesgs {
		dmesg("%v: %q flags %#x mode %#o: fd %v", origin(1), s, flags, mode, n)
	}
	return n
}

func Xopen(t *TLS, pathname uintptr, flags int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v args=%v, (%v:)", t, pathname, flags, args, origin(2))
	}
	return Xopen64(t, pathname, flags, args)
}

func Xopen64(t *TLS, pathname uintptr, flags int32, cmode uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v flags=%v cmode=%v, (%v:)", t, pathname, flags, cmode, origin(2))
	}

	var mode types.Mode_t
	if cmode != 0 {
		mode = (types.Mode_t)(VaUint32(&cmode))
	}

	h, err := syscall.Open(GoString(pathname), int(flags), uint32(mode))
	if err != nil {
		if dmesgs {
			dmesg("%v: %q %#x: %v", origin(1), GoString(pathname), flags, err)
		}

		t.setErrno(err)
		return -1
	}

	_, n := wrapFdHandle(h)
	if dmesgs {
		dmesg("%v: %q flags %#x mode %#o: fd %v", origin(1), GoString(pathname), flags, mode, n)
	}
	return n
}

func Xlseek(t *TLS, fd int32, offset types.Off_t, whence int32) types.Off_t {
	if __ccgo_strace {
		trc("t=%v fd=%v offset=%v whence=%v, (%v:)", t, fd, offset, whence, origin(2))
	}
	return types.Off_t(Xlseek64(t, fd, offset, whence))
}

func whenceStr(whence int32) string {
	switch whence {
	case syscall.FILE_CURRENT:
		return "SEEK_CUR"
	case syscall.FILE_END:
		return "SEEK_END"
	case syscall.FILE_BEGIN:
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

	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	err := syscall.FlushFileBuffers(f.Handle)
	if err != nil {
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
	panic(todo(""))

}

func Xclose(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}

	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	err := syscall.Close(f.Handle)
	if err != nil {
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

	b := make([]uint16, size)
	n, err := syscall.GetCurrentDirectory(uint32(len(b)), &b[0])
	if err != nil {
		t.setErrno(err)
		return 0
	}

	var wd = []byte(string(utf16.Decode(b[0:n])))
	if types.Size_t(len(wd)) > size {
		t.setErrno(errno.ERANGE)
		return 0
	}

	copy((*RawMem)(unsafe.Pointer(buf))[:], wd)
	(*RawMem)(unsafe.Pointer(buf))[len(wd)] = 0

	if dmesgs {
		dmesg("%v: %q: ok", origin(1), GoString(buf))
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
	return Xftruncate64(t, fd, length)
}

func Xfcntl(t *TLS, fd, cmd int32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v cmd=%v args=%v, (%v:)", t, cmd, args, origin(2))
	}
	return Xfcntl64(t, fd, cmd, args)
}

func Xread(t *TLS, fd int32, buf uintptr, count uint32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	var obuf = ((*RawMem)(unsafe.Pointer(buf)))[:count]
	n, err := syscall.Read(f.Handle, obuf)
	if err != nil {
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %d %#x: %#x", origin(1), fd, count, n)
	}
	return int32(n)
}

func X_read(t *TLS, fd int32, buf uintptr, count uint32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	return Xread(t, fd, buf, count)
}

func Xwrite(t *TLS, fd int32, buf uintptr, count uint32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	var obuf = ((*RawMem)(unsafe.Pointer(buf)))[:count]
	n, err := syscall.Write(f.Handle, obuf)
	if err != nil {
		if dmesgs {
			dmesg("%v: fd %v, count %#x: %v", origin(1), fd, count, err)
		}
		t.setErrno(err)
		return -1
	}

	if dmesgs {
		dmesg("%v: %d %#x: %#x", origin(1), fd, count, n)
	}
	return int32(n)
}

func X_write(t *TLS, fd int32, buf uintptr, count uint32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v buf=%v count=%v, (%v:)", t, fd, buf, count, origin(2))
	}
	return Xwrite(t, fd, buf, count)
}

func Xfchmod(t *TLS, fd int32, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v mode=%v, (%v:)", t, fd, mode, origin(2))
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
	panic(todo(""))

}

func Xgetsockopt(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))

}

func Xsetsockopt(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xioctl(t *TLS, fd int32, request ulong, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v request=%v va=%v, (%v:)", t, fd, request, va, origin(2))
	}
	panic(todo(""))

}

func Xselect(t *TLS, nfds int32, readfds, writefds, exceptfds, timeout uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v nfds=%v timeout=%v, (%v:)", t, nfds, timeout, origin(2))
	}
	panic(todo(""))

}

func Xmkfifo(t *TLS, pathname uintptr, mode types.Mode_t) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	panic(todo(""))

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

func Xsetrlimit64(t *TLS, resource int32, rlim uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v resource=%v rlim=%v, (%v:)", t, resource, rlim, origin(2))
	}
	panic(todo(""))

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

func Xgetpwuid(t *TLS, uid uint32) uintptr {
	if __ccgo_strace {
		trc("t=%v uid=%v, (%v:)", t, uid, origin(2))
	}
	panic(todo(""))

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
	if stream == 0 {
		t.setErrno(errno.EBADF)
		return -1
	}

	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	return f._fd
}

func Xmkstemps(t *TLS, template uintptr, suffixlen int32) int32 {
	if __ccgo_strace {
		trc("t=%v template=%v suffixlen=%v, (%v:)", t, template, suffixlen, origin(2))
	}
	return Xmkstemps64(t, template, suffixlen)
}

func Xmkstemps64(t *TLS, template uintptr, suffixlen int32) int32 {
	if __ccgo_strace {
		trc("t=%v template=%v suffixlen=%v, (%v:)", t, template, suffixlen, origin(2))
	}
	panic(todo(""))

}

func Xmkstemp64(t *TLS, template uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v template=%v, (%v:)", t, template, origin(2))
	}
	return Xmkstemps64(t, template, 0)
}

type ftstream struct {
	s []uintptr
	x int
}

func Xfts64_open(t *TLS, path_argv uintptr, options int32, compar uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v path_argv=%v options=%v compar=%v, (%v:)", t, path_argv, options, compar, origin(2))
	}
	panic(todo(""))

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
	panic(todo(""))

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
	panic(todo(""))

}

func Xtzset(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}

}

var strerrorBuf [256]byte

func Xstrerror(t *TLS, errnum int32) uintptr {
	if __ccgo_strace {
		trc("t=%v errnum=%v, (%v:)", t, errnum, origin(2))
	}
	copy((*RawMem)(unsafe.Pointer(&strerrorBuf[0]))[:len(strerrorBuf):len(strerrorBuf)], fmt.Sprintf("errno %d\x00", errnum))
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

func Xsetlocale(t *TLS, category int32, locale uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v category=%v locale=%v, (%v:)", t, category, locale, origin(2))
	}
	return 0
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
				dmesg("%v: %q: %v", origin(1), GoString(path), err)
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
	return resolved_path
}

func Xgmtime_r(t *TLS, timep, result uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v result=%v, (%v:)", t, result, origin(2))
	}
	panic(todo(""))
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
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	err := syscall.FlushFileBuffers(f.Handle)
	if err != nil {
		t.setErrno(err)
		return -1
	}
	return 0
}

func Xfread(t *TLS, ptr uintptr, size, nmemb types.Size_t, stream uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v ptr=%v nmemb=%v stream=%v, (%v:)", t, ptr, nmemb, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return 0
	}

	var sz = size * nmemb
	var obuf = ((*RawMem)(unsafe.Pointer(ptr)))[:sz]
	n, err := syscall.Read(f.Handle, obuf)
	if err != nil {
		f.setErr()
		return 0
	}

	if dmesgs {
		dmesg("%v: %d %#x x %#x: %#x\n%s", origin(1), f._fd, size, nmemb, types.Size_t(n)/size)
	}

	return types.Size_t(n) / size

}

func Xfwrite(t *TLS, ptr uintptr, size, nmemb types.Size_t, stream uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v ptr=%v nmemb=%v stream=%v, (%v:)", t, ptr, nmemb, stream, origin(2))
	}
	if ptr == 0 || size == 0 {
		return 0
	}

	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return 0
	}

	var sz = size * nmemb
	var obuf = ((*RawMem)(unsafe.Pointer(ptr)))[:sz]
	n, err := syscall.Write(f.Handle, obuf)
	if err != nil {
		f.setErr()
		return 0
	}

	if dmesgs {
		dmesg("%v: %d %#x x %#x: %#x\n%s", origin(1), f._fd, size, nmemb, types.Size_t(n)/size)
	}
	return types.Size_t(n) / size
}

func Xfclose(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	return f.close(t)
}

func Xfputc(t *TLS, c int32, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v stream=%v, (%v:)", t, c, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	if _, err := fwrite(f._fd, []byte{byte(c)}); err != nil {
		return -1
	}
	return int32(byte(c))
}

func Xfseek(t *TLS, stream uintptr, offset long, whence int32) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v offset=%v whence=%v, (%v:)", t, stream, offset, whence, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	if n := Xlseek(t, f._fd, types.Off_t(offset), whence); n < 0 {
		if dmesgs {
			dmesg("%v: fd %v, off %#x, whence %v: %v", origin(1), f._fd, offset, whenceStr(whence), n)
		}
		f.setErr()
		return -1
	}

	if dmesgs {
		dmesg("%v: fd %v, off %#x, whence %v: ok", origin(1), f._fd, offset, whenceStr(whence))
	}
	return 0
}

func Xftell(t *TLS, stream uintptr) long {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	n := Xlseek(t, f._fd, 0, syscall.FILE_CURRENT)
	if n < 0 {
		f.setErr()
		return -1
	}

	if dmesgs {
		dmesg("%v: fd %v, n %#x: ok %#x", origin(1), f._fd, n, long(n))
	}
	return long(n)
}

func Xferror(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	return Bool32(f.err())
}

func Xfgetc(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return stdio.EOF
	}

	var buf [1]byte
	if n, _ := syscall.Read(f.Handle, buf[:]); n != 0 {
		return int32(buf[0])
	}

	return stdio.EOF
}

func Xungetc(t *TLS, c int32, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v stream=%v, (%v:)", t, c, stream, origin(2))
	}
	panic(todo(""))
}

func Xfscanf(t *TLS, stream, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v va=%v, (%v:)", t, va, origin(2))
	}
	panic(todo(""))
}

func Xfputs(t *TLS, s, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	gS := GoString(s)
	if _, err := fwrite(f._fd, []byte(gS)); err != nil {
		return -1
	}
	return 0
}

func X_errno(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return t.errnop
}

func X__ms_vfscanf(t *TLS, stream, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	panic(todo(""))
}

func X__ms_vsscanf(t *TLS, str, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	panic(todo(""))
}

func X__ms_vscanf(t *TLS, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	panic(todo(""))
}

func X__ms_vsnprintf(t *TLS, str uintptr, size types.Size_t, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v str=%v size=%v ap=%v, (%v:)", t, str, size, ap, origin(2))
	}
	return Xvsnprintf(t, str, size, format, ap)
}

func X__ms_vfwscanf(t *TLS, stream uintptr, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v ap=%v, (%v:)", t, stream, ap, origin(2))
	}
	panic(todo(""))
}

func X__ms_vwscanf(t *TLS, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	panic(todo(""))
}

func X_vsnwprintf(t *TLS, buffer uintptr, count types.Size_t, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v buffer=%v count=%v ap=%v, (%v:)", t, buffer, count, ap, origin(2))
	}
	panic(todo(""))
}

func X__ms_vswscanf(t *TLS, stream uintptr, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v ap=%v, (%v:)", t, stream, ap, origin(2))
	}
	panic(todo(""))
}

func X__acrt_iob_func(t *TLS, fd uint32) uintptr {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}

	f, ok := fdToFile(int32(fd))
	if !ok {
		t.setErrno(EBADF)
		return 0
	}
	return f.t
}

func XSetEvent(t *TLS, hEvent uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hEvent=%v, (%v:)", t, hEvent, origin(2))
	}
	r0, _, err := syscall.Syscall(procSetEvent.Addr(), 1, hEvent, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_stricmp(t *TLS, string1, string2 uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v string2=%v, (%v:)", t, string2, origin(2))
	}
	var s1 = strings.ToLower(GoString(string1))
	var s2 = strings.ToLower(GoString(string2))
	return int32(strings.Compare(s1, s2))
}

func XHeapFree(t *TLS, hHeap uintptr, dwFlags uint32, lpMem uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hHeap=%v dwFlags=%v lpMem=%v, (%v:)", t, hHeap, dwFlags, lpMem, origin(2))
	}
	r0, _, err := syscall.Syscall(procHeapFree.Addr(), 3, hHeap, uintptr(dwFlags), lpMem)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetProcessHeap(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetProcessHeap.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func XHeapAlloc(t *TLS, hHeap uintptr, dwFlags uint32, dwBytes types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v hHeap=%v dwFlags=%v dwBytes=%v, (%v:)", t, hHeap, dwFlags, dwBytes, origin(2))
	}
	r0, _, err := syscall.Syscall(procHeapAlloc.Addr(), 3, hHeap, uintptr(dwFlags), uintptr(dwBytes))
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func Xgai_strerrorW(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func Xgetservbyname(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func XWspiapiGetAddrInfo(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xwcscmp(t *TLS, string1, string2 uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v string2=%v, (%v:)", t, string2, origin(2))
	}
	var s1 = goWideString(string1)
	var s2 = goWideString(string2)
	return int32(strings.Compare(s1, s2))
}

func XIsDebuggerPresent(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func XExitProcess(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetVersionExW(t *TLS, lpVersionInformation uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpVersionInformation=%v, (%v:)", t, lpVersionInformation, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetVersionExW.Addr(), 1, lpVersionInformation, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetVolumeNameForVolumeMountPointW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xwcslen(t *TLS, str uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v str=%v, (%v:)", t, str, origin(2))
	}
	r0, _, _ := syscall.Syscall(procLstrlenW.Addr(), 1, str, 0, 0)
	return types.Size_t(r0)
}

func XGetStdHandle(t *TLS, nStdHandle uint32) uintptr {
	if __ccgo_strace {
		trc("t=%v nStdHandle=%v, (%v:)", t, nStdHandle, origin(2))
	}
	h, err := syscall.GetStdHandle(int(nStdHandle))
	if err != nil {
		panic("no console")
	}
	return uintptr(h)
}

func XCloseHandle(t *TLS, hObject uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hObject=%v, (%v:)", t, hObject, origin(2))
	}
	r := syscall.CloseHandle(syscall.Handle(hObject))
	if r != nil {
		return errno.EINVAL
	}
	return 1
}

func XGetLastError(t *TLS) uint32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	var rv = *(*int32)(unsafe.Pointer(t.errnop))
	return uint32(rv)

}

func XSetFilePointer(t *TLS, hFile uintptr, lDistanceToMove long, lpDistanceToMoveHigh uintptr, dwMoveMethod uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v lDistanceToMove=%v lpDistanceToMoveHigh=%v dwMoveMethod=%v, (%v:)", t, hFile, lDistanceToMove, lpDistanceToMoveHigh, dwMoveMethod, origin(2))
	}
	r0, _, e1 := syscall.Syscall6(procSetFilePointer.Addr(), 4, hFile, uintptr(lDistanceToMove), lpDistanceToMoveHigh, uintptr(dwMoveMethod), 0, 0)
	var uOff = uint32(r0)
	if uOff == 0xffffffff {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return uint32(r0)
}

func XSetEndOfFile(t *TLS, hFile uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v, (%v:)", t, hFile, origin(2))
	}
	err := syscall.SetEndOfFile(syscall.Handle(hFile))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XReadFile(t *TLS, hFile, lpBuffer uintptr, nNumberOfBytesToRead uint32, lpNumberOfBytesRead, lpOverlapped uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nNumberOfBytesToRead=%v lpOverlapped=%v, (%v:)", t, lpBuffer, nNumberOfBytesToRead, lpOverlapped, origin(2))
	}
	r1, _, e1 := syscall.Syscall6(procReadFile.Addr(), 5,
		hFile, lpBuffer, uintptr(nNumberOfBytesToRead), uintptr(lpNumberOfBytesRead), uintptr(lpOverlapped), 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)
}

func XWriteFile(t *TLS, hFile, lpBuffer uintptr, nNumberOfBytesToWrite uint32, lpNumberOfBytesWritten, lpOverlapped uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nNumberOfBytesToWrite=%v lpOverlapped=%v, (%v:)", t, lpBuffer, nNumberOfBytesToWrite, lpOverlapped, origin(2))
	}
	r1, _, e1 := syscall.Syscall6(procWriteFile.Addr(), 5,
		hFile, lpBuffer, uintptr(nNumberOfBytesToWrite), lpNumberOfBytesWritten, lpOverlapped, 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)
}

func XGetFileAttributesW(t *TLS, lpFileName uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v, (%v:)", t, lpFileName, origin(2))
	}
	attrs, err := syscall.GetFileAttributes((*uint16)(unsafe.Pointer(lpFileName)))
	if attrs == syscall.INVALID_FILE_ATTRIBUTES {
		if err != nil {
			t.setErrno(err)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return attrs
}

func XCreateFileW(t *TLS, lpFileName uintptr, dwDesiredAccess, dwShareMode uint32, lpSecurityAttributes uintptr, dwCreationDisposition, dwFlagsAndAttributes uint32, hTemplateFile uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v dwShareMode=%v lpSecurityAttributes=%v dwFlagsAndAttributes=%v hTemplateFile=%v, (%v:)", t, lpFileName, dwShareMode, lpSecurityAttributes, dwFlagsAndAttributes, hTemplateFile, origin(2))
	}

	r0, _, e1 := syscall.Syscall9(procCreateFileW.Addr(), 7, lpFileName, uintptr(dwDesiredAccess), uintptr(dwShareMode), lpSecurityAttributes,
		uintptr(dwCreationDisposition), uintptr(dwFlagsAndAttributes), hTemplateFile, 0, 0)
	h := syscall.Handle(r0)
	if h == syscall.InvalidHandle {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return r0
	}
	return uintptr(h)
}

func XDuplicateHandle(t *TLS, hSourceProcessHandle, hSourceHandle, hTargetProcessHandle, lpTargetHandle uintptr, dwDesiredAccess uint32, bInheritHandle int32, dwOptions uint32) int32 {
	if __ccgo_strace {
		trc("t=%v lpTargetHandle=%v dwDesiredAccess=%v bInheritHandle=%v dwOptions=%v, (%v:)", t, lpTargetHandle, dwDesiredAccess, bInheritHandle, dwOptions, origin(2))
	}
	r0, _, err := syscall.Syscall9(procDuplicateHandle.Addr(), 7, hSourceProcessHandle, hSourceHandle, hTargetProcessHandle,
		lpTargetHandle, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(dwOptions), 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetCurrentProcess(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procGetCurrentProcess.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return r0
}

func XFlushFileBuffers(t *TLS, hFile uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v, (%v:)", t, hFile, origin(2))
	}
	err := syscall.FlushFileBuffers(syscall.Handle(hFile))
	if err != nil {
		t.setErrno(err)
		return -1
	}
	return 1

}

func XGetFileType(t *TLS, hFile uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v, (%v:)", t, hFile, origin(2))
	}
	n, err := syscall.GetFileType(syscall.Handle(hFile))
	if err != nil {
		t.setErrno(err)
	}
	return n
}

func XGetConsoleMode(t *TLS, hConsoleHandle, lpMode uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpMode=%v, (%v:)", t, lpMode, origin(2))
	}
	err := syscall.GetConsoleMode(syscall.Handle(hConsoleHandle), (*uint32)(unsafe.Pointer(lpMode)))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XGetCommState(t *TLS, hFile, lpDCB uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpDCB=%v, (%v:)", t, lpDCB, origin(2))
	}
	r1, _, err := syscall.Syscall(procGetCommState.Addr(), 2, hFile, lpDCB, 0)
	if r1 == 0 {
		t.setErrno(err)
		return 0
	}
	return int32(r1)
}

func X_wcsnicmp(t *TLS, string1, string2 uintptr, count types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v string2=%v count=%v, (%v:)", t, string2, count, origin(2))
	}

	var s1 = strings.ToLower(goWideString(string1))
	var l1 = len(s1)
	var s2 = strings.ToLower(goWideString(string2))
	var l2 = len(s2)

	if l1 < l2 {
		return -1
	}
	if l2 > l1 {
		return 1
	}

	var cmpLen = count
	if types.Size_t(l1) < cmpLen {
		cmpLen = types.Size_t(l1)
	}
	return int32(strings.Compare(s1[:cmpLen], s2[:cmpLen]))
}

func XReadConsoleW(t *TLS, hConsoleInput, lpBuffer uintptr, nNumberOfCharsToRead uint32, lpNumberOfCharsRead, pInputControl uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nNumberOfCharsToRead=%v pInputControl=%v, (%v:)", t, lpBuffer, nNumberOfCharsToRead, pInputControl, origin(2))
	}

	rv, _, err := syscall.Syscall6(procReadConsoleW.Addr(), 5, hConsoleInput,
		lpBuffer, uintptr(nNumberOfCharsToRead), lpNumberOfCharsRead, pInputControl, 0)
	if rv == 0 {
		t.setErrno(err)
	}

	return int32(rv)
}

func XWriteConsoleW(t *TLS, hConsoleOutput, lpBuffer uintptr, nNumberOfCharsToWrite uint32, lpNumberOfCharsWritten, lpReserved uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nNumberOfCharsToWrite=%v lpReserved=%v, (%v:)", t, lpBuffer, nNumberOfCharsToWrite, lpReserved, origin(2))
	}
	rv, _, err := syscall.Syscall6(procWriteConsoleW.Addr(), 5, hConsoleOutput,
		lpBuffer, uintptr(nNumberOfCharsToWrite), lpNumberOfCharsWritten, lpReserved, 0)
	if rv == 0 {
		t.setErrno(err)
	}
	return int32(rv)
}

func XWaitForSingleObject(t *TLS, hHandle uintptr, dwMilliseconds uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v hHandle=%v dwMilliseconds=%v, (%v:)", t, hHandle, dwMilliseconds, origin(2))
	}
	rv, err := syscall.WaitForSingleObject(syscall.Handle(hHandle), dwMilliseconds)
	if err != nil {
		t.setErrno(err)
	}
	return rv
}

func XResetEvent(t *TLS, hEvent uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hEvent=%v, (%v:)", t, hEvent, origin(2))
	}
	rv, _, err := syscall.Syscall(procResetEvent.Addr(), 1, hEvent, 0, 0)
	if rv == 0 {
		t.setErrno(err)
	}
	return int32(rv)
}

func XPeekConsoleInputW(t *TLS, hConsoleInput, lpBuffer uintptr, nLength uint32, lpNumberOfEventsRead uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nLength=%v lpNumberOfEventsRead=%v, (%v:)", t, lpBuffer, nLength, lpNumberOfEventsRead, origin(2))
	}
	r0, _, err := syscall.Syscall6(procPeekConsoleInputW.Addr(), 4, hConsoleInput, lpBuffer, uintptr(nLength), lpNumberOfEventsRead, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XwsprintfA(t *TLS, buf, format, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v args=%v, (%v:)", t, args, origin(2))
	}
	return Xsprintf(t, buf, format, args)
}

func XGetConsoleCP(t *TLS) uint32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetConsoleCP.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return uint32(r0)
}

func XCreateEventW(t *TLS, lpEventAttributes uintptr, bManualReset, bInitialState int32, lpName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpEventAttributes=%v bInitialState=%v lpName=%v, (%v:)", t, lpEventAttributes, bInitialState, lpName, origin(2))
	}
	r0, _, err := syscall.Syscall6(procCreateEventW.Addr(), 4, lpEventAttributes, uintptr(bManualReset),
		uintptr(bInitialState), lpName, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

type ThreadAdapter struct {
	token      uintptr
	tls        *TLS
	param      uintptr
	threadFunc func(*TLS, uintptr) uint32
}

func (ta *ThreadAdapter) run() uintptr {
	r := ta.threadFunc(ta.tls, ta.param)
	ta.tls.Close()
	removeObject(ta.token)
	return uintptr(r)
}

func ThreadProc(p uintptr) uintptr {
	adp, ok := winGetObject(p).(*ThreadAdapter)
	if !ok {
		panic("invalid thread")
	}
	return adp.run()
}

func XCreateThread(t *TLS, lpThreadAttributes uintptr, dwStackSize types.Size_t, lpStartAddress, lpParameter uintptr, dwCreationFlags uint32, lpThreadId uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpThreadAttributes=%v dwStackSize=%v lpParameter=%v dwCreationFlags=%v lpThreadId=%v, (%v:)", t, lpThreadAttributes, dwStackSize, lpParameter, dwCreationFlags, lpThreadId, origin(2))
	}
	f := (*struct{ f func(*TLS, uintptr) uint32 })(unsafe.Pointer(&struct{ uintptr }{lpStartAddress})).f
	var tAdp = ThreadAdapter{threadFunc: f, tls: NewTLS(), param: lpParameter}
	tAdp.token = addObject(&tAdp)

	r0, _, err := syscall.Syscall6(procCreateThread.Addr(), 6, lpThreadAttributes, uintptr(dwStackSize),
		threadCallback, tAdp.token, uintptr(dwCreationFlags), lpThreadId)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func XSetThreadPriority(t *TLS, hThread uintptr, nPriority int32) int32 {
	if __ccgo_strace {
		trc("t=%v hThread=%v nPriority=%v, (%v:)", t, hThread, nPriority, origin(2))
	}

	return 1
}

func XSetConsoleMode(t *TLS, hConsoleHandle uintptr, dwMode uint32) int32 {
	if __ccgo_strace {
		trc("t=%v hConsoleHandle=%v dwMode=%v, (%v:)", t, hConsoleHandle, dwMode, origin(2))
	}
	rv, _, err := syscall.Syscall(procSetConsoleMode.Addr(), 2, hConsoleHandle, uintptr(dwMode), 0)
	if rv == 0 {
		t.setErrno(err)
	}
	return int32(rv)
}

func XPurgeComm(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XClearCommError(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDeleteCriticalSection(t *TLS, lpCriticalSection uintptr) {
	if __ccgo_strace {
		trc("t=%v lpCriticalSection=%v, (%v:)", t, lpCriticalSection, origin(2))
	}
	syscall.Syscall(procDeleteCriticalSection.Addr(), 1, lpCriticalSection, 0, 0)
}

func XEnterCriticalSection(t *TLS, lpCriticalSection uintptr) {
	if __ccgo_strace {
		trc("t=%v lpCriticalSection=%v, (%v:)", t, lpCriticalSection, origin(2))
	}
	syscall.Syscall(procEnterCriticalSection.Addr(), 1, lpCriticalSection, 0, 0)
}

func XTryEnterCriticalSection(t *TLS, lpCriticalSection uintptr) (r int32) {
	if __ccgo_strace {
		trc("t=%v lpCriticalSection=%v, (%v:)", t, lpCriticalSection, origin(2))
	}
	r0, _, err := syscall.SyscallN(procTryEnterCriticalSection.Addr(), lpCriticalSection)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XLeaveCriticalSection(t *TLS, lpCriticalSection uintptr) {
	if __ccgo_strace {
		trc("t=%v lpCriticalSection=%v, (%v:)", t, lpCriticalSection, origin(2))
	}
	syscall.Syscall(procLeaveCriticalSection.Addr(), 1, lpCriticalSection, 0, 0)
}

func XGetOverlappedResult(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSetupComm(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSetCommTimeouts(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XInitializeCriticalSection(t *TLS, lpCriticalSection uintptr) {
	if __ccgo_strace {
		trc("t=%v lpCriticalSection=%v, (%v:)", t, lpCriticalSection, origin(2))
	}

	syscall.Syscall(procInitializeCriticalSection.Addr(), 1, lpCriticalSection, 0, 0)
}

func XBuildCommDCBW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSetCommState(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X_strnicmp(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XEscapeCommFunction(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetCommModemStatus(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XMoveFileW(t *TLS, lpExistingFileName, lpNewFileName uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpNewFileName=%v, (%v:)", t, lpNewFileName, origin(2))
	}
	r0, _, err := syscall.Syscall(procMoveFileW.Addr(), 2, lpExistingFileName, lpNewFileName, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetFullPathNameW(t *TLS, lpFileName uintptr, nBufferLength uint32, lpBuffer, lpFilePart uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v nBufferLength=%v lpFilePart=%v, (%v:)", t, lpFileName, nBufferLength, lpFilePart, origin(2))
	}
	r0, _, e1 := syscall.Syscall6(procGetFullPathNameW.Addr(), 4, lpFileName, uintptr(nBufferLength), uintptr(lpBuffer), uintptr(lpFilePart), 0, 0)
	n := uint32(r0)
	if n == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return n
}

func XCharLowerW(tls *TLS, _lpsz uintptr) (r uintptr) {
	if __ccgo_strace {
		trc("lpsz=%+v", _lpsz)
		defer func() { trc(`XCharLowerW->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procCharLowerW.Addr(), _lpsz)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return r0
}

func XCreateDirectoryW(t *TLS, lpPathName, lpSecurityAttributes uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpSecurityAttributes=%v, (%v:)", t, lpSecurityAttributes, origin(2))
	}
	err := syscall.CreateDirectory((*uint16)(unsafe.Pointer(lpPathName)),
		(*syscall.SecurityAttributes)(unsafe.Pointer(lpSecurityAttributes)))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XSetFileAttributesW(t *TLS, lpFileName uintptr, dwFileAttributes uint32) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v dwFileAttributes=%v, (%v:)", t, lpFileName, dwFileAttributes, origin(2))
	}
	err := syscall.SetFileAttributes((*uint16)(unsafe.Pointer(lpFileName)), dwFileAttributes)
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XGetTempFileNameW(t *TLS, lpPathName, lpPrefixString uintptr, uUnique uint32, lpTempFileName uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpPrefixString=%v uUnique=%v lpTempFileName=%v, (%v:)", t, lpPrefixString, uUnique, lpTempFileName, origin(2))
	}
	r0, _, e1 := syscall.Syscall6(procGetTempFileNameW.Addr(), 4, lpPathName, lpPrefixString, uintptr(uUnique), lpTempFileName, 0, 0)
	if r0 == 0 {
		t.setErrno(e1)
	}
	return uint32(r0)
}

func XCopyFileW(t *TLS, lpExistingFileName, lpNewFileName uintptr, bFailIfExists int32) int32 {
	if __ccgo_strace {
		trc("t=%v lpNewFileName=%v bFailIfExists=%v, (%v:)", t, lpNewFileName, bFailIfExists, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procCopyFileW.Addr(), 3, lpExistingFileName, lpNewFileName, uintptr(bFailIfExists))
	if r0 == 0 {
		t.setErrno(e1)
	}
	return int32(r0)
}

func XDeleteFileW(t *TLS, lpFileName uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v, (%v:)", t, lpFileName, origin(2))
	}
	err := syscall.DeleteFile((*uint16)(unsafe.Pointer(lpFileName)))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XRemoveDirectoryW(t *TLS, lpPathName uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpPathName=%v, (%v:)", t, lpPathName, origin(2))
	}
	err := syscall.RemoveDirectory((*uint16)(unsafe.Pointer(lpPathName)))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XFindFirstFileW(t *TLS, lpFileName, lpFindFileData uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpFindFileData=%v, (%v:)", t, lpFindFileData, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procFindFirstFileW.Addr(), 2, lpFileName, lpFindFileData, 0)
	handle := syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return r0
}

func XFindFirstFileExW(t *TLS, lpFileName uintptr, fInfoLevelId int32, lpFindFileData uintptr, fSearchOp int32, lpSearchFilter uintptr, dwAdditionalFlags uint32) uintptr {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v fInfoLevelId=%v lpFindFileData=%v fSearchOp=%v lpSearchFilter=%v dwAdditionalFlags=%v, (%v:)", t, lpFileName, fInfoLevelId, lpFindFileData, fSearchOp, lpSearchFilter, dwAdditionalFlags, origin(2))
	}
	r0, _, e1 := syscall.Syscall6(procFindFirstFileExW.Addr(), 6, lpFileName, uintptr(fInfoLevelId), lpFindFileData, uintptr(fSearchOp), lpSearchFilter, uintptr(dwAdditionalFlags))
	handle := syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return r0
}

func XFindClose(t *TLS, hFindFile uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hFindFile=%v, (%v:)", t, hFindFile, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procFindClose.Addr(), 1, hFindFile, 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return int32(r0)
}

func XFindNextFileW(t *TLS, hFindFile, lpFindFileData uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFindFileData=%v, (%v:)", t, lpFindFileData, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procFindNextFileW.Addr(), 2, hFindFile, lpFindFileData, 0)
	if r0 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return int32(r0)
}

func XGetLogicalDriveStringsA(t *TLS, nBufferLength uint32, lpBuffer uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v nBufferLength=%v lpBuffer=%v, (%v:)", t, nBufferLength, lpBuffer, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetLogicalDriveStringsA.Addr(), 2, uintptr(nBufferLength), lpBuffer, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return uint32(r0)
}

func XGetVolumeInformationA(t *TLS, lpRootPathName, lpVolumeNameBuffer uintptr, nVolumeNameSize uint32, lpVolumeSerialNumber, lpMaximumComponentLength, lpFileSystemFlags, lpFileSystemNameBuffer uintptr, nFileSystemNameSize uint32) int32 {
	if __ccgo_strace {
		trc("t=%v lpVolumeNameBuffer=%v nVolumeNameSize=%v lpFileSystemNameBuffer=%v nFileSystemNameSize=%v, (%v:)", t, lpVolumeNameBuffer, nVolumeNameSize, lpFileSystemNameBuffer, nFileSystemNameSize, origin(2))
	}
	r0, _, err := syscall.Syscall9(procGetVolumeInformationA.Addr(), 8,
		lpRootPathName,
		lpVolumeNameBuffer,
		uintptr(nVolumeNameSize),
		lpVolumeSerialNumber,
		lpMaximumComponentLength,
		lpFileSystemFlags,
		lpFileSystemNameBuffer,
		uintptr(nFileSystemNameSize),
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XCreateHardLinkW(t *TLS, lpFileName, lpExistingFileName, lpSecurityAttributes uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpSecurityAttributes=%v, (%v:)", t, lpSecurityAttributes, origin(2))
	}
	r0, _, err := syscall.Syscall(procCreateHardLinkW.Addr(), 1, lpFileName, lpExistingFileName, lpSecurityAttributes)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XDeviceIoControl(t *TLS, hDevice uintptr, dwIoControlCode uint32, lpInBuffer uintptr, nInBufferSize uint32, lpOutBuffer uintptr, nOutBufferSize uint32, lpBytesReturned, lpOverlapped uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hDevice=%v dwIoControlCode=%v lpInBuffer=%v nInBufferSize=%v lpOutBuffer=%v nOutBufferSize=%v lpOverlapped=%v, (%v:)", t, hDevice, dwIoControlCode, lpInBuffer, nInBufferSize, lpOutBuffer, nOutBufferSize, lpOverlapped, origin(2))
	}
	r0, _, err := syscall.Syscall9(procDeviceIoControl.Addr(), 8, hDevice, uintptr(dwIoControlCode), lpInBuffer,
		uintptr(nInBufferSize), lpOutBuffer, uintptr(nOutBufferSize), lpBytesReturned, lpOverlapped, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func Xwcsncmp(t *TLS, string1, string2 uintptr, count types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v string2=%v count=%v, (%v:)", t, string2, count, origin(2))
	}
	var s1 = goWideString(string1)
	var l1 = len(s1)
	var s2 = goWideString(string2)
	var l2 = len(s2)

	if l1 < l2 {
		return -1
	}
	if l2 > l1 {
		return 1
	}

	var cmpLen = count
	if types.Size_t(l1) < cmpLen {
		cmpLen = types.Size_t(l1)
	}
	return int32(strings.Compare(s1[:cmpLen], s2[:cmpLen]))
}

func XMultiByteToWideChar(t *TLS, CodePage uint32, dwFlags uint32, lpMultiByteStr uintptr, cbMultiByte int32, lpWideCharStr uintptr, cchWideChar int32) int32 {
	if __ccgo_strace {
		trc("t=%v CodePage=%v dwFlags=%v lpMultiByteStr=%v cbMultiByte=%v lpWideCharStr=%v cchWideChar=%v, (%v:)", t, CodePage, dwFlags, lpMultiByteStr, cbMultiByte, lpWideCharStr, cchWideChar, origin(2))
	}
	r1, _, _ := syscall.Syscall6(procMultiByteToWideChar.Addr(), 6,
		uintptr(CodePage), uintptr(dwFlags), uintptr(lpMultiByteStr),
		uintptr(cbMultiByte), uintptr(lpWideCharStr), uintptr(cchWideChar))
	return (int32(r1))
}

func XOutputDebugStringW(t *TLS, lpOutputString uintptr) {
	if __ccgo_strace {
		trc("t=%v lpOutputString=%v, (%v:)", t, lpOutputString, origin(2))
	}
	panic(todo(""))
}

func XMessageBeep(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X_InterlockedCompareExchange(t *TLS, Destination uintptr, Exchange, Comparand long) long {
	if __ccgo_strace {
		trc("t=%v Destination=%v Comparand=%v, (%v:)", t, Destination, Comparand, origin(2))
	}

	var v = *(*int32)(unsafe.Pointer(Destination))
	_ = atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(Destination)), Comparand, Exchange)
	return long(v)
}

func Xrename(t *TLS, oldpath, newpath uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v newpath=%v, (%v:)", t, newpath, origin(2))
	}
	panic(todo(""))
}

func XAreFileApisANSI(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}

	r0, _, _ := syscall.Syscall(procAreFileApisANSI.Addr(), 0, 0, 0, 0)
	return int32(r0)
}

func XCreateFileA(t *TLS, lpFileName uintptr, dwDesiredAccess, dwShareMode uint32,
	lpSecurityAttributes uintptr, dwCreationDisposition, dwFlagsAndAttributes uint32, hTemplateFile uintptr) uintptr {
	r0, _, e1 := syscall.Syscall9(procCreateFileA.Addr(), 7, lpFileName, uintptr(dwDesiredAccess), uintptr(dwShareMode), lpSecurityAttributes,
		uintptr(dwCreationDisposition), uintptr(dwFlagsAndAttributes), hTemplateFile, 0, 0)
	h := syscall.Handle(r0)
	if h == syscall.InvalidHandle {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return r0
	}
	return uintptr(h)

}

func XCreateFileMappingA(t *TLS, hFile, lpFileMappingAttributes uintptr, flProtect, dwMaximumSizeHigh, dwMaximumSizeLow uint32, lpName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpFileMappingAttributes=%v dwMaximumSizeLow=%v lpName=%v, (%v:)", t, lpFileMappingAttributes, dwMaximumSizeLow, lpName, origin(2))
	}
	panic(todo(""))
}

func XCreateFileMappingW(t *TLS, hFile, lpFileMappingAttributes uintptr, flProtect, dwMaximumSizeHigh, dwMaximumSizeLow uint32, lpName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpFileMappingAttributes=%v dwMaximumSizeLow=%v lpName=%v, (%v:)", t, lpFileMappingAttributes, dwMaximumSizeLow, lpName, origin(2))
	}
	h, _, e1 := syscall.Syscall6(procCreateFileMappingW.Addr(), 6, hFile, lpFileMappingAttributes, uintptr(flProtect),
		uintptr(dwMaximumSizeHigh), uintptr(dwMaximumSizeLow), lpName)
	if h == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return h
}

func XCreateMutexW(t *TLS, lpMutexAttributes uintptr, bInitialOwner int32, lpName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpMutexAttributes=%v bInitialOwner=%v lpName=%v, (%v:)", t, lpMutexAttributes, bInitialOwner, lpName, origin(2))
	}
	panic(todo(""))
}

func XDeleteFileA(t *TLS, lpFileName uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v, (%v:)", t, lpFileName, origin(2))
	}
	panic(todo(""))
}

func XFormatMessageA(t *TLS, dwFlagsAndAttributes uint32, lpSource uintptr, dwMessageId, dwLanguageId uint32, lpBuffer uintptr, nSize uint32, Arguments uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v dwFlagsAndAttributes=%v lpSource=%v dwLanguageId=%v lpBuffer=%v nSize=%v Arguments=%v, (%v:)", t, dwFlagsAndAttributes, lpSource, dwLanguageId, lpBuffer, nSize, Arguments, origin(2))
	}
	panic(todo(""))
}

func XFormatMessageW(t *TLS, dwFlags uint32, lpSource uintptr, dwMessageId, dwLanguageId uint32, lpBuffer uintptr, nSize uint32, Arguments uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v dwFlags=%v lpSource=%v dwLanguageId=%v lpBuffer=%v nSize=%v Arguments=%v, (%v:)", t, dwFlags, lpSource, dwLanguageId, lpBuffer, nSize, Arguments, origin(2))
	}
	r0, _, e1 := syscall.Syscall9(procFormatMessageW.Addr(), 7,
		uintptr(dwFlags), lpSource, uintptr(dwMessageId), uintptr(dwLanguageId),
		lpBuffer, uintptr(nSize), Arguments, 0, 0)
	n := uint32(r0)
	if n == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return n
}

func XFreeLibrary(t *TLS, hLibModule uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hLibModule=%v, (%v:)", t, hLibModule, origin(2))
	}
	panic(todo(""))
}

func XGetCurrentProcessId(t *TLS) uint32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, _ := syscall.Syscall(procGetCurrentProcessId.Addr(), 0, 0, 0, 0)
	pid := uint32(r0)
	return pid
}

func XGetDiskFreeSpaceA(t *TLS, lpRootPathName, lpSectorsPerCluster, lpBytesPerSector, lpNumberOfFreeClusters, lpTotalNumberOfClusters uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpTotalNumberOfClusters=%v, (%v:)", t, lpTotalNumberOfClusters, origin(2))
	}
	panic(todo(""))
}

func XGetDiskFreeSpaceW(t *TLS, lpRootPathName, lpSectorsPerCluster, lpBytesPerSector, lpNumberOfFreeClusters, lpTotalNumberOfClusters uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpTotalNumberOfClusters=%v, (%v:)", t, lpTotalNumberOfClusters, origin(2))
	}
	panic(todo(""))
}

func XGetFileAttributesA(t *TLS, lpFileName uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v, (%v:)", t, lpFileName, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetFileAttributesA.Addr(), 1, lpFileName, 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return uint32(r0)
}

func XGetFileAttributesExW(t *TLS, lpFileName uintptr, fInfoLevelId uint32, lpFileInformation uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v fInfoLevelId=%v lpFileInformation=%v, (%v:)", t, lpFileName, fInfoLevelId, lpFileInformation, origin(2))
	}
	r1, _, e1 := syscall.Syscall(procGetFileAttributesExW.Addr(), 3, lpFileName, uintptr(fInfoLevelId), lpFileInformation)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)
}

func XGetFileSize(t *TLS, hFile, lpFileSizeHigh uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpFileSizeHigh=%v, (%v:)", t, lpFileSizeHigh, origin(2))
	}
	r1, _, e1 := syscall.Syscall(procGetFileSize.Addr(), 2, hFile, lpFileSizeHigh, 0)
	if r1 == math.MaxUint32 {
		if lpFileSizeHigh == 0 {
			t.setErrno(e1)
			return math.MaxUint32
		} else {
			t.setErrno(e1)
			return math.MaxUint32
		}
	}
	return uint32(r1)
}

func XGetFullPathNameA(t *TLS, lpFileName uintptr, nBufferLength uint32, lpBuffer, lpFilePart uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v nBufferLength=%v lpFilePart=%v, (%v:)", t, lpFileName, nBufferLength, lpFilePart, origin(2))
	}
	panic(todo(""))
}

func XGetProcAddress(t *TLS, hModule, lpProcName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpProcName=%v, (%v:)", t, lpProcName, origin(2))
	}

	return 0

}

func XRtlGetVersion(t *TLS, lpVersionInformation uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpVersionInformation=%v, (%v:)", t, lpVersionInformation, origin(2))
	}
	panic(todo(""))
}

func XGetSystemInfo(t *TLS, lpSystemInfo uintptr) {
	if __ccgo_strace {
		trc("t=%v lpSystemInfo=%v, (%v:)", t, lpSystemInfo, origin(2))
	}
	syscall.Syscall(procGetSystemInfo.Addr(), 1, lpSystemInfo, 0, 0)
}

func XGetSystemTime(t *TLS, lpSystemTime uintptr) {
	if __ccgo_strace {
		trc("t=%v lpSystemTime=%v, (%v:)", t, lpSystemTime, origin(2))
	}
	syscall.Syscall(procGetSystemTime.Addr(), 1, lpSystemTime, 0, 0)
}

func XGetSystemTimeAsFileTime(t *TLS, lpSystemTimeAsFileTime uintptr) {
	if __ccgo_strace {
		trc("t=%v lpSystemTimeAsFileTime=%v, (%v:)", t, lpSystemTimeAsFileTime, origin(2))
	}
	syscall.Syscall(procGetSystemTimeAsFileTime.Addr(), 1, lpSystemTimeAsFileTime, 0, 0)
}

func XGetTempPathA(t *TLS, nBufferLength uint32, lpBuffer uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v nBufferLength=%v lpBuffer=%v, (%v:)", t, nBufferLength, lpBuffer, origin(2))
	}
	panic(todo(""))
}

func XGetTempPathW(t *TLS, nBufferLength uint32, lpBuffer uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v nBufferLength=%v lpBuffer=%v, (%v:)", t, nBufferLength, lpBuffer, origin(2))
	}
	rv, err := syscall.GetTempPath(nBufferLength, (*uint16)(unsafe.Pointer(lpBuffer)))
	if err != nil {
		t.setErrno(err)
	}
	return rv
}

func XGetTickCount(t *TLS) uint32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, _ := syscall.Syscall(procGetTickCount.Addr(), 0, 0, 0, 0)
	return uint32(r0)
}

func XGetVersionExA(t *TLS, lpVersionInformation uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpVersionInformation=%v, (%v:)", t, lpVersionInformation, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetVersionExA.Addr(), 1, lpVersionInformation, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XHeapCreate(t *TLS, flOptions uint32, dwInitialSize, dwMaximumSize types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v flOptions=%v dwMaximumSize=%v, (%v:)", t, flOptions, dwMaximumSize, origin(2))
	}
	panic(todo(""))
}

func XHeapDestroy(t *TLS, hHeap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hHeap=%v, (%v:)", t, hHeap, origin(2))
	}
	panic(todo(""))
}

func XHeapReAlloc(t *TLS, hHeap uintptr, dwFlags uint32, lpMem uintptr, dwBytes types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v hHeap=%v dwFlags=%v lpMem=%v dwBytes=%v, (%v:)", t, hHeap, dwFlags, lpMem, dwBytes, origin(2))
	}
	panic(todo(""))
}

func XHeapSize(t *TLS, hHeap uintptr, dwFlags uint32, lpMem uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v hHeap=%v dwFlags=%v lpMem=%v, (%v:)", t, hHeap, dwFlags, lpMem, origin(2))
	}
	panic(todo(""))
}

func XHeapValidate(t *TLS, hHeap uintptr, dwFlags uint32, lpMem uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hHeap=%v dwFlags=%v lpMem=%v, (%v:)", t, hHeap, dwFlags, lpMem, origin(2))
	}
	panic(todo(""))
}

func XHeapCompact(t *TLS, hHeap uintptr, dwFlags uint32) types.Size_t {
	if __ccgo_strace {
		trc("t=%v hHeap=%v dwFlags=%v, (%v:)", t, hHeap, dwFlags, origin(2))
	}
	panic(todo(""))
}

func XLoadLibraryA(t *TLS, lpLibFileName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpLibFileName=%v, (%v:)", t, lpLibFileName, origin(2))
	}
	panic(todo(""))
}

func XLoadLibraryW(t *TLS, lpLibFileName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpLibFileName=%v, (%v:)", t, lpLibFileName, origin(2))
	}
	panic(todo(""))
}

func XLocalFree(t *TLS, hMem uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v hMem=%v, (%v:)", t, hMem, origin(2))
	}
	h, err := syscall.LocalFree(syscall.Handle(hMem))
	if h != 0 {
		if err != nil {
			t.setErrno(err)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return uintptr(h)
	}
	return 0
}

func XLockFile(t *TLS, hFile uintptr, dwFileOffsetLow, dwFileOffsetHigh, nNumberOfBytesToLockLow, nNumberOfBytesToLockHigh uint32) int32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v nNumberOfBytesToLockHigh=%v, (%v:)", t, hFile, nNumberOfBytesToLockHigh, origin(2))
	}

	r1, _, e1 := syscall.Syscall6(procLockFile.Addr(), 5,
		hFile, uintptr(dwFileOffsetLow), uintptr(dwFileOffsetHigh), uintptr(nNumberOfBytesToLockLow), uintptr(nNumberOfBytesToLockHigh), 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)

}

func XLockFileEx(t *TLS, hFile uintptr, dwFlags, dwReserved, nNumberOfBytesToLockLow, nNumberOfBytesToLockHigh uint32, lpOverlapped uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v nNumberOfBytesToLockHigh=%v lpOverlapped=%v, (%v:)", t, hFile, nNumberOfBytesToLockHigh, lpOverlapped, origin(2))
	}
	r1, _, e1 := syscall.Syscall6(procLockFileEx.Addr(), 6,
		hFile, uintptr(dwFlags), uintptr(dwReserved), uintptr(nNumberOfBytesToLockLow), uintptr(nNumberOfBytesToLockHigh), lpOverlapped)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)
}

func XMapViewOfFile(t *TLS, hFileMappingObject uintptr, dwDesiredAccess, dwFileOffsetHigh, dwFileOffsetLow uint32, dwNumberOfBytesToMap types.Size_t) uintptr {
	if __ccgo_strace {
		trc("t=%v hFileMappingObject=%v dwFileOffsetLow=%v dwNumberOfBytesToMap=%v, (%v:)", t, hFileMappingObject, dwFileOffsetLow, dwNumberOfBytesToMap, origin(2))
	}
	h, _, e1 := syscall.Syscall6(procMapViewOfFile.Addr(), 5, hFileMappingObject, uintptr(dwDesiredAccess),
		uintptr(dwFileOffsetHigh), uintptr(dwFileOffsetLow), uintptr(dwNumberOfBytesToMap), 0)
	if h == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return h
}

func XQueryPerformanceCounter(t *TLS, lpPerformanceCount uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpPerformanceCount=%v, (%v:)", t, lpPerformanceCount, origin(2))
	}
	r0, _, _ := syscall.Syscall(procQueryPerformanceCounter.Addr(), 1, lpPerformanceCount, 0, 0)
	return int32(r0)
}

func XSleep(t *TLS, dwMilliseconds uint32) {
	if __ccgo_strace {
		trc("t=%v dwMilliseconds=%v, (%v:)", t, dwMilliseconds, origin(2))
	}
	gotime.Sleep(gotime.Duration(dwMilliseconds) * gotime.Millisecond)
}

func XSystemTimeToFileTime(t *TLS, lpSystemTime, lpFileTime uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileTime=%v, (%v:)", t, lpFileTime, origin(2))
	}
	r0, _, _ := syscall.Syscall(procSystemTimeToFileTime.Addr(), 2, lpSystemTime, lpFileTime, 0)
	return int32(r0)
}

func XUnlockFile(t *TLS, hFile uintptr, dwFileOffsetLow, dwFileOffsetHigh, nNumberOfBytesToUnlockLow, nNumberOfBytesToUnlockHigh uint32) int32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v nNumberOfBytesToUnlockHigh=%v, (%v:)", t, hFile, nNumberOfBytesToUnlockHigh, origin(2))
	}
	r1, _, e1 := syscall.Syscall6(procUnlockFile.Addr(), 5,
		hFile, uintptr(dwFileOffsetLow), uintptr(dwFileOffsetHigh), uintptr(nNumberOfBytesToUnlockLow), uintptr(nNumberOfBytesToUnlockHigh), 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)
}

func XUnlockFileEx(t *TLS, hFile uintptr, dwReserved, nNumberOfBytesToUnlockLow, nNumberOfBytesToUnlockHigh uint32, lpOverlapped uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hFile=%v nNumberOfBytesToUnlockHigh=%v lpOverlapped=%v, (%v:)", t, hFile, nNumberOfBytesToUnlockHigh, lpOverlapped, origin(2))
	}
	r1, _, e1 := syscall.Syscall6(procUnlockFileEx.Addr(), 5,
		hFile, uintptr(dwReserved), uintptr(nNumberOfBytesToUnlockLow), uintptr(nNumberOfBytesToUnlockHigh), lpOverlapped, 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return 0
	}
	return int32(r1)
}

func XUnmapViewOfFile(t *TLS, lpBaseAddress uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBaseAddress=%v, (%v:)", t, lpBaseAddress, origin(2))
	}
	err := syscall.UnmapViewOfFile(lpBaseAddress)
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XWideCharToMultiByte(t *TLS, CodePage uint32, dwFlags uint32, lpWideCharStr uintptr, cchWideChar int32, lpMultiByteStr uintptr, cbMultiByte int32, lpDefaultChar, lpUsedDefaultChar uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v CodePage=%v dwFlags=%v lpWideCharStr=%v cchWideChar=%v lpMultiByteStr=%v cbMultiByte=%v lpUsedDefaultChar=%v, (%v:)", t, CodePage, dwFlags, lpWideCharStr, cchWideChar, lpMultiByteStr, cbMultiByte, lpUsedDefaultChar, origin(2))
	}
	r1, _, _ := syscall.Syscall9(procWideCharToMultiByte.Addr(), 8,
		uintptr(CodePage), uintptr(dwFlags), lpWideCharStr,
		uintptr(cchWideChar), lpMultiByteStr, uintptr(cbMultiByte),
		lpDefaultChar, lpUsedDefaultChar, 0)
	return (int32(r1))
}

func XOutputDebugStringA(t *TLS, lpOutputString uintptr) {
	if __ccgo_strace {
		trc("t=%v lpOutputString=%v, (%v:)", t, lpOutputString, origin(2))
	}
	panic(todo(""))
}

func XFlushViewOfFile(t *TLS, lpBaseAddress uintptr, dwNumberOfBytesToFlush types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v lpBaseAddress=%v dwNumberOfBytesToFlush=%v, (%v:)", t, lpBaseAddress, dwNumberOfBytesToFlush, origin(2))
	}
	err := syscall.FlushViewOfFile(lpBaseAddress, uintptr(dwNumberOfBytesToFlush))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

type _ino_t = uint16
type _dev_t = uint32
type _stat64 = struct {
	Fst_dev   _dev_t
	Fst_ino   _ino_t
	Fst_mode  uint16
	Fst_nlink int16
	Fst_uid   int16
	Fst_gid   int16
	_         [2]byte
	Fst_rdev  _dev_t
	_         [4]byte
	Fst_size  int64
	Fst_atime int64
	Fst_mtime int64
	Fst_ctime int64
}

var (
	Windows_Tick   int64 = 10000000
	SecToUnixEpoch int64 = 11644473600
)

func WindowsTickToUnixSeconds(windowsTicks int64) int64 {
	return (windowsTicks/Windows_Tick - SecToUnixEpoch)
}

func X_stat64(t *TLS, path, buffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v buffer=%v, (%v:)", t, buffer, origin(2))
	}

	var fa syscall.Win32FileAttributeData
	r1, _, e1 := syscall.Syscall(procGetFileAttributesExA.Addr(), 3, path, syscall.GetFileExInfoStandard, (uintptr)(unsafe.Pointer(&fa)))
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
		return -1
	}

	var bStat64 = (*_stat64)(unsafe.Pointer(buffer))
	var accessTime = int64(fa.LastAccessTime.HighDateTime)<<32 + int64(fa.LastAccessTime.LowDateTime)
	bStat64.Fst_atime = WindowsTickToUnixSeconds(accessTime)
	var modTime = int64(fa.LastWriteTime.HighDateTime)<<32 + int64(fa.LastWriteTime.LowDateTime)
	bStat64.Fst_mtime = WindowsTickToUnixSeconds(modTime)
	var crTime = int64(fa.CreationTime.HighDateTime)<<32 + int64(fa.CreationTime.LowDateTime)
	bStat64.Fst_ctime = WindowsTickToUnixSeconds(crTime)
	var fSz = int64(fa.FileSizeHigh)<<32 + int64(fa.FileSizeLow)
	bStat64.Fst_size = fSz
	bStat64.Fst_mode = WindowsAttrbiutesToStat(fa.FileAttributes)

	return 0
}

func WindowsAttrbiutesToStat(fa uint32) uint16 {
	var src_mode = fa & 0xff
	var st_mode uint16
	if (src_mode & syscall.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		st_mode = syscall.S_IFDIR
	} else {
		st_mode = syscall.S_IFREG
	}

	if src_mode&syscall.FILE_ATTRIBUTE_READONLY != 0 {
		st_mode = st_mode | syscall.S_IRUSR
	} else {
		st_mode = st_mode | syscall.S_IRUSR | syscall.S_IWUSR
	}

	st_mode = st_mode | (st_mode&0x700)>>3
	st_mode = st_mode | (st_mode&0x700)>>6
	return st_mode
}

func X_chsize(t *TLS, fd int32, size long) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v size=%v, (%v:)", t, fd, size, origin(2))
	}

	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(EBADF)
		return -1
	}

	err := syscall.Ftruncate(f.Handle, int64(size))
	if err != nil {
		t.setErrno(err)
		return -1
	}

	return 0
}

func X_snprintf(t *TLS, str uintptr, size types.Size_t, format, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v str=%v size=%v args=%v, (%v:)", t, str, size, args, origin(2))
	}
	return Xsnprintf(t, str, size, format, args)
}

const wErr_ERROR_INSUFFICIENT_BUFFER = 122

func win32FindDataToFileInfo(t *TLS, fdata *stat.X_finddata64i32_t, wfd *syscall.Win32finddata) int32 {
	var accessTime = int64(wfd.LastAccessTime.HighDateTime)<<32 + int64(wfd.LastAccessTime.LowDateTime)
	fdata.Ftime_access = WindowsTickToUnixSeconds(accessTime)
	var modTime = int64(wfd.LastWriteTime.HighDateTime)<<32 + int64(wfd.LastWriteTime.LowDateTime)
	fdata.Ftime_write = WindowsTickToUnixSeconds(modTime)
	var crTime = int64(wfd.CreationTime.HighDateTime)<<32 + int64(wfd.CreationTime.LowDateTime)
	fdata.Ftime_create = WindowsTickToUnixSeconds(crTime)

	fdata.Fsize = wfd.FileSizeLow
	fdata.Fattrib = wfd.FileAttributes

	var cp = XGetConsoleCP(t)
	var wcFn = (uintptr)(unsafe.Pointer(&wfd.FileName[0]))
	var mbcsFn = (uintptr)(unsafe.Pointer(&fdata.Fname[0]))
	rv := XWideCharToMultiByte(t, cp, 0, wcFn, -1, mbcsFn, 260, 0, 0)
	if rv == wErr_ERROR_INSUFFICIENT_BUFFER {
		t.setErrno(errno.ENOMEM)
		return -1
	}
	return 0
}

func X_findfirst64i32(t *TLS, filespec, fileinfo uintptr) types.Intptr_t {
	if __ccgo_strace {
		trc("t=%v fileinfo=%v, (%v:)", t, fileinfo, origin(2))
	}

	var gsFileSpec = GoString(filespec)
	namep, err := syscall.UTF16PtrFromString(gsFileSpec)
	if err != nil {
		t.setErrno(err)
		return types.Intptr_t(-1)
	}

	var fdata = (*stat.X_finddata64i32_t)(unsafe.Pointer(fileinfo))
	var wfd syscall.Win32finddata
	h, err := syscall.FindFirstFile((*uint16)(unsafe.Pointer(namep)), &wfd)
	if err != nil {
		t.setErrno(err)
		return types.Intptr_t(-1)
	}
	rv := win32FindDataToFileInfo(t, fdata, &wfd)
	if rv != 0 {
		if h != 0 {
			syscall.FindClose(h)
		}
		return types.Intptr_t(-1)
	}
	return types.Intptr_t(h)
}

func X_findnext64i32(t *TLS, handle types.Intptr_t, fileinfo uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v handle=%v fileinfo=%v, (%v:)", t, handle, fileinfo, origin(2))
	}

	var fdata = (*stat.X_finddata64i32_t)(unsafe.Pointer(fileinfo))
	var wfd syscall.Win32finddata

	err := syscall.FindNextFile(syscall.Handle(handle), &wfd)
	if err != nil {
		t.setErrno(err)
		return -1
	}

	rv := win32FindDataToFileInfo(t, fdata, &wfd)
	if rv != 0 {
		return -1
	}
	return 0
}

func X_findclose(t *TLS, handle types.Intptr_t) int32 {
	if __ccgo_strace {
		trc("t=%v handle=%v, (%v:)", t, handle, origin(2))
	}

	err := syscall.FindClose(syscall.Handle(handle))
	if err != nil {
		t.setErrno(err)
		return -1
	}
	return 0
}

func XGetEnvironmentVariableA(t *TLS, lpName, lpBuffer uintptr, nSize uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nSize=%v, (%v:)", t, lpBuffer, nSize, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procGetEnvironmentVariableA.Addr(), 3, lpName, lpBuffer, uintptr(nSize))
	n := uint32(r0)
	if n == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return n
}

func X_fstat64(t *TLS, fd int32, buffer uintptr) int32 {
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

	var bStat64 = (*_stat64)(unsafe.Pointer(buffer))
	var accessTime = int64(d.LastAccessTime.HighDateTime)<<32 + int64(d.LastAccessTime.LowDateTime)
	bStat64.Fst_atime = WindowsTickToUnixSeconds(accessTime)
	var modTime = int64(d.LastWriteTime.HighDateTime)<<32 + int64(d.LastWriteTime.LowDateTime)
	bStat64.Fst_mtime = WindowsTickToUnixSeconds(modTime)
	var crTime = int64(d.CreationTime.HighDateTime)<<32 + int64(d.CreationTime.LowDateTime)
	bStat64.Fst_ctime = WindowsTickToUnixSeconds(crTime)
	var fSz = int64(d.FileSizeHigh)<<32 + int64(d.FileSizeLow)
	bStat64.Fst_size = fSz
	bStat64.Fst_mode = WindowsAttrbiutesToStat(d.FileAttributes)

	return 0
}

func XCreateEventA(t *TLS, lpEventAttributes uintptr, bManualReset, bInitialState int32, lpName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpEventAttributes=%v bInitialState=%v lpName=%v, (%v:)", t, lpEventAttributes, bInitialState, lpName, origin(2))
	}
	r0, _, err := syscall.Syscall6(procCreateEventA.Addr(), 4, lpEventAttributes, uintptr(bManualReset),
		uintptr(bInitialState), lpName, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func XCancelSynchronousIo(t *TLS, hThread uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hThread=%v, (%v:)", t, hThread, origin(2))
	}
	panic(todo(""))
}

func X_endthreadex(t *TLS, _ ...interface{}) {
}

func X_beginthread(t *TLS, procAddr uintptr, stack_sz uint32, args uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v procAddr=%v stack_sz=%v args=%v, (%v:)", t, procAddr, stack_sz, args, origin(2))
	}
	f := (*struct{ f func(*TLS, uintptr) uint32 })(unsafe.Pointer(&struct{ uintptr }{procAddr})).f
	var tAdp = ThreadAdapter{threadFunc: f, tls: NewTLS(), param: args}
	tAdp.token = addObject(&tAdp)

	r0, _, err := syscall.Syscall6(procCreateThread.Addr(), 6, 0, uintptr(stack_sz),
		threadCallback, tAdp.token, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_beginthreadex(t *TLS, _ uintptr, stack_sz uint32, procAddr uintptr, args uintptr, initf uint32, thAddr uintptr) int32 {
	f := (*struct{ f func(*TLS, uintptr) uint32 })(unsafe.Pointer(&struct{ uintptr }{procAddr})).f
	var tAdp = ThreadAdapter{threadFunc: f, tls: NewTLS(), param: args}
	tAdp.token = addObject(&tAdp)

	r0, _, err := syscall.Syscall6(procCreateThread.Addr(), 6, 0, uintptr(stack_sz),
		threadCallback, tAdp.token, uintptr(initf), thAddr)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetCurrentThreadId(t *TLS) uint32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, _ := syscall.Syscall(procGetCurrentThreadId.Addr(), 0, 0, 0, 0)
	return uint32(r0)

}

func XGetExitCodeThread(t *TLS, hThread, lpExitCode uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpExitCode=%v, (%v:)", t, lpExitCode, origin(2))
	}
	r0, _, _ := syscall.Syscall(procGetExitCodeThread.Addr(), 2, hThread, lpExitCode, 0)
	return int32(r0)
}

func XWaitForSingleObjectEx(t *TLS, hHandle uintptr, dwMilliseconds uint32, bAlertable int32) uint32 {
	if __ccgo_strace {
		trc("t=%v hHandle=%v dwMilliseconds=%v bAlertable=%v, (%v:)", t, hHandle, dwMilliseconds, bAlertable, origin(2))
	}
	rv, _, _ := syscall.Syscall(procWaitForSingleObjectEx.Addr(), 3, hHandle, uintptr(dwMilliseconds), uintptr(bAlertable))
	return uint32(rv)
}

func XMsgWaitForMultipleObjectsEx(t *TLS, nCount uint32, pHandles uintptr, dwMilliseconds, dwWakeMask, dwFlags uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v nCount=%v pHandles=%v dwFlags=%v, (%v:)", t, nCount, pHandles, dwFlags, origin(2))
	}
	r0, _, err := syscall.Syscall6(procMsgWaitForMultipleObjectsEx.Addr(), 5,
		uintptr(nCount),
		pHandles,
		uintptr(dwMilliseconds),
		uintptr(dwWakeMask),
		uintptr(dwFlags),
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return uint32(r0)
}

func XMessageBoxW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetModuleFileNameW(t *TLS, hModule, lpFileName uintptr, nSize uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v nSize=%v, (%v:)", t, lpFileName, nSize, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetModuleFileNameW.Addr(), 3, hModule, lpFileName, uintptr(nSize))
	if r0 == 0 {
		t.setErrno(err)
	}
	return uint32(r0)
}

func XNetGetDCName(t *TLS, ServerName, DomainName, Buffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v Buffer=%v, (%v:)", t, Buffer, origin(2))
	}
	r0, _, err := syscall.Syscall(procNetGetDCName.Addr(), 3, ServerName, DomainName, Buffer)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XNetUserGetInfo(t *TLS, servername, username uintptr, level uint32, bufptr uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v username=%v level=%v bufptr=%v, (%v:)", t, username, level, bufptr, origin(2))
	}
	r0, _, err := syscall.Syscall6(procNetUserGetInfo.Addr(), 4,
		servername,
		username,
		uintptr(level),
		bufptr,
		0,
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return uint32(r0)
}

func XlstrlenW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetProfilesDirectoryW(t *TLS, lpProfileDir, lpcchSize uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpcchSize=%v, (%v:)", t, lpcchSize, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetProfilesDirectoryW.Addr(), 2, lpProfileDir, lpcchSize, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XNetApiBufferFree(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetPrivateProfileStringA(t *TLS, lpAppName, lpKeyName, lpDefault, lpReturnedString uintptr, nSize uint32, lpFileName uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v lpReturnedString=%v nSize=%v lpFileName=%v, (%v:)", t, lpReturnedString, nSize, lpFileName, origin(2))
	}
	r0, _, err := syscall.Syscall6(procGetPrivateProfileStringA.Addr(), 4,
		lpAppName,
		lpKeyName,
		lpDefault,
		lpReturnedString,
		uintptr(nSize),
		lpFileName,
	)
	if err != 0 {
		t.setErrno(0x02)
	}
	return uint32(r0)
}

func XGetWindowsDirectoryA(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetFileSecurityW(t *TLS, lpFileName uintptr, RequestedInformation uint32, pSecurityDescriptor uintptr, nLength uint32, lpnLengthNeeded uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v RequestedInformation=%v pSecurityDescriptor=%v nLength=%v lpnLengthNeeded=%v, (%v:)", t, lpFileName, RequestedInformation, pSecurityDescriptor, nLength, lpnLengthNeeded, origin(2))
	}
	r0, _, err := syscall.Syscall6(procGetFileSecurityW.Addr(), 5, lpFileName, uintptr(RequestedInformation), pSecurityDescriptor, uintptr(nLength), lpnLengthNeeded, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetSecurityDescriptorOwner(t *TLS, pSecurityDescriptor, pOwner, lpbOwnerDefaulted uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpbOwnerDefaulted=%v, (%v:)", t, lpbOwnerDefaulted, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetSecurityDescriptorOwner.Addr(), 3, pSecurityDescriptor, pOwner, lpbOwnerDefaulted)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)

}

func XGetSidIdentifierAuthority(t *TLS, pSid uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v pSid=%v, (%v:)", t, pSid, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetSidIdentifierAuthority.Addr(), 1, pSid, 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return r0
}

func XImpersonateSelf(t *TLS, ImpersonationLevel int32) int32 {
	if __ccgo_strace {
		trc("t=%v ImpersonationLevel=%v, (%v:)", t, ImpersonationLevel, origin(2))
	}
	r0, _, err := syscall.Syscall(procImpersonateSelf.Addr(), 1, uintptr(ImpersonationLevel), 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XOpenThreadToken(t *TLS, ThreadHandle uintptr, DesiredAccess uint32, OpenAsSelf int32, TokenHandle uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ThreadHandle=%v DesiredAccess=%v OpenAsSelf=%v TokenHandle=%v, (%v:)", t, ThreadHandle, DesiredAccess, OpenAsSelf, TokenHandle, origin(2))
	}
	r0, _, err := syscall.Syscall6(procOpenThreadToken.Addr(), 4, ThreadHandle, uintptr(DesiredAccess), uintptr(OpenAsSelf), TokenHandle, 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetCurrentThread(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetCurrentThread.Addr(), 0, 0, 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return r0
}

func XRevertToSelf(t *TLS) int32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, err := syscall.Syscall(procRevertToSelf.Addr(), 0, 0, 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XAccessCheck(t *TLS, pSecurityDescriptor, ClientToken uintptr, DesiredAccess uint32, GenericMapping, PrivilegeSet, PrivilegeSetLength, GrantedAccess, AccessStatus uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ClientToken=%v DesiredAccess=%v AccessStatus=%v, (%v:)", t, ClientToken, DesiredAccess, AccessStatus, origin(2))
	}
	r0, _, err := syscall.Syscall9(procAccessCheck.Addr(), 8,
		pSecurityDescriptor,
		ClientToken,
		uintptr(DesiredAccess),
		GenericMapping,
		PrivilegeSet,
		PrivilegeSetLength,
		GrantedAccess,
		AccessStatus,
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func Xwcsicmp(t *TLS, string1, string2 uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v string2=%v, (%v:)", t, string2, origin(2))
	}
	var s1 = strings.ToLower(goWideString(string1))
	var s2 = strings.ToLower(goWideString(string2))
	return int32(strings.Compare(s1, s2))
}

func XSetCurrentDirectoryW(t *TLS, lpPathName uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpPathName=%v, (%v:)", t, lpPathName, origin(2))
	}
	err := syscall.SetCurrentDirectory((*uint16)(unsafe.Pointer(lpPathName)))
	if err != nil {
		t.setErrno(err)
		return 0
	}
	return 1
}

func XGetCurrentDirectoryW(t *TLS, nBufferLength uint32, lpBuffer uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v nBufferLength=%v lpBuffer=%v, (%v:)", t, nBufferLength, lpBuffer, origin(2))
	}
	n, err := syscall.GetCurrentDirectory(nBufferLength, (*uint16)(unsafe.Pointer(lpBuffer)))
	if err != nil {
		t.setErrno(err)
	}
	return n
}

func XGetFileInformationByHandle(t *TLS, hFile, lpFileInformation uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileInformation=%v, (%v:)", t, lpFileInformation, origin(2))
	}
	r1, _, e1 := syscall.Syscall(procGetFileInformationByHandle.Addr(), 2, hFile, lpFileInformation, 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return int32(r1)
}

func XGetVolumeInformationW(t *TLS, lpRootPathName, lpVolumeNameBuffer uintptr, nVolumeNameSize uint32, lpVolumeSerialNumber, lpMaximumComponentLength, lpFileSystemFlags, lpFileSystemNameBuffer uintptr, nFileSystemNameSize uint32) int32 {
	if __ccgo_strace {
		trc("t=%v lpVolumeNameBuffer=%v nVolumeNameSize=%v lpFileSystemNameBuffer=%v nFileSystemNameSize=%v, (%v:)", t, lpVolumeNameBuffer, nVolumeNameSize, lpFileSystemNameBuffer, nFileSystemNameSize, origin(2))
	}
	r0, _, err := syscall.Syscall9(procGetVolumeInformationW.Addr(), 8,
		lpRootPathName,
		lpVolumeNameBuffer,
		uintptr(nVolumeNameSize),
		lpVolumeSerialNumber,
		lpMaximumComponentLength,
		lpFileSystemFlags,
		lpFileSystemNameBuffer,
		uintptr(nFileSystemNameSize),
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func Xwcschr(t *TLS, str uintptr, c wchar_t) uintptr {
	if __ccgo_strace {
		trc("t=%v str=%v c=%v, (%v:)", t, str, c, origin(2))
	}
	var source = str
	for {
		var buf = *(*uint16)(unsafe.Pointer(source))
		if buf == 0 {
			return 0
		}
		if buf == c {
			return source
		}

		source++
		source++
	}
}

func XSetFileTime(t *TLS, _hFile uintptr, _lpCreationTime, _lpLastAccessTime, _lpLastWriteTime uintptr) (r int32) {
	if __ccgo_strace {
		trc("hFile=%+v lpCreationTime=%+v lpLastAccessTime=%+v lpLastWriteTime=%+v", _hFile, _lpCreationTime, _lpLastAccessTime, _lpLastWriteTime)
		defer func() { trc(`XSetFileTime->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procSetFileTime.Addr(), _hFile, _lpCreationTime, _lpLastAccessTime, _lpLastWriteTime)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		t.SetLastError(uint32(err))
	}
	return int32(r0)
}

func XGetNamedSecurityInfoW(tls *TLS, _pObjectName uintptr, _ObjectType int32, _SecurityInfo uint32, _ppsidOwner uintptr, _ppsidGroup uintptr, _ppDacl uintptr, _ppSacl uintptr, _ppSecurityDescriptor uintptr) (r uint32) {
	if __ccgo_strace {
		trc("pObjectName=%+v ObjectType=%+v SecurityInfo=%+v ppsidOwner=%+v ppsidGroup=%+v ppDacl=%+v ppSacl=%+v ppSecurityDescriptor=%+v", _pObjectName, _ObjectType, _SecurityInfo, _ppsidOwner, _ppsidGroup, _ppDacl, _ppSacl, _ppSecurityDescriptor)
		defer func() { trc(`XGetNamedSecurityInfoW->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procGetNamedSecurityInfoW.Addr(), _pObjectName, uintptr(_ObjectType), uintptr(_SecurityInfo), _ppsidOwner, _ppsidGroup, _ppDacl, _ppSacl, _ppSecurityDescriptor)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return uint32(r0)
}

func XOpenProcessToken(tls *TLS, _ProcessHandle uintptr, _DesiredAccess uint32, _TokenHandle uintptr) (r int32) {
	if __ccgo_strace {
		trc("ProcessHandle=%+v DesiredAccess=%+v TokenHandle=%+v", _ProcessHandle, _DesiredAccess, _TokenHandle)
		defer func() { trc(`XOpenProcessToken->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procOpenProcessToken.Addr(), _ProcessHandle, uintptr(_DesiredAccess), _TokenHandle)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return int32(r0)
}

func XGetTokenInformation(tls *TLS, _TokenHandle uintptr, _TokenInformationClass int32, _TokenInformation uintptr, _TokenInformationLength uint32, _ReturnLength uintptr) (r int32) {
	if __ccgo_strace {
		trc("TokenHandle=%+v TokenInformationClass=%+v TokenInformation=%+v TokenInformationLength=%+v ReturnLength=%+v", _TokenHandle, _TokenInformationClass, _TokenInformation, _TokenInformationLength, _ReturnLength)
		defer func() { trc(`XGetTokenInformation->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procGetTokenInformation.Addr(), _TokenHandle, uintptr(_TokenInformationClass), _TokenInformation, uintptr(_TokenInformationLength), _ReturnLength)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return int32(r0)
}

func XEqualSid(tls *TLS, _pSid1 uintptr, _pSid2 uintptr) (r int32) {
	if __ccgo_strace {
		trc("pSid1=%+v pSid2=%+v", _pSid1, _pSid2)
		defer func() { trc(`XEqualSid->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procEqualSid.Addr(), _pSid1, _pSid2)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return int32(r0)
}

func XWSAStartup(t *TLS, wVersionRequired uint16, lpWSAData uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v wVersionRequired=%v lpWSAData=%v, (%v:)", t, wVersionRequired, lpWSAData, origin(2))
	}
	r0, _, _ := syscall.Syscall(procWSAStartup.Addr(), 2, uintptr(wVersionRequired), lpWSAData, 0)
	if r0 != 0 {
		t.setErrno(r0)
	}
	return int32(r0)
}

func XGetModuleHandleA(t *TLS, lpModuleName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpModuleName=%v, (%v:)", t, lpModuleName, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetModuleHandleA.Addr(), 1, lpModuleName, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func XGetModuleHandleW(t *TLS, lpModuleName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v lpModuleName=%v, (%v:)", t, lpModuleName, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetModuleHandleW.Addr(), 1, lpModuleName, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func XGetEnvironmentVariableW(t *TLS, lpName, lpBuffer uintptr, nSize uint32) uint32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nSize=%v, (%v:)", t, lpBuffer, nSize, origin(2))
	}
	r0, _, e1 := syscall.Syscall(procGetEnvironmentVariableW.Addr(), 3, lpName, lpBuffer, uintptr(nSize))
	n := uint32(r0)
	if n == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return n
}

func XlstrcmpiA(t *TLS, lpString1, lpString2 uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpString2=%v, (%v:)", t, lpString2, origin(2))
	}
	var s1 = strings.ToLower(GoString(lpString1))
	var s2 = strings.ToLower(GoString(lpString2))
	return int32(strings.Compare(s1, s2))
}

func XGetModuleFileNameA(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetACP(t *TLS) uint32 {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	r0, _, _ := syscall.Syscall(procGetACP.Addr(), 0, 0, 0, 0)
	return uint32(r0)
}

func XGetUserNameW(t *TLS, lpBuffer, pcbBuffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pcbBuffer=%v, (%v:)", t, pcbBuffer, origin(2))
	}
	u, err := user.Current()
	if err != nil {
		panic(todo(""))
		return 0
	}

	wcnt := *(*uint16)(unsafe.Pointer(pcbBuffer))
	s := utf16.Encode([]rune(u.Username))
	if len(s)+1 > int(wcnt) {
		panic(todo(""))
	}

	*(*uint16)(unsafe.Pointer(pcbBuffer)) = uint16(len(s) + 1)
	for _, v := range s {
		*(*uint16)(unsafe.Pointer(lpBuffer)) = v
		lpBuffer += 2
	}
	return 1
}

func XLoadLibraryExW(t *TLS, lpLibFileName, hFile uintptr, dwFlags uint32) uintptr {
	if __ccgo_strace {
		trc("t=%v hFile=%v dwFlags=%v, (%v:)", t, hFile, dwFlags, origin(2))
	}
	return 0
}

func Xwcscpy(t *TLS, strDestination, strSource uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v strSource=%v, (%v:)", t, strSource, origin(2))
	}
	if strSource == 0 {
		return 0
	}

	d := strDestination
	for {
		c := *(*uint16)(unsafe.Pointer(strSource))
		strSource += 2
		*(*uint16)(unsafe.Pointer(d)) = c
		d += 2
		if c == 0 {
			return strDestination
		}
	}
}

func XwsprintfW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegisterClassW(t *TLS, lpWndClass uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpWndClass=%v, (%v:)", t, lpWndClass, origin(2))
	}
	r0, _, err := syscall.Syscall(procRegisterClassW.Addr(), 1, lpWndClass, 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XKillTimer(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDestroyWindow(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XUnregisterClassW(t *TLS, lpClassName, hInstance uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v hInstance=%v, (%v:)", t, hInstance, origin(2))
	}
	r0, _, err := syscall.Syscall(procUnregisterClassW.Addr(), 2, lpClassName, hInstance, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XPostMessageW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSetTimer(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XCreateWindowExW(t *TLS, dwExStyle uint32, lpClassName, lpWindowName uintptr, dwStyle uint32, x, y, nWidth, nHeight int32, hWndParent, hMenu, hInstance, lpParam uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v dwExStyle=%v lpWindowName=%v dwStyle=%v nHeight=%v lpParam=%v, (%v:)", t, dwExStyle, lpWindowName, dwStyle, nHeight, lpParam, origin(2))
	}
	r0, _, err := syscall.Syscall12(procCreateWindowExW.Addr(), 12,
		uintptr(dwExStyle),
		lpClassName,
		lpWindowName,
		uintptr(dwStyle),
		uintptr(x),
		uintptr(y),
		uintptr(nWidth),
		uintptr(nHeight),
		hWndParent,
		hMenu,
		hInstance,
		lpParam,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return r0
}

func XPeekMessageW(t *TLS, lpMsg, hWnd uintptr, wMsgFilterMin, wMsgFilterMax, wRemoveMsg uint32) int32 {
	if __ccgo_strace {
		trc("t=%v hWnd=%v wRemoveMsg=%v, (%v:)", t, hWnd, wRemoveMsg, origin(2))
	}
	r0, _, err := syscall.Syscall6(procPeekMessageW.Addr(), 5,
		lpMsg,
		hWnd,
		uintptr(wMsgFilterMin),
		uintptr(wMsgFilterMax),
		uintptr(wRemoveMsg),
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetMessageW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XPostQuitMessage(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XTranslateMessage(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDispatchMessageW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSleepEx(t *TLS, dwMilliseconds uint32, bAlertable int32) uint32 {
	if __ccgo_strace {
		trc("t=%v dwMilliseconds=%v bAlertable=%v, (%v:)", t, dwMilliseconds, bAlertable, origin(2))
	}
	r0, _, _ := syscall.Syscall(procSleepEx.Addr(), 2, uintptr(dwMilliseconds), uintptr(bAlertable), 0)
	return uint32(r0)
}

func XCreatePipe(t *TLS, hReadPipe, hWritePipe, lpPipeAttributes uintptr, nSize uint32) int32 {
	if __ccgo_strace {
		trc("t=%v lpPipeAttributes=%v nSize=%v, (%v:)", t, lpPipeAttributes, nSize, origin(2))
	}
	r0, _, err := syscall.Syscall6(procCreatePipe.Addr(), 4, hReadPipe, hWritePipe, lpPipeAttributes, uintptr(nSize), 0, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XCreateProcessW(t *TLS, lpApplicationName, lpCommandLine, lpProcessAttributes, lpThreadAttributes uintptr, bInheritHandles int32, dwCreationFlags uint32,
	lpEnvironment, lpCurrentDirectory, lpStartupInfo, lpProcessInformation uintptr) int32 {
	r1, _, e1 := syscall.Syscall12(procCreateProcessW.Addr(), 10, lpApplicationName, lpCommandLine, lpProcessAttributes, lpThreadAttributes,
		uintptr(bInheritHandles), uintptr(dwCreationFlags), lpEnvironment, lpCurrentDirectory, lpStartupInfo, lpProcessInformation, 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			t.setErrno(e1)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}

	return int32(r1)
}

func XWaitForInputIdle(t *TLS, hProcess uintptr, dwMilliseconds uint32) int32 {
	if __ccgo_strace {
		trc("t=%v hProcess=%v dwMilliseconds=%v, (%v:)", t, hProcess, dwMilliseconds, origin(2))
	}
	r0, _, _ := syscall.Syscall(procWaitForInputIdle.Addr(), 2, hProcess, uintptr(dwMilliseconds), 0)
	return int32(r0)
}

func XSearchPathW(t *TLS, lpPath, lpFileName, lpExtension uintptr, nBufferLength uint32, lpBuffer, lpFilePart uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpExtension=%v nBufferLength=%v lpFilePart=%v, (%v:)", t, lpExtension, nBufferLength, lpFilePart, origin(2))
	}
	r0, _, err := syscall.Syscall6(procSearchPathW.Addr(), 6, lpPath, lpFileName, lpExtension, uintptr(nBufferLength), lpBuffer, lpFilePart)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetShortPathNameW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetExitCodeProcess(t *TLS, hProcess, lpExitCode uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpExitCode=%v, (%v:)", t, lpExitCode, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetExitCodeProcess.Addr(), 2, hProcess, lpExitCode, 0)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XPeekNamedPipe(t *TLS, hNamedPipe, lpBuffer uintptr, nBufferSize uint32, lpBytesRead, lpTotalBytesAvail, lpBytesLeftThisMessage uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpBuffer=%v nBufferSize=%v lpBytesLeftThisMessage=%v, (%v:)", t, lpBuffer, nBufferSize, lpBytesLeftThisMessage, origin(2))
	}
	r0, _, err := syscall.Syscall6(procPeekNamedPipe.Addr(), 6, hNamedPipe, lpBuffer, uintptr(nBufferSize), lpBytesRead, lpTotalBytesAvail, lpBytesLeftThisMessage)
	if r0 == 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_InterlockedExchange(t *TLS, Target uintptr, Value long) long {
	if __ccgo_strace {
		trc("t=%v Target=%v Value=%v, (%v:)", t, Target, Value, origin(2))
	}
	old := atomic.SwapInt32((*int32)(unsafe.Pointer(Target)), Value)
	return old
}

func XTerminateThread(t *TLS, hThread uintptr, dwExitCode uint32) int32 {
	if __ccgo_strace {
		trc("t=%v hThread=%v dwExitCode=%v, (%v:)", t, hThread, dwExitCode, origin(2))
	}
	r0, _, err := syscall.Syscall(procTerminateThread.Addr(), 2, hThread, uintptr(dwExitCode), 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetComputerNameW(t *TLS, lpBuffer, nSize uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v nSize=%v, (%v:)", t, nSize, origin(2))
	}
	panic(todo(""))
}

func Xgethostname(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSendMessageW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XWSAGetLastError(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xclosesocket(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XWspiapiFreeAddrInfo(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XWspiapiGetNameInfo(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XIN6_ADDR_EQUAL(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X__ccgo_in6addr_anyp(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XIN6_IS_ADDR_V4MAPPED(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSetHandleInformation(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xioctlsocket(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGetWindowLongPtrW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XSetWindowLongPtrW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XWSAAsyncSelect(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func Xinet_ntoa(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func X_controlfp(t *TLS, _ ...interface{}) uint32 {
	panic(todo(""))
}

func XQueryPerformanceFrequency(t *TLS, lpFrequency uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFrequency=%v, (%v:)", t, lpFrequency, origin(2))
	}

	r1, _, err := syscall.Syscall(procQueryPerformanceFrequency.Addr(), 1, lpFrequency, 0, 0)
	if r1 == 0 {
		t.setErrno(err)
		return 0
	}
	return int32(r1)
}

func inDST(t gotime.Time) bool {
	jan1st := gotime.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())

	_, off1 := t.Zone()
	_, off2 := jan1st.Zone()

	return off1 != off2
}

func X_ftime(t *TLS, timeptr uintptr) {
	if __ccgo_strace {
		trc("t=%v timeptr=%v, (%v:)", t, timeptr, origin(2))
	}
	var tm = gotime.Now()
	var tPtr = (*time.X__timeb64)(unsafe.Pointer(timeptr))
	tPtr.Ftime = tm.Unix()
	tPtr.Fmillitm = uint16(gotime.Duration(tm.Nanosecond()) / gotime.Millisecond)
	if inDST(tm) {
		tPtr.Fdstflag = 1
	}
	_, offset := tm.Zone()
	tPtr.Ftimezone = int16(offset)
}

func XDdeInitializeW(t *TLS, _ ...interface{}) uint32 {
	panic(todo(""))
}

func XDdeCreateStringHandleW(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func XDdeNameService(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X_snwprintf(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeQueryStringW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X_wcsicmp(t *TLS, string1, string2 uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v string2=%v, (%v:)", t, string2, origin(2))
	}
	return Xwcsicmp(t, string1, string2)
}

func XDdeCreateDataHandle(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func XDdeAccessData(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func XDdeUnaccessData(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeUninitialize(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeConnect(t *TLS, _ ...interface{}) uintptr {
	panic(todo(""))
}

func XDdeFreeStringHandle(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegisterClassExW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGlobalGetAtomNameW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGlobalAddAtomW(t *TLS, _ ...interface{}) uint16 {
	panic(todo(""))
}

func XEnumWindows(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XIsWindow(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XGlobalDeleteAtom(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeGetLastError(t *TLS, _ ...interface{}) uint32 {
	panic(todo(""))
}

func XDdeClientTransaction(t *TLS, pData uintptr, cbData uint32, hConv uintptr, hszItem uintptr, wFmt, wType, dwTimeout uint32, pdwResult uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v pData=%v cbData=%v hConv=%v hszItem=%v dwTimeout=%v pdwResult=%v, (%v:)", t, pData, cbData, hConv, hszItem, dwTimeout, pdwResult, origin(2))
	}
	panic(todo(""))
}

func XDdeAbandonTransaction(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeFreeDataHandle(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeGetData(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XDdeDisconnect(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegCloseKey(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegDeleteValueW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegEnumKeyExW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegQueryValueExW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegEnumValueW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegConnectRegistryW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegCreateKeyExW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegOpenKeyExW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegDeleteKeyW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func XRegSetValueExW(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X__mingw_vsnwprintf(t *TLS, buffer uintptr, count types.Size_t, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v buffer=%v count=%v va=%v, (%v:)", t, buffer, count, va, origin(2))
	}
	panic(todo(""))
}

func X__mingw_vprintf(t *TLS, s, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	return Xvprintf(t, s, ap)
}

func X__mingw_vfscanf(t *TLS, stream, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	panic(todo(""))
}

func X__mingw_vsscanf(t *TLS, str, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	return Xsscanf(t, str, format, ap)
}

func X__mingw_vfprintf(t *TLS, f uintptr, format, va uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v f=%v va=%v, (%v:)", t, f, va, origin(2))
	}
	return Xvfprintf(t, f, format, va)
}

func X__mingw_vsprintf(t *TLS, s, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	return Xvsprintf(t, s, format, ap)
}

func X__mingw_vsnprintf(t *TLS, str uintptr, size types.Size_t, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v str=%v size=%v ap=%v, (%v:)", t, str, size, ap, origin(2))
	}
	return Xvsnprintf(t, str, size, format, ap)
}

func X_putchar(t *TLS, c int32) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v, (%v:)", t, c, origin(2))
	}
	if _, err := fwrite(unistd.STDOUT_FILENO, []byte{byte(c)}); err != nil {
		return -1
	}
	return int32(byte(c))
}

func X__mingw_vfwscanf(t *TLS, stream uintptr, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v ap=%v, (%v:)", t, stream, ap, origin(2))
	}
	panic(todo(""))
}

func X__mingw_vswscanf(t *TLS, stream uintptr, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v ap=%v, (%v:)", t, stream, ap, origin(2))
	}
	panic(todo(""))
}

func X__mingw_vfwprintf(t *TLS, stream, format, ap uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v ap=%v, (%v:)", t, ap, origin(2))
	}
	panic(todo(""))
}

func Xputchar(t *TLS, c int32) int32 {
	if __ccgo_strace {
		trc("t=%v c=%v, (%v:)", t, c, origin(2))
	}
	panic(todo(""))
}

func X_assert(t *TLS, message, filename uintptr, line uint32) {
	if __ccgo_strace {
		trc("t=%v filename=%v line=%v, (%v:)", t, filename, line, origin(2))
	}
	panic(todo(""))
}

func X_strdup(t *TLS, s uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v s=%v, (%v:)", t, s, origin(2))
	}
	panic(todo(""))
}

func X_access(t *TLS, pathname uintptr, mode int32) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}

	var path = GoString(pathname)

	info, err := os.Stat(path)
	if err != nil {
		return errno.ENOENT
	}

	switch mode {
	case 0:
		return 0
	case 2:
		if info.Mode().Perm()&(1<<(uint(7))) == 1 {
			return 0
		}
	case 4:
		if info.Mode().Perm()&(1<<(uint(7))) == 0 {
			return 0
		}
	case 6:
		if info.Mode().Perm()&(1<<(uint(7))) == 1 {
			return 0
		}
	}

	return errno.EACCES

}

func XSetConsoleCtrlHandler(t *TLS, HandlerRoutine uintptr, Add int32) int32 {
	if __ccgo_strace {
		trc("t=%v HandlerRoutine=%v Add=%v, (%v:)", t, HandlerRoutine, Add, origin(2))
	}

	return 0
}

func XDebugBreak(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func X_isatty(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}

	f, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return 0
	}

	if fd == unistd.STDOUT_FILENO ||
		fd == unistd.STDIN_FILENO ||
		fd == unistd.STDERR_FILENO {
		var mode uint32
		err := syscall.GetConsoleMode(f.Handle, &mode)
		if err != nil {
			t.setErrno(errno.EINVAL)
			return 0
		}

		return 1
	}

	return 0
}

func XSetConsoleTextAttribute(t *TLS, hConsoleOutput uintptr, wAttributes uint16) int32 {
	if __ccgo_strace {
		trc("t=%v hConsoleOutput=%v wAttributes=%v, (%v:)", t, hConsoleOutput, wAttributes, origin(2))
	}
	r1, _, _ := syscall.Syscall(procSetConsoleTextAttribute.Addr(), 2, hConsoleOutput, uintptr(wAttributes), 0)
	return int32(r1)
}

func XGetConsoleScreenBufferInfo(t *TLS, hConsoleOutput, lpConsoleScreenBufferInfo uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpConsoleScreenBufferInfo=%v, (%v:)", t, lpConsoleScreenBufferInfo, origin(2))
	}
	r1, _, _ := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, hConsoleOutput, lpConsoleScreenBufferInfo, 0)
	return int32(r1)
}

func X_popen(t *TLS, command, mode uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v mode=%v, (%v:)", t, mode, origin(2))
	}
	panic(todo(""))
}

func X_wunlink(t *TLS, filename uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v filename=%v, (%v:)", t, filename, origin(2))
	}
	panic(todo(""))
}

func Xclosedir(tls *TLS, dir uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v dir=%v, (%v:)", tls, dir, origin(2))
	}
	panic(todo(""))
}

func Xopendir(tls *TLS, name uintptr) uintptr {
	if __ccgo_strace {
		trc("tls=%v name=%v, (%v:)", tls, name, origin(2))
	}
	panic(todo(""))
}

func Xreaddir(tls *TLS, dir uintptr) uintptr {
	if __ccgo_strace {
		trc("tls=%v dir=%v, (%v:)", tls, dir, origin(2))
	}
	panic(todo(""))
}

func X_unlink(t *TLS, filename uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v filename=%v, (%v:)", t, filename, origin(2))
	}
	panic(todo(""))
}

func X_pclose(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	panic(todo(""))
}

func Xsetmode(t *TLS, fd, mode int32) int32 {
	if __ccgo_strace {
		trc("t=%v mode=%v, (%v:)", t, mode, origin(2))
	}
	return X_setmode(t, fd, mode)
}

func X_setmode(t *TLS, fd, mode int32) int32 {
	if __ccgo_strace {
		trc("t=%v mode=%v, (%v:)", t, mode, origin(2))
	}

	_, ok := fdToFile(fd)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}

	if mode == fcntl.O_BINARY {
		return fcntl.O_BINARY
	} else {
		t.setErrno(errno.EINVAL)
		return -1
	}
}

func X_mkdir(t *TLS, dirname uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v dirname=%v, (%v:)", t, dirname, origin(2))
	}
	panic(todo(""))
}

func X_chmod(t *TLS, filename uintptr, pmode int32) int32 {
	if __ccgo_strace {
		trc("t=%v filename=%v pmode=%v, (%v:)", t, filename, pmode, origin(2))
	}
	panic(todo(""))
}

func X_fileno(t *TLS, stream uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EBADF)
		return -1
	}
	return f._fd
}

func Xrewind(t *TLS, stream uintptr) {
	if __ccgo_strace {
		trc("t=%v stream=%v, (%v:)", t, stream, origin(2))
	}
	Xfseek(t, stream, 0, unistd.SEEK_SET)
}

func X__atomic_load_n(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func X__atomic_store_n(t *TLS, _ ...interface{}) int32 {
	panic(todo(""))
}

func X__builtin_add_overflow(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func X__builtin_mul_overflow(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func X__builtin_sub_overflow(t *TLS) {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	panic(todo(""))
}

func goWideBytes(p uintptr, n int) []uint16 {
	b := GoBytes(p, 2*n)
	var w []uint16
	for i := 0; i < len(b); i += 2 {
		w = append(w, *(*uint16)(unsafe.Pointer(&b[i])))
	}
	return w
}

func goWideString(p uintptr) string {
	if p == 0 {
		return ""
	}
	var w []uint16
	var raw = (*RawMem)(unsafe.Pointer(p))
	var i = 0
	for {
		wc := *(*uint16)(unsafe.Pointer(&raw[i]))
		w = append(w, wc)

		if wc == 0 {
			break
		}
		i = i + 2
	}
	s := utf16.Decode(w)
	return string(s)
}

func goWideStringN(p uintptr, n int) string {
	panic(todo(""))
}

func goWideStringNZ(p uintptr) string {
	if p == 0 {
		return ""
	}

	var w []uint16
	var raw = (*RawMem)(unsafe.Pointer(p))
	var i = 0
	for {
		wc := *(*uint16)(unsafe.Pointer(&raw[i]))
		if wc == 0 {
			break
		}

		w = append(w, wc)
		i = i + 2
	}
	s := utf16.Decode(w)
	return string(s)
}

func XGetCommandLineW(t *TLS) uintptr {
	if __ccgo_strace {
		trc("t=%v, (%v:)", t, origin(2))
	}
	return uintptr(unsafe.Pointer(syscall.GetCommandLine()))
}

func XAddAccessDeniedAce(t *TLS, pAcl uintptr, dwAceRevision, AccessMask uint32, pSid uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v pAcl=%v AccessMask=%v pSid=%v, (%v:)", t, pAcl, AccessMask, pSid, origin(2))
	}
	panic(todo(""))
}

func XAddAce(tls *TLS, _pAcl uintptr, _dwAceRevision uint32, _dwStartingAceIndex uint32, _pAceList uintptr, _nAceListLength uint32) (r uint32) {
	if __ccgo_strace {
		trc("pAcl=%+v dwAceRevision=%+v dwStartingAceIndex=%+v pAceList=%+v nAceListLength=%+v", _pAcl, _dwAceRevision, _dwStartingAceIndex, _pAceList, _nAceListLength)
		defer func() { trc(`XAddAce->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procAddAce.Addr(), _pAcl, uintptr(_dwAceRevision), uintptr(_dwStartingAceIndex), _pAceList, uintptr(_nAceListLength))
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return uint32(r0)
}

func XGetAce(tls *TLS, _pAcl uintptr, _dwAceIndex uint32, _pAce uintptr) (r int32) {
	if __ccgo_strace {
		trc("pAcl=%+v dwAceIndex=%+v pAce=%+v", _pAcl, _dwAceIndex, _pAce)
		defer func() { trc(`XGetAce->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procGetAce.Addr(), _pAcl, uintptr(_dwAceIndex), _pAce)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return int32(r0)
}

func XGetAclInformation(t *TLS, pAcl, pAclInformation uintptr, nAclInformationLength, dwAclInformationClass uint32) int32 {
	if __ccgo_strace {
		trc("t=%v pAclInformation=%v dwAclInformationClass=%v, (%v:)", t, pAclInformation, dwAclInformationClass, origin(2))
	}
	r0, _, err := syscall.Syscall6(procGetAclInformation.Addr(), 4,
		pAclInformation,
		pAclInformation,
		uintptr(nAclInformationLength),
		uintptr(dwAclInformationClass),
		0,
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetFileSecurityA(t *TLS, lpFileName uintptr, RequestedInformation uint32, pSecurityDescriptor uintptr, nLength uint32, lpnLengthNeeded uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpFileName=%v RequestedInformation=%v pSecurityDescriptor=%v nLength=%v lpnLengthNeeded=%v, (%v:)", t, lpFileName, RequestedInformation, pSecurityDescriptor, nLength, lpnLengthNeeded, origin(2))
	}
	r0, _, err := syscall.Syscall6(procGetFileSecurityA.Addr(), 5,
		lpFileName,
		uintptr(RequestedInformation),
		pSecurityDescriptor,
		uintptr(nLength),
		lpnLengthNeeded,
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetLengthSid(tls *TLS, _pSid uintptr) (r uint32) {
	if __ccgo_strace {
		trc("pSid=%+v", _pSid)
		defer func() { trc(`XGetLengthSid->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procGetLengthSid.Addr(), _pSid)
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return uint32(r0)
}

func XGetSecurityDescriptorDacl(t *TLS, pSecurityDescriptor, lpbDaclPresent, pDacl, lpbDaclDefaulted uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v lpbDaclDefaulted=%v, (%v:)", t, lpbDaclDefaulted, origin(2))
	}
	r0, _, err := syscall.Syscall6(procGetSecurityDescriptorDacl.Addr(), 4,
		pSecurityDescriptor,
		lpbDaclPresent,
		pDacl,
		lpbDaclDefaulted,
		0,
		0,
	)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetSidLengthRequired(t *TLS, nSubAuthorityCount uint8) int32 {
	if __ccgo_strace {
		trc("t=%v nSubAuthorityCount=%v, (%v:)", t, nSubAuthorityCount, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetSidLengthRequired.Addr(), 1, uintptr(nSubAuthorityCount), 0, 0)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetSidSubAuthority(t *TLS, pSid uintptr, nSubAuthority uint32) uintptr {
	if __ccgo_strace {
		trc("t=%v pSid=%v nSubAuthority=%v, (%v:)", t, pSid, nSubAuthority, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetSidSubAuthority.Addr(), 2, pSid, uintptr(nSubAuthority), 0)
	if err != 0 {
		t.setErrno(err)
	}
	return r0
}

func XInitializeAcl(tls *TLS, _pAcl uintptr, _nAclLength uint32, _dwAclRevision uint32) (r int32) {
	if __ccgo_strace {
		trc("pAcl=%+v nAclLength=%+v dwAclRevision=%+v", _pAcl, _nAclLength, _dwAclRevision)
		defer func() { trc(`XInitializeAcl->%+v`, r) }()
	}
	r0, r1, err := syscall.SyscallN(procInitializeAcl.Addr(), _pAcl, uintptr(_nAclLength), uintptr(_dwAclRevision))
	if err != 0 {
		if __ccgo_strace {
			trc(`r0=%v r1=%v err=%v`, r0, r1, err)
		}
		tls.SetLastError(uint32(err))
	}
	return int32(r0)
}

func XInitializeSid(t *TLS, Sid, pIdentifierAuthority uintptr, nSubAuthorityCount uint8) int32 {
	if __ccgo_strace {
		trc("t=%v pIdentifierAuthority=%v nSubAuthorityCount=%v, (%v:)", t, pIdentifierAuthority, nSubAuthorityCount, origin(2))
	}
	r0, _, err := syscall.Syscall(procInitializeSid.Addr(), 3, Sid, pIdentifierAuthority, uintptr(nSubAuthorityCount))
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XRaiseException(t *TLS, dwExceptionCode, dwExceptionFlags, nNumberOfArguments uint32, lpArguments uintptr) {
	if __ccgo_strace {
		trc("t=%v nNumberOfArguments=%v lpArguments=%v, (%v:)", t, nNumberOfArguments, lpArguments, origin(2))
	}
	panic(todo(""))
}

func XSetErrorMode(t *TLS, uMode uint32) int32 {
	if __ccgo_strace {
		trc("t=%v uMode=%v, (%v:)", t, uMode, origin(2))
	}
	panic(todo(""))
}

func XSetNamedSecurityInfoA(t *TLS, pObjectName uintptr, ObjectType, SecurityInfo uint32, psidOwner, psidGroup, pDacl, pSacl uintptr) uint32 {
	if __ccgo_strace {
		trc("t=%v pObjectName=%v SecurityInfo=%v pSacl=%v, (%v:)", t, pObjectName, SecurityInfo, pSacl, origin(2))
	}
	panic(todo(""))
}

func XCreateProcessA(t *TLS, lpApplicationName, lpCommandLine, lpProcessAttributes, lpThreadAttributes uintptr, bInheritHandles int32,
	dwCreationFlags uint32, lpEnvironment, lpCurrentDirectory, lpStartupInfo, lpProcessInformation uintptr) int32 {
	r1, _, err := syscall.Syscall12(procCreateProcessA.Addr(), 10, lpApplicationName, lpCommandLine, lpProcessAttributes, lpThreadAttributes,
		uintptr(bInheritHandles), uintptr(dwCreationFlags), lpEnvironment, lpCurrentDirectory, lpStartupInfo, lpProcessInformation, 0, 0)
	if r1 == 0 {
		if err != 0 {
			t.setErrno(err)
		} else {
			t.setErrno(errno.EINVAL)
		}
	}
	return int32(r1)
}

func X_set_abort_behavior(t *TLS, _ ...interface{}) uint32 {
	panic(todo(""))
}

func XOpenEventA(t *TLS, dwDesiredAccess uint32, bInheritHandle uint32, lpName uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v dwDesiredAccess=%v bInheritHandle=%v lpName=%v, (%v:)", t, dwDesiredAccess, bInheritHandle, lpName, origin(2))
	}
	r0, _, err := syscall.Syscall(procOpenEventA.Addr(), 3, uintptr(dwDesiredAccess), uintptr(bInheritHandle), lpName)
	if r0 == 0 {
		t.setErrno(err)
	}
	return r0
}

func X_msize(t *TLS, memblock uintptr) types.Size_t {
	if __ccgo_strace {
		trc("t=%v memblock=%v, (%v:)", t, memblock, origin(2))
	}
	return types.Size_t(UsableSize(memblock))
}

func X_byteswap_ulong(t *TLS, val ulong) ulong {
	if __ccgo_strace {
		trc("t=%v val=%v, (%v:)", t, val, origin(2))
	}
	return X__builtin_bswap32(t, val)
}

func X_byteswap_uint64(t *TLS, val uint64) uint64 {
	if __ccgo_strace {
		trc("t=%v val=%v, (%v:)", t, val, origin(2))
	}
	return X__builtin_bswap64(t, val)
}

func X_commit(t *TLS, fd int32) int32 {
	if __ccgo_strace {
		trc("t=%v fd=%v, (%v:)", t, fd, origin(2))
	}
	return Xfsync(t, fd)
}

func X_stati64(t *TLS, path, buffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v buffer=%v, (%v:)", t, path, buffer, origin(2))
	}
	r0, _, err := syscall.SyscallN(procStati64.Addr(), uintptr(path), uintptr(buffer))
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_fstati64(t *TLS, fd int32, buffer uintptr) int32 {
	return X_fstat64(t, fd, buffer)
}

func X_findnext32(t *TLS, handle types.Intptr_t, buffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v handle=%v buffer=%v, (%v:)", t, handle, buffer, origin(2))
	}
	r0, _, err := syscall.SyscallN(procFindnext32.Addr(), uintptr(handle), buffer)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_findfirst32(t *TLS, filespec, fileinfo uintptr) types.Intptr_t {
	if __ccgo_strace {
		trc("t=%v fileinfo=%v, (%v:)", t, fileinfo, origin(2))
	}
	r0, _, err := syscall.SyscallN(procFindfirst32.Addr(), filespec, fileinfo)
	if err != 0 {
		t.setErrno(err)
	}
	return types.Intptr_t(r0)
}

func Xstrtol(t *TLS, nptr, endptr uintptr, base int32) long {
	if __ccgo_strace {
		trc("t=%v endptr=%v base=%v, (%v:)", t, endptr, base, origin(2))
	}

	var s uintptr = nptr
	var acc ulong
	var c byte
	var cutoff ulong
	var neg int32
	var any int32
	var cutlim int32

	for {
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
		var sp = strings.TrimSpace(string(c))
		if len(sp) > 0 {
			break
		}
	}

	if c == '-' {
		neg = 1
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
	} else if c == '+' {
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
	}

	sp := *(*byte)(unsafe.Pointer(s))

	if (base == 0 || base == 16) &&
		c == '0' && (sp == 'x' || sp == 'X') {
		PostIncUintptr(&s, 1)
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
		base = 16
	}
	if base == 0 {
		if c == '0' {
			base = 0
		} else {
			base = 10
		}
	}

	var ULONG_MAX ulong = 0xFFFFFFFF
	var LONG_MAX long = long(ULONG_MAX >> 1)
	var LONG_MIN long = ^LONG_MAX

	if neg == 1 {
		cutoff = ulong(-1 * LONG_MIN)
	} else {
		cutoff = ulong(LONG_MAX)
	}
	cutlim = int32(cutoff % ulong(base))
	cutoff = cutoff / ulong(base)

	acc = 0
	any = 0

	for {
		var cs = string(c)
		if unicode.IsDigit([]rune(cs)[0]) {
			c -= '0'
		} else if unicode.IsLetter([]rune(cs)[0]) {
			if unicode.IsUpper([]rune(cs)[0]) {
				c -= 'A' - 10
			} else {
				c -= 'a' - 10
			}
		} else {
			break
		}

		if int32(c) >= base {
			break
		}
		if any < 0 || acc > cutoff || (acc == cutoff && int32(c) > cutlim) {
			any = -1

		} else {
			any = 1
			acc *= ulong(base)
			acc += ulong(c)
		}

		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
	}

	if any < 0 {
		if neg == 1 {
			acc = ulong(LONG_MIN)
		} else {
			acc = ulong(LONG_MAX)
		}
		t.setErrno(errno.ERANGE)
	} else if neg == 1 {
		acc = -acc
	}

	if endptr != 0 {
		if any == 1 {
			PostDecUintptr(&s, 1)
			AssignPtrUintptr(endptr, s)
		} else {
			AssignPtrUintptr(endptr, nptr)
		}
	}
	return long(acc)
}

func Xstrtoul(t *TLS, nptr, endptr uintptr, base int32) ulong {
	if __ccgo_strace {
		trc("t=%v endptr=%v base=%v, (%v:)", t, endptr, base, origin(2))
	}
	var s uintptr = nptr
	var acc ulong
	var c byte
	var cutoff ulong
	var neg int32
	var any int32
	var cutlim int32

	for {
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
		var sp = strings.TrimSpace(string(c))
		if len(sp) > 0 {
			break
		}
	}

	if c == '-' {
		neg = 1
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
	} else if c == '+' {
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
	}

	sp := *(*byte)(unsafe.Pointer(s))

	if (base == 0 || base == 16) &&
		c == '0' && (sp == 'x' || sp == 'X') {
		PostIncUintptr(&s, 1)
		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
		base = 16
	}
	if base == 0 {
		if c == '0' {
			base = 0
		} else {
			base = 10
		}
	}
	var ULONG_MAX ulong = 0xFFFFFFFF

	cutoff = ULONG_MAX / ulong(base)
	cutlim = int32(ULONG_MAX % ulong(base))

	acc = 0
	any = 0

	for {
		var cs = string(c)
		if unicode.IsDigit([]rune(cs)[0]) {
			c -= '0'
		} else if unicode.IsLetter([]rune(cs)[0]) {
			if unicode.IsUpper([]rune(cs)[0]) {
				c -= 'A' - 10
			} else {
				c -= 'a' - 10
			}
		} else {
			break
		}

		if int32(c) >= base {
			break
		}
		if any < 0 || acc > cutoff || (acc == cutoff && int32(c) > cutlim) {
			any = -1

		} else {
			any = 1
			acc *= ulong(base)
			acc += ulong(c)
		}

		c = *(*byte)(unsafe.Pointer(s))
		PostIncUintptr(&s, 1)
	}

	if any < 0 {
		acc = ULONG_MAX
		t.setErrno(errno.ERANGE)
	} else if neg == 1 {
		acc = -acc
	}

	if endptr != 0 {
		if any == 1 {
			PostDecUintptr(&s, 1)
			AssignPtrUintptr(endptr, s)
		} else {
			AssignPtrUintptr(endptr, nptr)
		}
	}
	return acc
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

func Xrint(tls *TLS, x float64) float64 {
	if __ccgo_strace {
		trc("tls=%v x=%v, (%v:)", tls, x, origin(2))
	}
	switch {
	case x == 0:
		return 0
	case math.IsInf(x, 0), math.IsNaN(x):
		return x
	case x >= math.MinInt64 && x <= math.MaxInt64 && float64(int64(x)) == x:
		return x
	case x >= 0:
		return math.Floor(x + 0.5)
	default:
		return math.Ceil(x - 0.5)
	}
}

func Xfdopen(t *TLS, fd int32, mode uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v fd=%v mode=%v, (%v:)", t, fd, mode, origin(2))
	}
	panic(todo(""))
}

func X_gmtime64(t *TLS, sourceTime uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v sourceTime=%v, (%v:)", t, sourceTime, origin(2))
	}
	r0, _, err := syscall.SyscallN(procGmtime64.Addr(), uintptr(sourceTime))
	if err != 0 {
		t.setErrno(err)
	}
	return uintptr(r0)
}

func X_mktime64(t *TLS, timeptr uintptr) time.X__time64_t {
	if __ccgo_strace {
		trc("t=%v timeptr=%v, (%v:)", t, timeptr, origin(2))
	}
	return time.X__time64_t(Xmktime(t, timeptr))
}

func Xgai_strerrorA(t *TLS, ecode int32) uintptr {
	if __ccgo_strace {
		trc("t=%v ecode=%v, (%v:)", t, ecode, origin(2))
	}
	panic(todo(""))
}

type __timeb64 struct {
	time     types.X__time64_t
	millitm  uint32
	timezone int16
	dstflag  int16
}

func X_ftime64(t *TLS, timeptr uintptr) {
	if __ccgo_strace {
		trc("t=%v timeptr=%v, (%v:)", t, timeptr, origin(2))
	}
	tm := gotime.Now()
	(*__timeb64)(unsafe.Pointer(timeptr)).time = types.X__time64_t(tm.Unix())

	(*__timeb64)(unsafe.Pointer(timeptr)).millitm = uint32(int64(tm.Nanosecond()) / 1e6)
}

func X__ccgo_getMutexType(tls *TLS, m uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v m=%v, (%v:)", tls, m, origin(2))
	}
	return *(*int32)(unsafe.Pointer(m)) & 15
}

func X__ccgo_pthreadAttrGetDetachState(tls *TLS, a uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v, (%v:)", tls, a, origin(2))
	}
	return *(*int32)(unsafe.Pointer(a))
}

func X__ccgo_pthreadMutexattrGettype(tls *TLS, a uintptr) int32 {
	if __ccgo_strace {
		trc("tls=%v a=%v, (%v:)", tls, a, origin(2))
	}
	return *(*int32)(unsafe.Pointer(a)) & int32(3)
}

func Xchmod(t *TLS, pathname uintptr, mode int32) int32 {
	r0, _, err := syscall.SyscallN(procChmod.Addr(), pathname, uintptr(mode))
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func XGetComputerNameExW(t *TLS, nameType int32, lpBuffer, nSize uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v nameType=%v nSize=%v, (%v:)", t, nameType, nSize, origin(2))
	}
	r0, _, err := syscall.Syscall(procGetComputerNameExW.Addr(), 3, uintptr(nameType), lpBuffer, nSize)
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_copysign(t *TLS, x, y float64) float64 {
	if __ccgo_strace {
		trc("t=%v y=%v, (%v:)", t, y, origin(2))
	}
	return Xcopysign(t, x, y)
}

func X_wtoi(t *TLS, str uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v str=%v, (%v:)", t, str, origin(2))
	}
	panic(todo(""))
}

func allocW(t *TLS, v string) (r uintptr) {
	s := utf16.Encode([]rune(v))
	p := Xcalloc(t, types.Size_t(len(s)+1), 2)
	if p == 0 {
		panic(todo(""))
	}

	r = p
	for _, v := range s {
		*(*uint16)(unsafe.Pointer(p)) = v
		p += 2
	}
	return r
}

func X_wgetenv(t *TLS, varname uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v varname=%v, (%v:)", t, varname, origin(2))
	}
	if !wenvValid {
		bootWinEnviron(t)
	}
	k := strings.ToLower(goWideStringNZ(varname))
	for _, v := range winEnviron[:len(winEnviron)-1] {
		s := strings.ToLower(goWideStringNZ(v))
		x := strings.IndexByte(s, '=')
		if s[:x] == k {
			return v
		}
	}

	return 0
}

func X_wputenv(t *TLS, envstring uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v envstring=%v, (%v:)", t, envstring, origin(2))
	}
	if !wenvValid {
		bootWinEnviron(t)
	}
	s0 := goWideStringNZ(envstring)
	s := strings.ToLower(s0)
	x := strings.IndexByte(s, '=')
	k := s[:x]
	for i, v := range winEnviron[:len(winEnviron)-1] {
		s2 := strings.ToLower(goWideStringNZ(v))
		x := strings.IndexByte(s2, '=')
		if s2[:x] == k {
			Xfree(t, v)
			winEnviron[i] = allocW(t, s0)
			return 0
		}
	}

	np := allocW(t, s0)
	winEnviron = winEnviron[:len(winEnviron)-1]
	winEnviron = append(winEnviron, np, 0)
	wenviron = uintptr(unsafe.Pointer(&winEnviron[0]))
	return 0
}

func bootWinEnviron(t *TLS) {
	winEnviron = winEnviron[:0]
	p := Environ()
	for {
		q := *(*uintptr)(unsafe.Pointer(p))
		p += unsafe.Sizeof(uintptr(0))
		if q == 0 {
			break
		}

		s := GoString(q)

		r := allocW(t, s)
		winEnviron = append(winEnviron, r)
	}
	wenviron = uintptr(unsafe.Pointer(&winEnviron[0]))
	wenvValid = true
}

func Xfabsl(t *TLS, x float64) float64 {
	if __ccgo_strace {
		trc("t=%v x=%v, (%v:)", t, x, origin(2))
	}
	return math.Abs(x)
}

func X__stdio_common_vfprintf(t *TLS, args ...interface{}) int32     { panic("TODO") }
func X__stdio_common_vfprintf_p(t *TLS, args ...interface{}) int32   { panic("TODO") }
func X__stdio_common_vfprintf_s(t *TLS, args ...interface{}) int32   { panic("TODO") }
func X__stdio_common_vfscanf(t *TLS, args ...interface{}) int32      { panic("TODO") }
func X__stdio_common_vfwprintf_s(t *TLS, args ...interface{}) int32  { panic("TODO") }
func X__stdio_common_vfwscanf(t *TLS, args ...interface{}) int32     { panic("TODO") }
func X__stdio_common_vsnprintf_s(t *TLS, args ...interface{}) int32  { panic("TODO") }
func X__stdio_common_vsnwprintf_s(t *TLS, args ...interface{}) int32 { panic("TODO") }
func X__stdio_common_vsprintf(t *TLS, args ...interface{}) int32     { panic("TODO") }
func X__stdio_common_vsprintf_p(t *TLS, args ...interface{}) int32   { panic("TODO") }
func X__stdio_common_vsprintf_s(t *TLS, args ...interface{}) int32   { panic("TODO") }
func X__stdio_common_vsscanf(t *TLS, args ...interface{}) int32      { panic("TODO") }
func X__stdio_common_vswprintf(t *TLS, args ...interface{}) int32    { panic("TODO") }
func X__stdio_common_vswprintf_s(t *TLS, args ...interface{}) int32  { panic("TODO") }
func X__stdio_common_vswscanf(t *TLS, args ...interface{}) int32     { panic("TODO") }

func X_lseeki64(t *TLS, fd int32, offset int64, whence int32) int64 {
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

func Xislower(tls *TLS, c int32) int32 {
	if __ccgo_strace {
		trc("tls=%v c=%v, (%v:)", tls, c, origin(2))
	}
	return Bool32(uint32(c)-uint32('a') < uint32(26))
}

func Xisupper(tls *TLS, c int32) int32 {
	if __ccgo_strace {
		trc("tls=%v c=%v, (%v:)", tls, c, origin(2))
	}
	return Bool32(uint32(c)-uint32('A') < uint32(26))
}

func Xaccess(t *TLS, pathname uintptr, mode int32) int32 {
	if __ccgo_strace {
		trc("t=%v pathname=%v mode=%v, (%v:)", t, pathname, mode, origin(2))
	}
	r0, _, err := syscall.SyscallN(procAccess.Addr(), uintptr(pathname), uintptr(mode))
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func X_vscprintf(t *TLS, format uintptr, argptr uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v format=%v argptr=%v, (%v:)", t, format, argptr, origin(2))
	}

	return int32(len(printf(format, argptr)))
}

func X_stat64i32(t *TLS, path uintptr, buffer uintptr) int32 {
	if __ccgo_strace {
		trc("t=%v path=%v buffer=%v, (%v:)", t, path, buffer, origin(2))
	}
	r0, _, err := syscall.SyscallN(procStat64i32.Addr(), uintptr(path), uintptr(buffer))
	if err != 0 {
		t.setErrno(err)
	}
	return int32(r0)
}

func AtomicLoadNUint8(ptr uintptr, memorder int32) uint8 {
	return byte(a_load_8(ptr))
}

func Xgmtime(t *TLS, sourceTime uintptr) uintptr {
	if __ccgo_strace {
		trc("t=%v sourceTime=%v, (%v:)", t, sourceTime, origin(2))
	}
	r0, _, err := syscall.SyscallN(procGmtime.Addr(), uintptr(sourceTime))
	if err != 0 {
		t.setErrno(err)
	}
	return uintptr(r0)
}

func Xstrftime(tls *TLS, s uintptr, n size_t, f uintptr, tm uintptr) (r size_t) {
	if __ccgo_strace {
		trc("tls=%v s=%v n=%v f=%v tm=%v, (%v:)", tls, s, n, f, tm, origin(2))
		defer func() { trc("-> %v", r) }()
	}
	tt := gotime.Date(
		int((*time.Tm)(unsafe.Pointer(tm)).Ftm_year+1900),
		gotime.Month((*time.Tm)(unsafe.Pointer(tm)).Ftm_mon+1),
		int((*time.Tm)(unsafe.Pointer(tm)).Ftm_mday),
		int((*time.Tm)(unsafe.Pointer(tm)).Ftm_hour),
		int((*time.Tm)(unsafe.Pointer(tm)).Ftm_min),
		int((*time.Tm)(unsafe.Pointer(tm)).Ftm_sec),
		0,
		gotime.UTC,
	)
	fmt := GoString(f)
	var result string
	if fmt != "" {
		result = strftime.Format(fmt, tt)
	}
	switch r = size_t(len(result)); {
	case r > n:
		r = 0
	default:
		copy((*RawMem)(unsafe.Pointer(s))[:r:r], result)
		*(*byte)(unsafe.Pointer(s + uintptr(r))) = 0
	}
	return r

}

func X__mingw_strtod(t *TLS, s uintptr, p uintptr) float64 {
	return Xstrtod(t, s, p)
}

func Xstrtod(t *TLS, s uintptr, p uintptr) float64 {
	if __ccgo_strace {
		trc("tls=%v s=%v p=%v, (%v:)", t, s, p, origin(2))
	}
	r0, _, err := syscall.SyscallN(procStrtod.Addr(), uintptr(s), uintptr(p))
	if err != 0 {
		t.setErrno(err)
	}
	return math.Float64frombits(uint64(r0))
}
