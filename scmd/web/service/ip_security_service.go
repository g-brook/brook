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

package service

import (
	"strconv"

	"github.com/g-brook/brook/scmd/web/sql"
)

type IpSecurity struct {
	Ips      []string `json:"ips"`
	Strategy string   `json:"strategy"`
	Name     string   `json:"name"`
}

func SelectIpSecurity(ipStrategies string) (*IpSecurity, error) {
	id, err := strconv.Atoi(ipStrategies)
	if err != nil {
		return nil, nil
	}
	strategy, err := sql.SelectIpStrategyById(int16(id))
	if err != nil || strategy == nil {
		return nil, err
	}
	ipRule, err := sql.SelectByStrategyId(strategy.Id)
	if err != nil || ipRule == nil {
		return &IpSecurity{
			Strategy: strategy.Type,
			Name:     strategy.Name,
		}, nil
	}
	var ips []string
	for _, rules := range ipRule {
		ip := rules.Ip
		ips = append(ips, ip)
	}
	ip := &IpSecurity{
		Ips:      ips,
		Strategy: strategy.Type,
		Name:     strategy.Name,
	}
	return ip, nil
}
