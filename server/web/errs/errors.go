package errs

import (
	"errors"
)

type Code string

const (
	CodeOk       Code = "OK"
	CodeSysErr   Code = "SYS_ERR"
	CodeInternal Code = "CODE_INTERNAL"
	CodeNotAuth  Code = "NOT_ATH"
)

type E struct {
	Code Code
	Msg  string
	Err  error
}

func (e *E) Error() string {
	switch {
	case e.Msg != "":
		return e.Msg
	default:
		if e.Err != nil {
			return e.Err.Error()
		}
		return string(e.Code)
	}
}

func New(code Code, msg string) *E {
	e := &E{Code: code, Msg: msg}
	return e
}

func CodeOf(err error) Code {
	var e *E
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeInternal
}
