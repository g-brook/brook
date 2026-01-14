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
	"testing"
	"time"
)

// ==================== 测试方法 ====================

// TestConverter 对转换器进行完整测试
func TestConverter(t *testing.T) {
	fmt.Println("开始测试 Converter...")

	// 定义测试结构体
	type Address struct {
		City    string `maps:"city"`
		ZipCode string `maps:"zip_code"`
	}

	type Person struct {
		Name    string  `maps:"name"`
		Age     int     `maps:"age"`
		Address Address `maps:"address"`
		Email   string  `maps:"email"`
		Ignore  string  `maps:"-"`
	}

	// 1. 测试 StructToMap 功能
	fmt.Println("\n1. 测试 StructToMap...")
	person := Person{
		Name:   "张三",
		Age:    30,
		Email:  "zhangsan@example.com",
		Ignore: "should_be_ignored",
		Address: Address{
			City:    "北京",
			ZipCode: "100000",
		},
	}

	converter := NewConverter()
	personMap, err := converter.StructToMap(person)
	if err != nil {
		fmt.Printf("StructToMap 错误: %v\n", err)
	} else {
		fmt.Printf("StructToMap 结果: %+v\n", personMap)
	}

	// 2. 测试 MapToStruct 功能
	fmt.Println("\n2. 测试 MapToStruct...")
	newPerson := Person{}
	err = converter.MapToStruct(personMap, &newPerson)
	if err != nil {
		fmt.Printf("MapToStruct 错误: %v\n", err)
	} else {
		fmt.Printf("MapToStruct 结果: %+v\n", newPerson)
	}

	// 3. 测试 Convert 功能
	fmt.Println("\n3. 测试 Convert...")
	testMap := map[string]interface{}{
		"name":  "李四",
		"age":   25,
		"email": "lisi@example.com",
		"address": map[string]interface{}{
			"city":     "上海",
			"zip_code": "200000",
		},
	}

	convertedPerson := Person{}
	err = converter.Convert(testMap, &convertedPerson)
	if err != nil {
		fmt.Printf("Convert 错误: %v\n", err)
	} else {
		fmt.Printf("Convert 结果: %+v\n", convertedPerson)
	}

	// 4. 测试切片转换
	fmt.Println("\n4. 测试 ConvertSlice...")
	peopleMaps := []map[string]interface{}{
		{
			"name":  "王五",
			"age":   28,
			"email": "wangwu@example.com",
			"address": map[string]interface{}{
				"city":     "广州",
				"zip_code": "510000",
			},
		},
		{
			"name":  "赵六",
			"age":   32,
			"email": "zhaoliu@example.com",
			"address": map[string]interface{}{
				"city":     "深圳",
				"zip_code": "518000",
			},
		},
	}

	var people []Person
	err = converter.ConvertSlice(peopleMaps, &people)
	if err != nil {
		fmt.Printf("ConvertSlice 错误: %v\n", err)
	} else {
		fmt.Printf("ConvertSlice 结果: %+v\n", people)
	}

	// 5. 测试全局函数
	fmt.Println("\n5. 测试全局函数...")
	globalMap, err := ToMap(person)
	if err != nil {
		fmt.Printf("ToMap 错误: %v\n", err)
	} else {
		fmt.Printf("ToMap 结果: %+v\n", globalMap)
	}

	newPerson2 := Person{}
	err = ToStruct(globalMap, &newPerson2)
	if err != nil {
		fmt.Printf("ToStruct 错误: %v\n", err)
	} else {
		fmt.Printf("ToStruct 结果: %+v\n", newPerson2)
	}

	// 6. 测试字符串转换函数
	fmt.Println("\n6. 测试字符串转换函数...")
	snakeStr := "hello_world_test"
	camelStr := SnakeToCamel(snakeStr)
	fmt.Printf("SnakeToCamel('%s') = '%s'\n", snakeStr, camelStr)

	camelStr2 := "HelloWorldTest"
	snakeStr2 := CamelToSnake(camelStr2)
	fmt.Printf("CamelToSnake('%s') = '%s'\n", camelStr2, snakeStr2)

	// 7. 测试时间转换（如果有时间字段）
	fmt.Println("\n7. 测试时间转换...")
	type TimeTest struct {
		CreatedAt time.Time `maps:"created_at"`
		Name      string    `maps:"name"`
	}

	timeMap := map[string]interface{}{
		"created_at": "2023-12-25 10:30:45",
		"name":       "time_test",
	}

	var timeObj TimeTest
	err = converter.MapToStruct(timeMap, &timeObj)
	if err != nil {
		fmt.Printf("时间转换错误: %v\n", err)
	} else {
		fmt.Printf("时间转换结果: %+v\n", timeObj)
	}

	fmt.Println("\nConverter 测试完成!")
}
