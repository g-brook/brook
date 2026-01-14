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
)

type WebProxyConfig struct {
	Id         string        `db:"id" json:"id"`
	RefProxyId int           `db:"ref_proxy_id" json:"refProxyId"`
	Proxy      string        `db:"proxy" json:"proxy"`
	CertId     sql.NullInt32 `db:"cert_id" json:"certId"`
}

func AddWebProxyConfig(p *WebProxyConfig) error {
	err := Exec(`
				INSERT INTO web_proxy_config("ref_proxy_id","proxy","cert_id")
				VALUES (?, ?, ?);
			`, p.RefProxyId, p.Proxy)
	return err
}

func UpdateWebProxyConfig(p *WebProxyConfig) error {
	return Exec(`
				UPDATE web_proxy_config SET  "proxy"=?,"cert_id"=? WHERE "ref_proxy_id"=?;
			`, p.Proxy, p.CertId, p.RefProxyId)
}

func GetWebProxyConfig(refProxyId int) *WebProxyConfig {
	res, err := Query("select id,ref_proxy_id,proxy,cert_id from web_proxy_config where ref_proxy_id=?", refProxyId)
	if err != nil {
		return nil
	}
	defer res.Close()
	for res.rows.Next() {
		var p WebProxyConfig
		if err := res.rows.Scan(&p.Id, &p.RefProxyId, &p.Proxy, &p.CertId); err != nil {
			return nil
		}
		return &p
	}
	return nil
}
