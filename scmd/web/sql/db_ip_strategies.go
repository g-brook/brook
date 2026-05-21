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
	Id        int16     `db:"id" maps:"id"`
	Name      string    `db:"name" maps:"name"`
	Type      string    `db:"type" maps:"type"`
	Status    int16     `db:"status" maps:"status"`
	CreatedAt time.Time `db:"created_at" maps:"created_at"`
	UpdatedAt time.Time `db:"updated_at" maps:"updated_at"`
}

var sqltext = "id,name,type,status,created_at,updated_at"

func AddIpStrategy(s *IpStrategy) error {
	return Exec(
		`INSERT INTO ip_strategies(name, type, status, created_at, updated_at)
         VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		s.Name,
		s.Type,
		s.Status,
	)
}

func UpdateIpStrategy(s *IpStrategy) error {
	return Exec(
		`UPDATE ip_strategies
         SET name = ?, type = ?,  status = ?, updated_at = CURRENT_TIMESTAMP
         WHERE id = ?`,
		s.Name,
		s.Type,
		s.Status,
		s.Id,
	)
}

func DelIpStrategy(id int16) error {
	return Exec("DELETE FROM ip_strategies WHERE id = ?", id)
}

func SelectIpStrategyById(id int16) (*IpStrategy, error) {
	selectSQL := fmt.Sprintf("select %s from ip_strategies where id = ?", sqltext)
	res, err := Query(selectSQL, id)
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

func SelectIpStrategyAll() ([]*IpStrategy, error) {
	selectSQL := fmt.Sprintf("select %s from ip_strategies", sqltext)
	res, err := Query(selectSQL)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var list []*IpStrategy
	for res.rows.Next() {
		if p, err := scanIpStrategies(res.rows); err != nil {
			return nil, err
		} else {
			list = append(list, p)
		}
	}
	return list, nil
}

func scanIpStrategies(rows *sql.Rows) (*IpStrategy, error) {
	var p IpStrategy
	err := rows.Scan(
		&p.Id,
		&p.Name,
		&p.Type,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
