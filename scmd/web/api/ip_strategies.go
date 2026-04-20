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

package api

import (
	"github.com/g-brook/brook/common/transform"
	"github.com/g-brook/brook/scmd/web/errs"
	"github.com/g-brook/brook/scmd/web/sql"
)

func init() {
	RegisterRoute(NewRoute("/strategies/getAll", "POST"), getStrategiesAll)
	RegisterRoute(NewRoute("/strategies/add", "POST"), addStrategy)
	RegisterRoute(NewRoute("/strategies/update", "POST"), updateStrategy)
	RegisterRoute(NewRoute("/strategies/del", "POST"), delStrategy)
}

func getStrategiesAll(*Request[any]) *Response {
	all, err := sql.SelectIpStrategyAll()
	if err != nil {
		return NewResponseSuccess(nil)
	}
	return NewResponseSuccess(fromIpStrategyDb(all))
}

func addStrategy(request *Request[IpStrategy]) *Response {
	body := request.Body
	if body.Name == "" {
		return NewResponseFail(errs.CodeSysErr, "name is empty")
	}
	if body.Type != "WL" && body.Type != "BL" && body.Type != "IL" {
		return NewResponseFail(errs.CodeSysErr, "type is invalid")
	}
	if body.BindHandler == "" {
		return NewResponseFail(errs.CodeSysErr, "bind_handler is empty")
	}
	if body.Status != 0 && body.Status != 1 {
		return NewResponseFail(errs.CodeSysErr, "status is invalid")
	}
	err := sql.AddIpStrategy(&sql.IpStrategy{
		Name:        body.Name,
		Type:        body.Type,
		BindHandler: body.BindHandler,
		Status:      body.Status,
	})
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "add strategy failed")
	}
	return NewResponseSuccess(nil)
}

func updateStrategy(request *Request[IpStrategy]) *Response {
	body := request.Body
	if body.Id <= 0 {
		return NewResponseFail(errs.CodeSysErr, "id is empty")
	}
	if body.Name == "" {
		return NewResponseFail(errs.CodeSysErr, "name is empty")
	}
	if body.Type != "WL" && body.Type != "BL" && body.Type != "IL" {
		return NewResponseFail(errs.CodeSysErr, "type is invalid")
	}
	if body.BindHandler == "" {
		return NewResponseFail(errs.CodeSysErr, "bind_handler is empty")
	}
	if body.Status != 0 && body.Status != 1 {
		return NewResponseFail(errs.CodeSysErr, "status is invalid")
	}
	err := sql.UpdateIpStrategy(&sql.IpStrategy{
		Id:          body.Id,
		Name:        body.Name,
		Type:        body.Type,
		BindHandler: body.BindHandler,
		Status:      body.Status,
	})
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "update strategy failed")
	}
	return NewResponseSuccess(nil)
}

func delStrategy(request *Request[IpStrategy]) *Response {
	body := request.Body
	if body.Id <= 0 {
		return NewResponseFail(errs.CodeSysErr, "id is empty")
	}
	_ = sql.DelIpRulesByStrategyId(body.Id)
	err := sql.DelIpStrategy(body.Id)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete strategy failed")
	}
	return NewResponseSuccess(nil)
}

func fromIpStrategyDb(dbs []*sql.IpStrategy) []*IpStrategy {
	converter := transform.NewConverter()
	var out []*IpStrategy
	err := converter.ConvertSlice(dbs, &out)
	if err != nil {
		return nil
	}
	return out
}
