package api

import (
	"time"

	"github.com/g-brook/brook/common/transform"
	"github.com/g-brook/brook/scmd/web/errs"
	"github.com/g-brook/brook/scmd/web/sql"
)

type IpRule struct {
	Id         int16     `json:"id" maps:"id"`
	StrategyId int16     `json:"strategyId" maps:"strategy_id"`
	Ip         string    `json:"ip" maps:"ip"`
	Remark     string    `json:"remark" maps:"remark"`
	CreatedAt  time.Time `json:"created_at" maps:"created_at"`
}

type QueryIpRule struct {
	StrategyId int16 `json:"strategyId"`
}

type DelIpRuleReq struct {
	Id int16 `json:"id"`
}

func init() {
	RegisterRoute(NewRoute("/rules/getByStrategyId", "POST"), getRulesByStrategyId)
	RegisterRoute(NewRoute("/rules/add", "POST"), addRule)
	RegisterRoute(NewRoute("/rules/del", "POST"), delRule)
}

func getRulesByStrategyId(req *Request[QueryIpRule]) *Response {
	if req.Body.StrategyId <= 0 {
		return NewResponseFail(errs.CodeSysErr, "strategyId is empty")
	}
	list, err := sql.SelectByStrategyId(req.Body.StrategyId)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "get rules failed")
	}
	converter := transform.NewConverter()
	var out []*IpRule
	err = converter.ConvertSlice(list, &out)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "convert rules failed")
	}
	return NewResponseSuccess(out)
}

func addRule(req *Request[IpRule]) *Response {
	body := req.Body
	if body.StrategyId <= 0 {
		return NewResponseFail(errs.CodeSysErr, "strategyId is empty")
	}
	if body.Ip == "" {
		return NewResponseFail(errs.CodeSysErr, "ip is empty")
	}
	_, err := sql.AddIpRule(body.StrategyId, body.Ip, body.Remark)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "add rule failed")
	}
	return NewResponseSuccess(nil)
}

func delRule(req *Request[DelIpRuleReq]) *Response {
	if req.Body.Id <= 0 {
		return NewResponseFail(errs.CodeSysErr, "id is empty")
	}
	err := sql.DelIpRule(req.Body.Id)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete rule failed")
	}
	return NewResponseSuccess(nil)
}
