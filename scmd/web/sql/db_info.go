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
	"regexp"
	"strconv"
	"strings"

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

func GetInfoValue(key string) (string, error) {
	query := `SELECT value FROM info WHERE key = ?`
	result, err := Query(query, key)
	if err != nil {
		return "", err
	}
	defer result.Close()
	if result.rows.Next() {
		var value string
		if err := result.rows.Scan(&value); err != nil {
			return "", err
		}
		return value, nil
	}
	return "", nil
}

func AddInfoValue(key string, value string) error {
	insertSQL := `INSERT INTO info (key, value) VALUES (?, ?)`
	return Exec(insertSQL, key, value)
}

func UpdateInfoValue(key string, value string) error {
	updateSQL := `UPDATE info SET value = ? WHERE key = ?`
	return Exec(updateSQL, key, value)
}

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
	dbVersion, err := getCurrentDBVersion()
	if err == nil && dbVersion > 0 {
		return nil
	}
	return AddInfoValue(DBVersionKey, strconv.Itoa(version.GetDbVersion()))
}

// getCurrentDBVersion 获取当前数据库版本号
func getCurrentDBVersion() (int, error) {
	value, err := GetInfoValue(DBVersionKey)
	if err != nil || value == "" {
		return 0, err
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return v, nil
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
		newCsql := removeSQLComments(sqlText)
		sqlList := strings.Split(newCsql, ";")
		for _, sql := range sqlList {
			sql = strings.TrimSpace(sql)
			if sql == "" {
				continue
			}
			log.Debug("current execute sql: %v", sql)
			err = Exec(sql)
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "duplicate column name") {
					log.Info("duplicate column skip  %s", sql)
					continue
				}
				log.Warn("sql execute error: %s, err: %v", sql, err)
				return err
			}
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
	err = UpdateInfoValue(DBVersionKey, strconv.Itoa(version.GetDbVersion()))
	if err != nil {
		return err
	}
	return nil
}

func clearAllTables() error {
	rows, err := Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return err
	}
	defer rows.Close()
	r := rows.rows
	var tables []string
	for r.Next() {
		var table string
		_ = r.Scan(&table)
		tables = append(tables, table)
	}
	if len(tables) != 0 {
		rows.Close()
		for _, table := range tables {
			dropSQL := "DROP TABLE IF EXISTS " + table
			if err := Exec(dropSQL); err != nil {
				return err
			}
		}
	}
	return Exec("VACUUM")
}
func InitDBStruct() error {
	err := clearAllTables()
	if err != nil {
		log.Warn("clear All table:init db err: %v", err)
		return err
	}
	readFile := func(sqlFile string) error {
		sqlContent, err := fs.ReadFile(staticFs, sqlFile)
		if err != nil {
			log.Warn("error reading sql file %v:%v", sqlFile, err)
			return err
		}
		sqlText := string(sqlContent)
		newCsql := removeSQLComments(sqlText)
		sqlList := strings.Split(newCsql, ";")
		for _, sql := range sqlList {
			sql = strings.TrimSpace(sql)
			log.Debug("current execute sql: %v", sql)
			err = Exec(sql)
			if err != nil {
				log.Warn("initDB:sql execute error: %s, err: %v", sql, err)
				return err
			}
		}
		return nil

	}
	err = readFile("full.sql")
	if err != nil {
		return nil
	}
	return insertConfig()
}

func insertConfig() error {
	return ensureVersionInfo()
}
func removeSQLComments(sql string) string {
	// 匹配 /* 开头 */ 结尾的所有注释
	re := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	return re.ReplaceAllString(sql, "")
}
