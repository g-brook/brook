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

package transform

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// Converter 结构体转换器
type Converter struct {
	config *mapstructure.DecoderConfig
}

// NewConverter 创建新的转换器
func NewConverter() *Converter {
	return &Converter{
		config: &mapstructure.DecoderConfig{
			Metadata:         nil,
			Result:           nil,
			TagName:          "maps",
			WeaklyTypedInput: true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
				mapstructure.StringToSliceHookFunc(","),
				mapstructure.TextUnmarshallerHookFunc(),
			),
		},
	}
}

// WithWeaklyTyped 设置弱类型转换
func (c *Converter) WithWeaklyTyped(enable bool) *Converter {
	c.config.WeaklyTypedInput = enable
	return c
}

// WithTagName 设置自定义标签名
func (c *Converter) WithTagName(tagName string) *Converter {
	c.config.TagName = tagName
	return c
}

// WithDecodeHook 添加自定义解码钩子
func (c *Converter) WithDecodeHook(hook mapstructure.DecodeHookFunc) *Converter {
	c.config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		c.config.DecodeHook,
		hook,
	)
	return c
}

// MapToStruct 将 map 转换为结构体
func (c *Converter) MapToStruct(input map[string]interface{}, output interface{}) error {
	config := *c.config
	config.Result = output
	config.Metadata = nil

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	return decoder.Decode(input)
}

// StructToMap 将结构体转换为 map（支持嵌套结构体）
func (c *Converter) StructToMap(input interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 跳过不可导出的字段
		if !field.IsExported() {
			continue
		}

		tagName := field.Tag.Get(c.config.TagName)
		if tagName == "" {
			tagName = field.Name
		}

		// 处理嵌套结构体
		if fieldValue.Kind() == reflect.Struct && field.Anonymous {
			// 嵌入字段，展开
			embeddedMap, err := c.StructToMap(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			for k, v := range embeddedMap {
				result[k] = v
			}
		} else if tagName != "-" { // 跳过标记为忽略的字段
			// 递归处理嵌套结构体
			if fieldValue.Kind() == reflect.Struct ||
				(fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() &&
					fieldValue.Elem().Kind() == reflect.Struct) {

				if fieldValue.Kind() == reflect.Ptr {
					nestedMap, err := c.StructToMap(fieldValue.Interface())
					if err != nil {
						return nil, err
					}
					result[tagName] = nestedMap
				} else {
					nestedMap, err := c.StructToMap(fieldValue.Interface())
					if err != nil {
						return nil, err
					}
					result[tagName] = nestedMap
				}
			} else {
				result[tagName] = fieldValue.Interface()
			}
		}
	}

	return result, nil
}

// Convert 通用转换方法，支持任意类型到任意类型
func (c *Converter) Convert(input, output interface{}) error {
	// 如果输入是 map，直接使用 MapToStruct
	if m, ok := input.(map[string]interface{}); ok {
		return c.MapToStruct(m, output)
	}

	// 先转换为 map，再转换为目标结构体
	inputMap, err := c.StructToMap(input)
	if err != nil {
		return fmt.Errorf("failed to convert input to map: %w", err)
	}

	return c.MapToStruct(inputMap, output)
}

// MustConvert 安全转换，如果出错会 panic
func (c *Converter) MustConvert(input, output interface{}) {
	if err := c.Convert(input, output); err != nil {
		panic(fmt.Sprintf("Convert failed: %v", err))
	}
}

// ConvertSlice 转换切片
func (c *Converter) ConvertSlice(inputSlice interface{}, outputSlice interface{}) error {
	inputVal := reflect.ValueOf(inputSlice)
	if inputVal.Kind() != reflect.Slice {
		return fmt.Errorf("input must be a slice")
	}

	outputVal := reflect.ValueOf(outputSlice)
	if outputVal.Kind() != reflect.Ptr || outputVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("output must be a pointer to slice")
	}

	sliceType := outputVal.Elem().Type()
	elemType := sliceType.Elem()
	length := inputVal.Len()

	resultSlice := reflect.MakeSlice(sliceType, length, length)

	for i := 0; i < length; i++ {
		inputElem := inputVal.Index(i).Interface()
		outputElem := reflect.New(elemType).Interface()

		if err := c.Convert(inputElem, outputElem); err != nil {
			return fmt.Errorf("failed to convert element at index %d: %w", i, err)
		}

		resultSlice.Index(i).Set(reflect.ValueOf(outputElem).Elem())
	}

	outputVal.Elem().Set(resultSlice)
	return nil
}

// SnakeToCamel 蛇形转驼峰
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// CamelToSnake 驼峰转蛇形
func CamelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

var defaultConverter = NewConverter()

// ToMap 将结构体转换为 map（使用默认转换器）
func ToMap(input interface{}) (map[string]interface{}, error) {
	return defaultConverter.StructToMap(input)
}

// ToStruct 将 map 转换为结构体（使用默认转换器）
func ToStruct(input map[string]interface{}, output interface{}) error {
	return defaultConverter.MapToStruct(input, output)
}

// Transform 通用转换（使用默认转换器）
func Transform(input, output interface{}) error {
	return defaultConverter.Convert(input, output)
}
