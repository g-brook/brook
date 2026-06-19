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
)

type VisitorConfig struct {
	Id        int    `db:"id"`
	ProxyId   int64  `db:"proxy_id"`
	Token     string `db:"token"`
	LocalPort int    `db:"local_port"`
}

var visitorQuerySQL = "id,proxy_id,token,local_port"

// GetVisitorConfig Get Visitor config By proxyId.
func GetVisitorConfig(proxyId int) (*VisitorConfig, error) {
	query := fmt.Sprintf("select %s from visitor_proxy_cconfig where proxy_id = ?", visitorQuerySQL)
	res, err := Query(query, proxyId)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.rows.Next() {
		return scanVisitor(res.rows)
	}
	return nil, nil
}

func AddVisitorConfig(v *VisitorConfig) error {
	return Exec(
		"insert into visitor_proxy_cconfig(proxy_id,token,local_port) values(?,?,?)",
		v.ProxyId,
		v.Token,
		v.LocalPort,
	)
}

func UpdateVisitorConfig(v *VisitorConfig) error {
	return Exec(
		"update visitor_proxy_cconfig set token=?,local_port=? where proxy_id=?",
		v.Token,
		v.LocalPort,
		v.ProxyId,
	)
}

func SaveVisitorConfig(v *VisitorConfig) error {
	old, err := GetVisitorConfig(int(v.ProxyId))
	if err != nil {
		return err
	}
	if old == nil {
		return AddVisitorConfig(v)
	}
	return UpdateVisitorConfig(v)
}

func DelVisitorConfig(proxyId int) error {
	return Exec("delete from visitor_proxy_cconfig where proxy_id = ?", proxyId)
}

func scanVisitor(rows *sql.Rows) (*VisitorConfig, error) {
	var u VisitorConfig
	err := rows.Scan(
		&u.Id,
		&u.ProxyId,
		&u.Token,
		&u.LocalPort,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
