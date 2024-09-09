// Copyright 2019 The Sqlite Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"modernc.org/sqlite/internal/libc"
)

func X__ccgo_sqlite3_log(t *libc.TLS, iErrCode int32, zFormat uintptr, va uintptr) {
	libc.X__ccgo_sqlite3_log(t, iErrCode, zFormat, va)
}
