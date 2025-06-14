package srv

import "sync/atomic"

// Here are some definitions of variables and constants

type Cmd int8

type PType int8

type RspCode int16

var counter int64

func increment() int64 {
	return atomic.AddInt64(&counter, 1)
}

const (
	// REQUEST InBound.
	REQUEST PType = 0

	// RESPONSE Receiver.
	RESPONSE PType = 1
)

// InBound cmd
const (
	// Heart Client request ping.
	Heart Cmd = 1

	// Register : Register tunnel port.
	Register Cmd = 2

	// Communication communication　connection.
	Communication Cmd = 3
)

// RspSuccess RspCode.
const (
	// RspSuccess success.
	RspSuccess RspCode = 0

	RspFail = 101
)

// Protocol
// @Description: Internal request
// struct.
type Protocol struct {
	Data []byte

	//reqId.
	ReqId int64

	//InBound cmd.
	Cmd Cmd

	// 0 request 1 response.
	PType PType

	//responseCode.
	//request never is zero.
	RspCode RspCode
}

func NewRequest(cmd Cmd, data []byte) Protocol {
	return Protocol{
		Data:    data,
		Cmd:     cmd,
		ReqId:   increment(),
		PType:   REQUEST,
		RspCode: RspSuccess,
	}
}

func NewResponse(cmd Cmd, reqId int64) Protocol {
	return Protocol{
		ReqId:   reqId,
		Cmd:     cmd,
		PType:   RESPONSE,
		RspCode: RspSuccess,
	}
}

// InBound
// @Description:  request.
type InBound interface {

	//
	// Cmd
	//  @Description: get cmd.
	//  @return Cmd
	//
	Cmd() Cmd
}

// RegisterReq
// @Description: Register Info.
type RegisterReq struct {
	TunnelPort int32 `json:"tunnel_port"`

	BindId string `json:"bind_id"`
}

func (r RegisterReq) Cmd() Cmd {

	return Register
}

// Heartbeat
// @Description: Ping InBound info. This is empty request,server use Cmd　discern.
type Heartbeat struct {
	Value string `json:"value"`
}

func (p Heartbeat) Cmd() Cmd {
	return Heart
}

// CommunicationInfo
// @Description: This connection use for to server communication.
type CommunicationInfo struct {
	//this bindId eq RegisterReq bindId.
	BindId string `json:"bind_id"`
}

func (c CommunicationInfo) Cmd() Cmd {
	return Communication
}
