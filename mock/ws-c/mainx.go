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

package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func main() {
	st := "-----BEGIN CERTIFICATE-----\nMIIDJTCCAg2gAwIBAgIUb4hXrtvrl2/VnE5s9zyaFBPdkt0wDQYJKoZIhvcNAQEL\nBQAwFDESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTI1MDgxOTAyNTQyOVoXDTI2MDgx\nOTAyNTQyOVowFDESMBAGA1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEF\nAAOCAQ8AMIIBCgKCAQEAxzCVTpdZ0cNg1hallpxzRUKzhf3vCJ1UvUyTDNahAvWb\nupHwAJ2ukI4TPi5eLFO/b+sQFw8tJhagpSFKW0ite6mxo5KbJMR0w0DSGO7Y/TA7\nrJgqJbRuRovrMCm6/S9MzAHsZVpCpFCMULjsK2NSg0DSbOSvJ6MD6qneK5AxxlVz\ncfUyl70VpngIWELemGWizniKXqPAzqp4ZyPSaJ4WWmn5mfPoNWVnI4mPBunbjUKr\noPmzzGIYsbVBNWvE2XJUbUnQmhrV8u+Rre5jShLPWWKg3u4vr5p/zvzMc9GY2Lz7\nkIs9iDsdNI+enZuWbdh0tOZMo6iG6FxVJU/xQPgp7QIDAQABo28wbTAdBgNVHQ4E\nFgQU68nOcpLyCOfns/PGOCsxQY7jDHcwHwYDVR0jBBgwFoAU68nOcpLyCOfns/PG\nOCsxQY7jDHcwDwYDVR0TAQH/BAUwAwEB/zAaBgNVHREEEzARgglsb2NhbGhvc3SH\nBH8AAAEwDQYJKoZIhvcNAQELBQADggEBAGDYz66XgUZ+Gq9DrRkDu+z2UFihVArd\nYVBJB74zF0oIWQEIEV0iNfsFSdBEnp48fX4Q1d7bv3n87zQdCzz6l/7QN8IPYqkf\np4xTaiIH3bvhYWDJkl3xcRjz/fTQ9bb3PjO1hxv3h2ys4lENjyhtEoEFhmYax7ab\nU2+pxwGLmxDAm8GBaj8qI6S2CFfe8hvEOZHUpsuP0k09M7byW7tv/UsCeuhwIUSj\nZ1c3VvmjbPZSbUT5dRK3fgo3K+HHHjokZq2keyh7E33lodQOXJFRDiWp76RtME+P\nFwISNv335Qif7fDPhxXRjM8gBUy/boVUEy/QM5HJSkdM9FZA524B32c=\n-----END CERTIFICATE-----\n"
	p, _ := pem.Decode([]byte(st))
	if p == nil {
		panic("no PEM data found")
	}
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		panic(err)
	}

	fmt.Println("NotBefore:", cert.NotBefore)
	fmt.Println("NotAfter :", cert.NotAfter)

}
