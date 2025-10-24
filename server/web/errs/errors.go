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
