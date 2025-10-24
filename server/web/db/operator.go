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

package db

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/brook/common/log"
	"github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

type dbLog struct {
}

func (d dbLog) Errorf(s string, i ...interface{}) {
	log.Error("db:"+s, i)
}

func (d dbLog) Warningf(s string, i ...interface{}) {
	log.Warn("db:"+s, i)
}

func (d dbLog) Infof(s string, i ...interface{}) {
	log.Info("db:"+s, i)
}

func (d dbLog) Debugf(s string, i ...interface{}) {
	log.Info("db:"+s, i)
}

func Open() {
	var err error
	options := badger.DefaultOptions("./fdb")
	options.Logger = dbLog{}
	DB, err = badger.Open(options)
	if err != nil {
		log.Debug("open db err", err)
	}
}

func Close() {
	_ = DB.Close()
}

// Put stores a key-value pair in the database
// key: the string key to store the value under
// value: the value to store (can be any type that can be marshaled to JSON)
// returns an error if the key is empty or value is nil
func Put(key string, value any) error {
	// Validate input parameters
	if value == nil || key == "" {
		return errors.New("key or value is nil")
	}
	// Marshal the value to JSON
	data, _ := json.Marshal(value)
	// Update the database with the key-value pair
	return DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func PutWithTtl(key string, value any, ttl time.Duration) error {
	// Validate input parameters
	if value == nil || key == "" {
		return errors.New("key or value is nil")
	}
	// Marshal the value to JSON
	data, _ := json.Marshal(value)
	// Update the database with the key-value pair
	return DB.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte(key), data).WithTTL(ttl))
	})
}

func Get[T any](key string) (*T, error) {
	var value *T
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		value = new(T)
		return json.Unmarshal(data, value)
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func Delete(key string) error {
	return DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func UpdateTTL(auth string, ttl time.Duration) {
	_ = DB.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(auth))
		if err != nil {
			return err
		}
		valueCopy, err := item.ValueCopy(nil)
		_ = txn.SetEntry(badger.NewEntry([]byte(auth), valueCopy).WithTTL(ttl))
		return nil
	})
}
