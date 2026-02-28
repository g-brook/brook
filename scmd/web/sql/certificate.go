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
)

// Certificate 证书结构体
type Certificate struct {
	ID         int            `db:"id" maps:"id"`
	Name       string         `db:"name" maps:"name"`
	Content    string         `db:"content" maps:"content"`
	PrivateKey string         `db:"private_key" maps:"private_key"`
	Desc       string         `db:"desc" maps:"desc"`
	ExpireTime sql.NullString `db:"expireTime" maps:"-"`
}

// AddCertificate 添加证书
func AddCertificate(cert *Certificate) error {
	query := `INSERT INTO certificate (name, content, private_key, desc,expire_time) VALUES (?, ?, ?, ?,?)`
	err := Exec(query, cert.Name, cert.Content, cert.PrivateKey, cert.Desc, cert.ExpireTime)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCertificate 删除证书
func DeleteCertificate(id int) error {
	query := `DELETE FROM certificate WHERE id = ?`
	err := Exec(query, id)
	return err
}

// GetAllCertificates 查询全部证书
func GetAllCertificates() ([]*Certificate, error) {
	query := `SELECT id, name, content, private_key, desc,expire_time FROM certificate`
	res, err := Query(query)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var certificates []*Certificate
	for res.rows.Next() {
		cert := &Certificate{}
		err := res.rows.Scan(&cert.ID, &cert.Name, &cert.Content, &cert.PrivateKey, &cert.Desc, &cert.ExpireTime)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, cert)
	}
	return certificates, nil
}

// GetCertificateByID 根据ID查询证书
func GetCertificateByID(id int) (*Certificate, error) {
	query := `SELECT id, name, content, private_key, desc,expire_time FROM certificate WHERE id = ?`
	rs, err := Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	cert := &Certificate{}
	if rs.rows.Next() {
		err = rs.rows.Scan(&cert.ID, &cert.Name, &cert.Content, &cert.PrivateKey, &cert.Desc, &cert.ExpireTime)
		if err != nil {
			return nil, err
		}
	}
	return cert, nil
}
