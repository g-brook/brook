package exchange

import (
	"encoding/json"
	"errors"
	"sync/atomic"
)

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

	//QueryTunnel Query tunnel config.
	QueryTunnel = 4

	// OpenTunnel Open Tunnel
	OpenTunnel = 5
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

// NewRequest
//
//	@Description: new request Protocol.
//	@param cmd
//	@param data
//	@return *Protocol
//	@return error
func NewRequest(data InBound) (*Protocol, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("new Request error," + err.Error())
	}
	return &Protocol{
		Data:    b,
		ReqId:   increment(),
		Cmd:     data.Cmd(),
		PType:   REQUEST,
		RspCode: RspSuccess,
	}, nil
}

// NewResponse
//
//	@Description: new response Protocol.
//	@param cmd
//	@param reqId
//	@return *Protocol
//	@return error
func NewResponse(cmd Cmd, reqId int64) (*Protocol, error) {
	return &Protocol{
		ReqId:   reqId,
		Cmd:     cmd,
		PType:   RESPONSE,
		RspCode: RspSuccess,
	}, nil
}

// Bytes
//
//	@Description: to bytes.
//	@receiver receiver
//	@return []byte
func (receiver *Protocol) Bytes() []byte {
	return Encoder(receiver)
}

// Parse
//
//	@Description: parse
//	@receiver receiver
//	@param value
//	@return error
func Parse[T any](data []byte) (*T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
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
