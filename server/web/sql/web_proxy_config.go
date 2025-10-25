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

type WebProxyConfig struct {
	Id         string `db:"id" json:"id"`
	RefProxyId int    `db:"ref_proxy_id" json:"refProxyId"`
	CertFile   string `db:"cert_file" json:"certFile"`
	KeyFile    string `db:"key_file" json:"keyFile"`
	Proxy      string `db:"proxy" json:"proxy"`
}

func AddWebProxyConfig(p WebProxyConfig) error {
	err := Exec(`
				INSERT INTO web_proxy_config("ref_proxy_id","cert_file","key_file","proxy")
				VALUES (?, ?, ?, ?);
			`, p.RefProxyId, p.CertFile, p.KeyFile, p.Proxy)
	return err
}

func UpdateWebProxyConfig(p WebProxyConfig) error {
	return Exec(`
				UPDATE web_proxy_config SET "cert_file"=?, "key_file"=?, "proxy"=? WHERE "ref_proxy_id"=?;
			`, p.CertFile, p.KeyFile, p.Proxy, p.RefProxyId)
}

func GetWebProxyConfig(refProxyId int) *WebProxyConfig {
	res, err := Query("select id,ref_proxy_id,cert_file,key_file,proxy from web_proxy_config where ref_proxy_id=?", refProxyId)
	if err != nil {
		return nil
	}
	defer res.Close()
	for res.rows.Next() {
		var p WebProxyConfig
		if err := res.rows.Scan(&p.Id, &p.RefProxyId, &p.CertFile, &p.KeyFile, &p.Proxy); err != nil {
			return nil
		}
		return &p
	}
	return nil
}
