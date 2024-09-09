// Copyright 2021 The Sqlite Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"modernc.org/sqlite/internal/libc"
	"modernc.org/sqlite/internal/libc/sys/types"
)

func init() {
	tls := libc.NewTLS()
	if Xsqlite3_threadsafe(tls) == 0 {
		panic(fmt.Errorf("sqlite: thread safety configuration error"))
	}

	varArgs := libc.Xmalloc(tls, types.Size_t(unsafe.Sizeof(uintptr(0))))
	if varArgs == 0 {
		panic(fmt.Errorf("cannot allocate memory"))
	}

	if rc := Xsqlite3_config(tls, SQLITE_CONFIG_MUTEX, libc.VaList(varArgs, uintptr(unsafe.Pointer(&mutexMethods)))); rc != SQLITE_OK {
		p := Xsqlite3_errstr(tls, rc)
		str := libc.GoString(p)
		panic(fmt.Errorf("sqlite: failed to configure mutex methods: %v", str))
	}

	libc.Xfree(tls, varArgs)
	tls.Close()
}

var (
	mutexMethods = Sqlite3_mutex_methods{
		FxMutexInit: *(*uintptr)(unsafe.Pointer(&struct{ f func(*libc.TLS) int32 }{mutexInit})),
		FxMutexEnd:  *(*uintptr)(unsafe.Pointer(&struct{ f func(*libc.TLS) int32 }{mutexEnd})),
		FxMutexAlloc: *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, int32) uintptr
		}{mutexAlloc})),
		FxMutexFree:  *(*uintptr)(unsafe.Pointer(&struct{ f func(*libc.TLS, uintptr) }{mutexFree})),
		FxMutexEnter: *(*uintptr)(unsafe.Pointer(&struct{ f func(*libc.TLS, uintptr) }{mutexEnter})),
		FxMutexTry: *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, uintptr) int32
		}{mutexTry})),
		FxMutexLeave: *(*uintptr)(unsafe.Pointer(&struct{ f func(*libc.TLS, uintptr) }{mutexLeave})),
		FxMutexHeld: *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, uintptr) int32
		}{mutexHeld})),
		FxMutexNotheld: *(*uintptr)(unsafe.Pointer(&struct {
			f func(*libc.TLS, uintptr) int32
		}{mutexNotheld})),
	}

	MutexCounters = libc.NewPerfCounter([]string{
		"enter-fast",
		"enter-recursive",
		"enter-recursive-loop",
		"try-fast",
		"try-recursive",
	})
	MutexEnterCallers = libc.NewStackCapture(4)

	mutexes mutexPool

	mutexApp1   = mutexes.alloc(false)
	mutexApp2   = mutexes.alloc(false)
	mutexApp3   = mutexes.alloc(false)
	mutexLRU    = mutexes.alloc(false)
	mutexMaster = mutexes.alloc(false)
	mutexMem    = mutexes.alloc(false)
	mutexOpen   = mutexes.alloc(false)
	mutexPMem   = mutexes.alloc(false)
	mutexPRNG   = mutexes.alloc(false)
	mutexVFS1   = mutexes.alloc(false)
	mutexVFS2   = mutexes.alloc(false)
	mutexVFS3   = mutexes.alloc(false)
)

type mutexPool struct {
	sync.Mutex
	a        []*[256]mutex
	freeList []int
}

func mutexFromPtr(p uintptr) *mutex {
	if p == 0 {
		return nil
	}

	ix := p - 1

	mutexes.Lock()
	defer mutexes.Unlock()

	return &mutexes.a[ix>>8][ix&255]
}

func (m *mutexPool) alloc(recursive bool) uintptr {
	m.Lock()
	defer m.Unlock()

	n := len(m.freeList)
	if n == 0 {
		outer := len(m.a) << 8
		m.a = append(m.a, &[256]mutex{})
		for i := 0; i < 256; i++ {
			m.freeList = append(m.freeList, outer+i)
		}
		n = len(m.freeList)
	}
	ix := m.freeList[n-1]
	outer := ix >> 8
	inner := ix & 255
	m.freeList = m.freeList[:n-1]
	p := &m.a[outer][inner]
	p.poolIndex = ix
	p.recursive = recursive
	return uintptr(ix) + 1
}

func (m *mutexPool) free(p uintptr) {
	ptr := mutexFromPtr(p)
	ix := ptr.poolIndex
	*ptr = mutex{}

	m.Lock()
	defer m.Unlock()

	m.freeList = append(m.freeList, ix)
}

type mutex struct {
	sync.Mutex
	wait sync.Mutex

	poolIndex int

	cnt int32
	id  int32

	recursive bool
}

func (m *mutex) enter(id int32) {
	if !m.recursive {
		m.Lock()
		m.id = id
		return
	}

	for {
		m.Lock()
		switch m.id {
		case 0:
			m.cnt = 1
			m.id = id
			m.wait.Lock()
			m.Unlock()
			return
		case id:
			m.cnt++
			m.Unlock()
			return
		}

		m.Unlock()
		m.wait.Lock()

		m.wait.Unlock()
	}
}

func (m *mutex) try(id int32) int32 {
	if !m.recursive {
		return SQLITE_BUSY
	}

	m.Lock()
	switch m.id {
	case 0:
		m.cnt = 1
		m.id = id
		m.wait.Lock()
		m.Unlock()
		return SQLITE_OK
	case id:
		m.cnt++
		m.Unlock()
		return SQLITE_OK
	}

	m.Unlock()
	return SQLITE_BUSY
}

func (m *mutex) leave(id int32) {
	if !m.recursive {
		m.id = 0
		m.Unlock()
		return
	}

	m.Lock()
	m.cnt--
	if m.cnt == 0 {
		m.id = 0
		m.wait.Unlock()
	}
	m.Unlock()
}

func mutexInit(tls *libc.TLS) int32 { return SQLITE_OK }

func mutexEnd(tls *libc.TLS) int32 { return SQLITE_OK }

func mutexAlloc(tls *libc.TLS, typ int32) uintptr {
	defer func() {
	}()
	switch typ {
	case SQLITE_MUTEX_FAST:
		return mutexes.alloc(false)
	case SQLITE_MUTEX_RECURSIVE:
		return mutexes.alloc(true)
	case SQLITE_MUTEX_STATIC_MASTER:
		return mutexMaster
	case SQLITE_MUTEX_STATIC_MEM:
		return mutexMem
	case SQLITE_MUTEX_STATIC_OPEN:
		return mutexOpen
	case SQLITE_MUTEX_STATIC_PRNG:
		return mutexPRNG
	case SQLITE_MUTEX_STATIC_LRU:
		return mutexLRU
	case SQLITE_MUTEX_STATIC_PMEM:
		return mutexPMem
	case SQLITE_MUTEX_STATIC_APP1:
		return mutexApp1
	case SQLITE_MUTEX_STATIC_APP2:
		return mutexApp2
	case SQLITE_MUTEX_STATIC_APP3:
		return mutexApp3
	case SQLITE_MUTEX_STATIC_VFS1:
		return mutexVFS1
	case SQLITE_MUTEX_STATIC_VFS2:
		return mutexVFS2
	case SQLITE_MUTEX_STATIC_VFS3:
		return mutexVFS3
	default:
		return 0
	}
}

func mutexFree(tls *libc.TLS, m uintptr) { mutexes.free(m) }

func mutexEnter(tls *libc.TLS, m uintptr) {
	if m == 0 {
		return
	}
	mutexFromPtr(m).enter(tls.ID)
}

func mutexTry(tls *libc.TLS, m uintptr) int32 {
	if m == 0 {
		return SQLITE_OK
	}

	return mutexFromPtr(m).try(tls.ID)
}

func mutexLeave(tls *libc.TLS, m uintptr) {
	if m == 0 {
		return
	}

	mutexFromPtr(m).leave(tls.ID)
}

func mutexHeld(tls *libc.TLS, m uintptr) int32 {
	if m == 0 {
		return 1
	}

	return libc.Bool32(atomic.LoadInt32(&mutexFromPtr(m).id) == tls.ID)
}

func mutexNotheld(tls *libc.TLS, m uintptr) int32 {
	if m == 0 {
		return 1
	}

	return libc.Bool32(atomic.LoadInt32(&mutexFromPtr(m).id) != tls.ID)
}
