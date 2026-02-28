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
	"database/sql"
	"time"
)

const (
	TokenTtl = 7 * 24 * time.Hour
)

// convertInt32ToPointer 将 int32 转换为 *int
func convertInt32ToPointer(nullInt32 sql.NullInt32) *int {
	if !nullInt32.Valid {
		return nil
	}
	value := nullInt32.Int32
	v := int(value)
	return &v
}

func convertStringToPointer(nullString sql.NullString) *string {
	if !nullString.Valid {
		return nil
	}
	value := nullString.String
	return &value
}

func convertInt64ToPointer(nullInt64 sql.NullInt64) *int64 {
	if !nullInt64.Valid {
		return nil
	}
	return &nullInt64.Int64
}

func convertInt16ToPointer(nullInt16 sql.NullInt16) *int16 {
	if !nullInt16.Valid {
		return nil
	}
	return &nullInt16.Int16
}

func convertBoolToPointer(nullBool sql.NullBool) *bool {
	if !nullBool.Valid {
		return nil
	}
	return &nullBool.Bool
}

func convertFloat64ToPointer(nullFloat64 sql.NullFloat64) *float64 {
	if !nullFloat64.Valid {
		return nil
	}
	return &nullFloat64.Float64
}

func convertTimeToPointer(nullTime sql.NullTime) *time.Time {
	if !nullTime.Valid {
		return nil
	}
	return &nullTime.Time
}
