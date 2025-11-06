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

type ProxyConfig struct {
	Idx        int    `db:"idx" json:"id"`
	Name       string `db:"name" json:"name"`
	Tag        string `db:"tag" json:"tag"`
	RemotePort int    `db:"remote_port" json:"remotePort"`
	ProxyID    string `db:"proxy_id" json:"proxyId"`
	Protocol   string `db:"protocol" json:"protocol"`
	State      int    `db:"state" json:"state"`
	RunState   int    `db:"run_state"`
	IsRunning  bool   `json:"isRunning"`
	Runtime    string `json:"runtime"`
	IsExistWeb bool   `json:"isExistWeb"`
	Clients    int    `json:"clients"`
}

func (r *ProxyConfig) IsHttpOrHttps() bool {
	return r.Protocol == "HTTP" || r.Protocol == "HTTPS"
}

func AddProxyConfig(p ProxyConfig) error {
	err := Exec(`
            INSERT INTO proxy_config(name, tag, remote_port, proxy_id, protocol,state,run_state)
            VALUES (?, ?, ?, ?, ?,?,?);
        `, p.Name, p.Tag, p.RemotePort, p.ProxyID, p.Protocol, p.State, p.RunState)
	return err
}

func DelProxyConfig(id int) error {
	err := Exec("DELETE FROM proxy_config WHERE idx = ?", id)
	return err
}

func UpdateProxyConfig(p ProxyConfig) error {
	err := Exec("update proxy_config set name=?,tag=?,proxy_id=?,protocol=? where idx=?", p.Name, p.Tag, p.ProxyID, p.Protocol, p.Idx)
	return err
}

func UpdateProxyState(p ProxyConfig) error {
	err := Exec("update proxy_config set state=? where idx=?", p.State, p.Idx)
	return err
}

func GetAllProxyConfig() []*ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config where state = 1")
	if err != nil {
		return nil
	}
	defer res.Close()

	var list []*ProxyConfig
	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err != nil {
			return nil
		}
		list = append(list, &p)
	}
	return list
}

func GetProxyConfigByProxyId(proxyId string) *ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config where state = 1 and proxy_id = ?", proxyId)
	if err != nil {
		return nil
	}
	defer res.Close()

	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err == nil {
			return &p
		}
	}
	return nil
}

func GetProxyConfigById(id int) *ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config where state = 1 and idx = ?", id)
	if err != nil {
		return nil
	}
	defer res.Close()

	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err == nil {
			return &p
		}
	}
	return nil
}

func QueryProxyConfig() []*ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config")
	if err != nil {
		return nil
	}
	defer res.Close()

	var list []*ProxyConfig
	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err != nil {
			return nil
		}
		list = append(list, &p)
	}
	return list
}
