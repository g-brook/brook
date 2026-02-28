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
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	sql2 "database/sql"
	"encoding/pem"
	"fmt"

	"github.com/g-brook/brook/common/transform"
	"github.com/g-brook/brook/scmd/web/errs"
	"github.com/g-brook/brook/scmd/web/sql"
	"golang.org/x/crypto/ed25519"
)

func init() {
	RegisterRoute(NewRoute("/getCertificates", "POST"), getCertificates)
	RegisterRoute(NewRoute("/getCertificateById", "POST"), getCertificateById)
	RegisterRoute(NewRoute("/addCertificate", "POST"), addCertificate)
	RegisterRoute(NewRoute("/deleteCertificate", "POST"), deleteCertificate)
}

func deleteCertificate(req *Request[Certificate]) *Response {
	err := sql.DeleteCertificate(req.Body.ID)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete certificate error")
	}
	return NewResponseSuccess(nil)
}
func getCertificateById(req *Request[Certificate]) *Response {
	ft, err := sql.GetCertificateByID(req.Body.ID)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "get certificate error")
	}
	var ct Certificate
	converter := transform.NewConverter()
	err = converter.Convert(ft, &ct)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "get certificate error")
	}
	ct.ExpireTime = convertStringToPointer(ft.ExpireTime)
	return NewResponseSuccess(ct)
}

func addCertificate(req *Request[Certificate]) *Response {
	body := req.Body
	p, _ := pem.Decode([]byte(body.Content))
	if body.Name == "" {
		return NewResponseFail(errs.CodeSysErr, "name is null")
	}
	if body.Desc == "" {
		return NewResponseFail(errs.CodeSysErr, "description is null")
	}
	if p == nil {
		return NewResponseFail(errs.CodeSysErr, "certificate is empty")
	}
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "certificate format error")
	}

	keyBlock, _ := pem.Decode([]byte(body.PrivateKey))
	if keyBlock == nil {
		return NewResponseFail(errs.CodeSysErr, "private key is null")
	}
	var privKey any
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY": // PKCS#8
		privKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	default:
		return NewResponseFail(errs.CodeSysErr, "private key type error")
	}

	if !matchPublicKey(cert.PublicKey, privKey) {
		return NewResponseFail(errs.CodeSysErr, "certificate and private key not match")
	}
	db := body.toDb()
	if db == nil {
		return NewResponseFail(errs.CodeSysErr, "")
	}
	db.ExpireTime = sql2.NullString{
		String: cert.NotAfter.Format("2006-01-02 15:04:05"),
		Valid:  true,
	}
	err = sql.AddCertificate(db)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "add certificate error")
	}
	return NewResponseSuccess(nil)
}

func matchPublicKey(certPub any, priv any) bool {
	switch key := priv.(type) {
	case *rsa.PrivateKey:
		return certPub.(*rsa.PublicKey).N.Cmp(key.PublicKey.N) == 0
	case *ecdsa.PrivateKey:
		return certPub.(*ecdsa.PublicKey).X.Cmp(key.PublicKey.X) == 0 &&
			certPub.(*ecdsa.PublicKey).Y.Cmp(key.PublicKey.Y) == 0
	case ed25519.PrivateKey:
		return certPub.(ed25519.PublicKey).Equal(key.Public().(ed25519.PublicKey))
	default:
		return false
	}
}

func getCertificates(req *Request[Certificate]) *Response {
	certificates, err := sql.GetAllCertificates()
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "get certificates error")
	}
	converter := transform.NewConverter()
	var outputs []*Certificate
	for _, cert := range certificates {
		var ct Certificate
		err = converter.Convert(cert, &ct)
		if err != nil {
			break
		}
		ct.ExpireTime = convertStringToPointer(cert.ExpireTime)
		outputs = append(outputs, &ct)
	}
	if err != nil {
		fmt.Println(err.Error())
		return NewResponseFail(errs.CodeSysErr, "get certificates error")
	}
	return NewResponseSuccess(outputs)
}
