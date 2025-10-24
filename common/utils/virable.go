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

package utils

type TunnelType string
type Network string

const (
	Http  TunnelType = "http"
	Https TunnelType = "https"
	Tcp   TunnelType = "tcp"
	Udp   TunnelType = "udp"
)

const NetworkTcp = Network(Tcp)

const NetworkUdp = Network(Udp)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// NilDefault NilDefault[T comparable]
//
//	@Description:
//	@param value
//	@param defaultValue
//	@return T
func NilDefault[T comparable](value T, defaultValue T) T {
	return DefaultValue(!IsNotNil(value), value, defaultValue)
}

// NumberDefault NumberDefault[T Number]
//
//	@Description: If value not eq 0 return value, if not defaultValue.
//	@param value
//	@param defaultValue
//	@return T
func NumberDefault[T Number](value T, defaultValue T) T {
	return DefaultValue(value != 0, value, defaultValue)
}

func IsNotNil[T comparable](value T) bool {
	var zero T
	if zero == value {
		return false
	}
	return true
}

// DefaultValue is a generic function that returns either a value or a default based on a boolean condition
// It uses Go's generics feature to work with any type T
//
// Parameters:
//
//	b: boolean condition to determine which value to return
//	val: the value to return if b is true
//	def: the default value to return if b is false
//
// Returns:
//
//	T: either val or def based on the boolean condition b
func DefaultValue[T any](b bool, val T, def T) T {
	if b { // if condition is true, return the provided value
		return val
	}
	return def // otherwise return the default value
}
