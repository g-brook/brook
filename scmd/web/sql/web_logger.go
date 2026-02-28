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

	"github.com/g-brook/brook/common/log"
)

type DBWebLogger struct {
	Id       int            `db:"id" json:"id"`
	Protocol string         `db:"protocol" json:"protocol"`
	Path     string         `db:"path" json:"path"`
	Host     string         `db:"host" json:"host"`
	Method   string         `db:"method" json:"method"`
	Status   int            `db:"status" json:"status"`
	ProxyId  string         `db:"proxy_id" json:"proxyId"`
	HttpId   string         `db:"http_id" json:"httpId"`
	Time     sql.NullString `db:"time" json:"time"`
}

func AddWebLog(log *DBWebLogger) error {
	err := Exec(`
            INSERT INTO web_logger(protocol, path, host, method, status, proxy_id, http_id,time)
            VALUES (?, ?, ?, ?, ?,?,?,?);
        `, log.Protocol, log.Path, log.Host, log.Method, log.Status, log.ProxyId, log.HttpId, log.Time)
	return err
}

func QueryWebLogByProxyId(proxyId string) []*DBWebLogger {
	res, err := Query("select protocol, path, host, method, status, proxy_id, http_id,time from web_logger where proxy_id=? order by id desc limit 100", proxyId)
	if err != nil {
		return nil
	}
	defer res.Close()
	var list []*DBWebLogger
	for res.rows.Next() {
		var p DBWebLogger
		if err := res.rows.Scan(&p.Protocol, &p.Path, &p.Host, &p.Method, &p.Status, &p.ProxyId, &p.HttpId, &p.Time); err != nil {
			log.Error("query web log error %v", err)
			return nil
		}
		list = append(list, &p)
	}
	return list
}
