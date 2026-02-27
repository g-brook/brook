/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sql

import (
	"embed"
	"io/fs"
	"strconv"

	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/version"
)

type Info struct {
	Id    int    `db:"id"`
	Key   string `db:"key"`
	Value string `db:"value"`
}

const (
	DBVersionKey = "db_version"
	sqlFileDir   = "sql_files"
)

//go:embed sql_files/*
var sqlFiles embed.FS

var staticFs, _ = fs.Sub(sqlFiles, sqlFileDir)

func CheckInfoDB() error {
	// 检查并创建表
	if err := ensureInfoTableExists(); err != nil {
		return err
	}
	// 检查并初始化版本信息
	return ensureVersionInfo()
}

// ensureInfoTableExists 确保 info 表存在
func ensureInfoTableExists() error {
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='info'`
	result, err := Query(query)
	if err != nil {
		return err
	}
	defer result.Close()
	var count int
	if result.rows.Next() {
		if err := result.rows.Scan(&count); err != nil {
			return err
		}
	}

	if count == 0 {
		result.Close()
		createTableSQL := `
		CREATE TABLE info (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT
		)`
		if err := Exec(createTableSQL); err != nil {
			return err
		}
	}

	return nil
}

// ensureVersionInfo 确保版本信息存在
func ensureVersionInfo() error {
	query := `SELECT key, value FROM info WHERE key = ?`
	result, err := Query(query, DBVersionKey)
	if err != nil {
		return err
	}
	defer result.Close()
	if result.rows.Next() {
		return nil
	}
	result.Close()
	// 版本信息不存在，插入默认值
	insertSQL := `INSERT INTO info (key, value) VALUES (?, ?)`
	return Exec(insertSQL, DBVersionKey, version.GetDbVersion())
}

// getCurrentDBVersion 获取当前数据库版本号
func getCurrentDBVersion() (int, error) {
	query := `SELECT value FROM info WHERE key = ?`
	result, err := Query(query, DBVersionKey)
	if err != nil {
		return 0, err
	}
	defer result.Close()

	var currentVersion int
	if result.rows.Next() {
		var versionStr string
		if err := result.rows.Scan(&versionStr); err != nil {
			return 0, err
		}
		currentVersion, err = strconv.Atoi(versionStr)
		if err != nil {
			return 0, err
		}
	}
	return currentVersion, nil
}

func CheckDBVersion() (bool, error) {
	currentVersion, err := getCurrentDBVersion()
	if err != nil {
		return false, err
	}

	return currentVersion < version.GetDbVersion(), nil
}

func UpdateTableStruct() error {
	currentVersion, err := getCurrentDBVersion()
	if err != nil {
		return err
	}
	targetVersion := version.GetDbVersion()
	if currentVersion >= targetVersion {
		return nil
	}

	readFile := func(sqlFile string) error {
		sqlContent, err := fs.ReadFile(staticFs, sqlFile)
		if err != nil {
			log.Warn("error reading sql file %v:%v", sqlFile, err)
			return err
		}
		sqlText := string(sqlContent)
		log.Info("executing sql file %v", sqlText)
		if err := Exec(sqlText); err != nil {
			log.Warn("error executing sql %v:%v,%v", sqlFile, sqlText, err)
			return err
		}
		return nil
	}
	log.Info("current db version %v, target db version %v", currentVersion, targetVersion)
	for v := currentVersion + 1; v <= targetVersion; v++ {
		sqlFile := strconv.Itoa(v) + ".sql"
		if err := readFile(sqlFile); err != nil {
			log.Warn("error reading sql file %v:%v", sqlFile, err)
			return err
		}
	}
	updateSQL := `UPDATE info SET value = ? WHERE key = ?`
	if err := Exec(updateSQL, strconv.Itoa(targetVersion), DBVersionKey); err != nil {
		return err
	}
	return nil
}
