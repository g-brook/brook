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

package httpx

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

var html = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>404 - Not Found</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial;
            background: #f4f6f8;
            height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            text-align: center;
            background: white;
            padding: 40px 60px;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        h1 {
            font-size: 72px;
            margin: 0;
            background: linear-gradient(45deg, #ff6b6b, #ff8787);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        p {
            color: #495057;
            font-size: 18px;
            margin: 20px 0 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>404</h1>
        <p>The requested route path could not be found.</p>
		<a href="https://github.com/g-brook/brook">https://github.com/g-brook/brook</a>
    </div>
</body>
</html>`

var (
	htmlOnce  sync.Once
	htmlBytes []byte
)

func init() {
	htmlOnce.Do(func() {
		htmlBytes = []byte(html)
	})
}

var serverUnreachable = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Server Unreachable</title>
</head>
<body>
<div style="text-align: center;">
<h1 style="color: red;">Server Bad Gateway(502)</h1>
<p>The requested server could not be reached.</p>
</div>
</body>
</html>`

var (
	serverUnreachableOnce  sync.Once
	serverUnreachableBytes []byte
)

func init() {
	serverUnreachableOnce.Do(func() {
		serverUnreachableBytes = []byte(serverUnreachable)
	})
}

// GetServerUnreachable returns the server unreachable error page content
func GetServerUnreachable() []byte {
	return serverUnreachableBytes
}

// GetResponse returns the server unreachable error page response
func GetResponse(status int) *http.Response {
	body := GetServerUnreachable()
	header := make(http.Header)
	header.Set("Content-Type", "text/html; charset=utf-8")
	return &http.Response{
		StatusCode:    status,
		Body:          io.NopCloser(bytes.NewReader(body)),
		Proto:         "HTTP/1.1",
		Header:        header,
		ContentLength: int64(len(body)),
	}
}

// GetPageNotFound returns the 404 error page content
func GetPageNotFound(state int) []byte {
	return htmlBytes
}
