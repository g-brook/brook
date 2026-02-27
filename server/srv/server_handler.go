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

package srv

import (
	trp "github.com/g-brook/brook/common/transport"
)

type BootServer interface {
	Start(opt ...ServerOption) error
}

type ServerHandler interface {
	//
	// Close
	//  @Description: Shutdown conn notify.
	//  @param conn
	//
	Close(ch trp.Channel, traverse TraverseBy) error

	//
	// Open
	//  @Description: Open conn notify.
	//  @param conn
	//
	Open(ch trp.Channel, traverse TraverseBy) error

	//
	// Reader
	//  @Description: Reader conn data notify.
	//  @param conn
	//
	Reader(ch trp.Channel, traverse TraverseBy) error

	Error(ch trp.Channel, error error, traverse TraverseBy)

	Boot(server BootServer, traverse TraverseBy) error
}

type BaseServerHandler struct {
}

func (b *BaseServerHandler) Close(_ trp.Channel, traverse TraverseBy) error {
	traverse()
	return nil
}

func (b *BaseServerHandler) Open(_ trp.Channel, traverse TraverseBy) error {
	traverse()
	return nil
}

func (b *BaseServerHandler) Reader(_ trp.Channel, traverse TraverseBy) error {
	traverse()
	return nil
}

func (b *BaseServerHandler) Boot(_ BootServer, traverse TraverseBy) error {
	traverse()
	return nil
}

func (b *BaseServerHandler) Error(_ trp.Channel, _ error, traverse TraverseBy) {
	traverse()
}
