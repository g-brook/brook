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
	"database/sql"

	"github.com/brook/common/log"
	_ "modernc.org/sqlite"
)

var SqlDB *sql.DB

type Result struct {
	rows *sql.Rows
}

func (t *Result) Close() {
	err := t.rows.Close()
	if err != nil {
		log.Error("err: %v", err)
	}
}

// InitSQLDB initializes the SQLite database connection with specific settings
// It returns an error if the connection fails or if there are issues configuring the connection parameters
func InitSQLDB() error {
	// Open a connection to the SQLite database file named "db.db" in the current directory
	db, err := sql.Open("sqlite", "./db.db")
	if err != nil {
		// Return the error if the database connection cannot be established
		log.Error("err: %v", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		log.Error("db.Ping err: %v", err)
		return err
	}
	// Set the maximum number of idle connections in the connection pool to 1
	db.SetMaxIdleConns(1)
	// Set the maximum number of open connections to 1
	db.SetMaxOpenConns(1)
	// Set the maximum lifetime of a connection to 0 (connections can be reused indefinitely)
	db.SetConnMaxLifetime(0)
	// Assign the database connection to the global SqlDB variable
	SqlDB = db
	// Return nil to indicate successful initialization
	return nil
}

func Query(sql string, args ...any) (*Result, error) {
	rows, err := SqlDB.Query(sql, args...)
	if err != nil {
		log.Error("sql: %s, err: %v", sql, err)
		return nil, err
	}
	return &Result{
		rows: rows,
	}, nil
}

func Exec(sql string, args ...any) error {
	_, err := SqlDB.Exec(sql, args...)
	if err != nil {
		log.Error("sql: %s, err: %v", sql, err)
	}
	return err
}
