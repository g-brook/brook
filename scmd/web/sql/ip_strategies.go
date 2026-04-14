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
	"fmt"
	"time"
)

type IpStrategy struct {
	Id          int16     `db:"id"`
	Name        string    `db:"name"`
	Type        string    `db:"type"`
	BindHandler string    `db:"bind_handler"`
	Status      int16     `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

var sqltext = "id,name,type,bind_handler,status,created_at,updated_at"

func SelectByBindHandler(handler string) (*IpStrategy, error) {
	selectSQL := fmt.Sprintf("select %s from ip_strategies where bind_handler = ? and status = 1", sqltext)
	res, err := Query(selectSQL, handler)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.rows.Next() {
		if p, err := scanIpStrategies(res.rows); err != nil {
			return nil, err
		} else {
			return p, nil
		}
	}
	return nil, nil
}

func scanIpStrategies(rows *sql.Rows) (*IpStrategy, error) {
	var p IpStrategy
	err := rows.Scan(
		&p.Id,
		&p.Name,
		&p.Type,
		&p.BindHandler,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
