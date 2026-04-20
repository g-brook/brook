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

type IpRules struct {
	Id         int16     `db:"id"`
	StrategyId int16     `db:"strategy_id"`
	Ip         string    `db:"ip"`
	Remark     string    `db:"remark"`
	CreateAt   time.Time `db:"created_at"`
}

var ipRulesSql = "id,strategy_id,ip,remark,created_at"

func AddIpRule(strategyId int16, ip string, remark string) (int64, error) {
	return ExecWithId(
		`INSERT INTO ip_rules(strategy_id, ip, remark, created_at)
         VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		strategyId,
		ip,
		remark,
	)
}

func DelIpRule(id int16) error {
	return Exec("DELETE FROM ip_rules WHERE id = ?", id)
}

func DelIpRulesByStrategyId(strategyId int16) error {
	return Exec("DELETE FROM ip_rules WHERE strategy_id = ?", strategyId)
}

func SelectByStrategyId(strategyId int16) ([]*IpRules, error) {
	selectSQL := fmt.Sprintf("select %s from ip_rules where strategy_id = ?", ipRulesSql)
	res, err := Query(selectSQL, strategyId)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var list []*IpRules
	for res.rows.Next() {
		if p, err := scanIpRules(res.rows); err != nil {
			return nil, err
		} else {
			list = append(list, p)
		}
	}
	return list, nil
}

func scanIpRules(rows *sql.Rows) (*IpRules, error) {
	var p IpRules
	err := rows.Scan(
		&p.Id,
		&p.StrategyId,
		&p.Ip,
		&p.Remark,
		&p.CreateAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
