// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sql_store_upgrade

import (
	"database/sql"
	"embed"
	"fmt"

	"maunium.net/go/mautrix/util/dbutil"
)

var Table dbutil.UpgradeTable

const VersionTableName = "crypto_version"

//go:embed *.sql
var fs embed.FS

func init() {
	Table.Register(-1, 3, "Unsupported version", func(tx dbutil.Transaction, database *dbutil.Database) error {
		return fmt.Errorf("upgrading from versions 1 and 2 of the crypto store is no longer supported in mautrix-go v0.12+")
	})
	Table.RegisterFS(fs)
}

// Upgrade upgrades the database from the current to the latest version available.
func Upgrade(sqlDB *sql.DB, dialect string) error {
	db, err := dbutil.NewWithDB(sqlDB, dialect)
	if err != nil {
		return err
	}
	db.VersionTable = VersionTableName
	db.UpgradeTable = Table
	return db.Upgrade()
}
