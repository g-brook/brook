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

package sql

import (
	"database/sql"
	"fmt"
)

const CreateUsersTableSQL = `create table users
(
    id       integer
        constraint users_pk
            primary key autoincrement,
    user_id  text
        constraint users_pk_2
            unique,
    password text,
    icon     text,
    is_admin bool
);`

type Users struct {
	Id       int64  `db:"id" maps:"id"`
	UserId   string `db:"user_id" maps:"user_id"`
	Password string `db:"password" maps:"password"`
	Icon     string `db:"icon" maps:"icon"`
	IsAdmin  bool   `db:"is_admin" maps:"is_admin"`
}

var usersQuerySQL = "id,user_id,password,icon,is_admin"

func AddUser(u *Users) (error, int64) {
	id, err := ExecWithId(
		`INSERT INTO users(user_id, password, icon, is_admin) VALUES (?, ?, ?, ?)`,
		u.UserId,
		u.Password,
		u.Icon,
		u.IsAdmin,
	)
	return err, id
}

func UpdateUser(u *Users) error {
	return Exec(
		`UPDATE users SET user_id = ?, password = ?, icon = ?, is_admin = ? WHERE id = ?`,
		u.UserId,
		u.Password,
		u.Icon,
		u.IsAdmin,
		u.Id,
	)
}

func DeleteUser(id int64) error {
	return Exec(`DELETE FROM users WHERE id = ?`, id)
}

func GetUserById(id int64) (*Users, error) {
	query := fmt.Sprintf("select %s from users where id = ?", usersQuerySQL)
	res, err := Query(query, id)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.rows.Next() {
		return scanUser(res.rows)
	}
	return nil, nil
}

func GetUserByUserId(userId string) (*Users, error) {
	query := fmt.Sprintf("select %s from users where user_id = ?", usersQuerySQL)
	res, err := Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.rows.Next() {
		return scanUser(res.rows)
	}
	return nil, nil
}

func GetAllUsers() ([]*Users, error) {
	query := fmt.Sprintf("select %s from users", usersQuerySQL)
	res, err := Query(query)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var list []*Users
	for res.rows.Next() {
		u, err := scanUser(res.rows)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}

func scanUser(rows *sql.Rows) (*Users, error) {
	var u Users
	err := rows.Scan(
		&u.Id,
		&u.UserId,
		&u.Password,
		&u.Icon,
		&u.IsAdmin,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
