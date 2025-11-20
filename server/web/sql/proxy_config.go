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

type ProxyConfig struct {
	Idx         int            `db:"idx"`
	Name        string         `db:"name"`
	Tag         string         `db:"tag"`
	RemotePort  int            `db:"remote_port"`
	ProxyID     string         `db:"proxy_id"`
	Protocol    string         `db:"protocol"`
	State       int            `db:"state"`
	Destination sql.NullString `db:"destination"`
	RunState    int            `db:"run_state"`
}

var (
	ProxyQuerySQL = "idx,name, tag, remote_port, proxy_id, protocol,state,run_state,destination"
)

func AddProxyConfig(p ProxyConfig) error {
	err := Exec(`
            INSERT INTO proxy_config(name, tag, remote_port, proxy_id, protocol,state,run_state, destination)
            VALUES (?, ?, ?, ?, ?,?,?,?);
        `, p.Name, p.Tag, p.RemotePort, p.ProxyID, p.Protocol, p.State, p.RunState, p.Destination)
	return err
}

func DelProxyConfig(id int) error {
	err := Exec("DELETE FROM proxy_config WHERE idx = ?", id)
	return err
}

func UpdateProxyConfig(p ProxyConfig) error {
	err := Exec("update proxy_config set name=?,tag=?,proxy_id=?,protocol=?,destination=? where idx=?", p.Name, p.Tag, p.ProxyID, p.Protocol, p.Destination.String, p.Idx)
	return err
}

func UpdateProxyState(p ProxyConfig) error {
	err := Exec("update proxy_config set state=? where idx=?", p.State, p.Idx)
	return err
}

func GetAllProxyConfig() []*ProxyConfig {
	selectSQL := fmt.Sprintf("select %s from proxy_config where state = 1", ProxyQuerySQL)
	res, err := Query(selectSQL)
	if err != nil {
		return nil
	}
	defer res.Close()

	var list []*ProxyConfig
	for res.rows.Next() {
		if p, err := scanProxyConfig(res.rows); err != nil {
			return nil
		} else {
			list = append(list, p)
		}
	}
	return list
}

func GetProxyConfigByProxyId(proxyId string) *ProxyConfig {
	selectSQL := fmt.Sprintf("select %s from proxy_config where  state = 1 and proxy_id = ?", ProxyQuerySQL)
	res, err := Query(selectSQL, proxyId)
	if err != nil {
		return nil
	}
	defer res.Close()

	for res.rows.Next() {
		if p, err := scanProxyConfig(res.rows); err != nil {
			return nil
		} else {
			return p
		}
	}
	return nil
}

func GetProxyConfigById(id int) *ProxyConfig {
	return GetProxyConfigByIdAndState(id, 1, true)
}

func GetProxyConfigByIdNotState(id int) *ProxyConfig {
	return GetProxyConfigByIdAndState(id, 1, false)
}

func GetProxyConfigByIdAndState(id int, state int, isCheckState bool) *ProxyConfig {
	var res *Result
	var err error
	selectSQL := fmt.Sprintf("select %s from proxy_config", ProxyQuerySQL)
	if isCheckState {
		selectSQL += " where state = ? and idx = ? "
		res, err = Query(selectSQL, state, id)
	} else {
		selectSQL += " where idx = ? "
		res, err = Query(selectSQL, id)
	}
	if err != nil {
		return nil
	}
	defer res.Close()

	for res.rows.Next() {
		if p, err := scanProxyConfig(res.rows); err != nil {
			return nil
		} else {
			return p
		}
	}
	return nil
}

func QueryProxyConfig() []*ProxyConfig {
	selectSQL := fmt.Sprintf("select %s from proxy_config", ProxyQuerySQL)
	res, err := Query(selectSQL)
	if err != nil {
		return nil
	}
	defer res.Close()

	var list []*ProxyConfig
	for res.rows.Next() {
		if p, err := scanProxyConfig(res.rows); err != nil {
			return nil
		} else {
			list = append(list, p)
		}
	}
	return list
}

func scanProxyConfig(rows *sql.Rows) (*ProxyConfig, error) {
	var p ProxyConfig
	err := rows.Scan(
		&p.Idx,
		&p.Name,
		&p.Tag,
		&p.RemotePort,
		&p.ProxyID,
		&p.Protocol,
		&p.State,
		&p.RunState,
		&p.Destination,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
