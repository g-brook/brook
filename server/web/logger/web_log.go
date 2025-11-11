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

package logger

import (
	sql2 "database/sql"
	"time"

	"github.com/brook/common/log"
	"github.com/brook/server/web/sql"
)

type WebLogger struct {
	Protocol string    `json:"protocol"`
	Path     string    `json:"path"`
	Host     string    `json:"host"`
	Method   string    `json:"method"`
	Status   int       `json:"status"`
	ProxyId  string    `json:"proxyId"`
	HttpId   string    `json:"httpId"`
	Time     time.Time `json:"time"`
}

func WithWebLog(logger *WebLogger) {
	log.Debug("info %v,%v,%v,%v,%v,%v,%v", logger.ProxyId, logger.Protocol, logger.Path, logger.Host, logger.Method, logger.Status, logger.HttpId)
	_ = sql.AddWebLog(&sql.DBWebLogger{
		Protocol: logger.Protocol,
		Path:     logger.Path,
		Host:     logger.Host,
		Method:   logger.Method,
		Status:   logger.Status,
		ProxyId:  logger.ProxyId,
		HttpId:   logger.HttpId,
		Time:     sql2.NullString{String: logger.Time.Format("2006-01-02 15:04:05"), Valid: true},
	})
}
