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
	"encoding/json"
	"io"
	"net/http"

	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/server/web/db"
	"github.com/g-brook/brook/server/web/errs"
)

type Request[T any] struct {
	Body T `json:"body"`
}

type Response struct {
	Code errs.Code `json:"code"`

	Message string `json:"message"`

	Data any `json:"data"`
}

func NewResponseSuccess(data any) *Response {
	return &Response{
		Code:    errs.CodeOk,
		Message: "success",
		Data:    data,
	}
}

func NewResponseFail(code errs.Code, msg string) *Response {
	return &Response{
		Code:    code,
		Message: msg,
	}
}

type WebHandlerFaction[T any] func(request *Request[T]) *Response

type handlerEntry[T any] struct {
	newRequest func(data []byte) (*Request[T], error)
	process    func(request *Request[T]) (*Response, error)
}

// getHandler is a generic function that creates a new WebHandler for a given function type
// T is a generic type parameter that represents the type of request body
// function is the WebHandlerFaction[T] function that will be processed by the handler
func getHandler[T any](function WebHandlerFaction[T], needAuth bool) *WebHandler[T] {
	// Create a new handlerEntry with request processing logic
	h := &handlerEntry[T]{
		newRequest: func(data []byte) (*Request[T], error) {
			var req T // Declare a variable of type T
			// If there's no data, return an empty Request
			if len(data) == 0 {
				return &Request[T]{
					Body: req,
				}, nil
			}
			// Unmarshal JSON data into the request body
			err := json.Unmarshal(data, &req)
			if err != nil {
				return nil, err // Return error if unmarshalling fails
			}
			// Return the new Request with unmarshalled data
			return &Request[T]{
				Body: req,
			}, err
		},
		// process is a function that processes the request and returns a response
		process: func(request *Request[T]) (*Response, error) {
			// Call the provided function with the request
			response := function(request)
			return response, nil // Return the response and no error
		},
	}
	return &WebHandler[T]{
		// Create and return a new WebHandler with the configured handlerEntry
		handlerEntry: h,
		needAuth:     needAuth,
	}
}

type WebHandler[T any] struct {
	handlerEntry *handlerEntry[T]
	needAuth     bool
}

func (w *WebHandler[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if w.needAuth {
		auth := request.Header.Get("Authorization")
		if auth == "" {
			writeAuthError(writer)
			return
		}
		info, err := db.Get[UserInfo](auth)
		if err != nil || info == nil {
			writeAuthError(writer)
			return
		}
		if info.Username != "cli" {
			//update ttl
			updateTtl(auth)
		}
	}
	// 读取请求 body
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return
	}
	defer request.Body.Close()
	req, err := w.handlerEntry.newRequest(body)
	if err != nil {
		writeError(writer)
		return
	}
	rsp, err := w.handlerEntry.process(req)
	if err != nil {
		writeError(writer)
		return
	}
	marshal, err := json.Marshal(rsp)
	if err != nil {
		writeError(writer)
		return
	}
	_, _ = writer.Write(marshal)
}

func updateTtl(auth string) {
	db.UpdateTTL(auth, TokenTtl)
}

func writeError(writer http.ResponseWriter) {
	log.Error("system error....")
	fail := NewResponseFail(errs.CodeSysErr, "system error")
	marshal, _ := json.Marshal(fail)
	_, _ = writer.Write(marshal)
}

func writeAuthError(writer http.ResponseWriter) {
	log.Error("system error....")
	fail := NewResponseFail(errs.CodeNotAuth, "not authorization")
	marshal, _ := json.Marshal(fail)
	_, _ = writer.Write(marshal)
}
