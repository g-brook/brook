package utils

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
</head>
<body>
<div style="text-align: center;">
<h1 style="color: red;">404</h1>
<p>The requested route path could not be found.</p>
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
func GetResponse(status int) http.Response {
	body := GetServerUnreachable()
	header := make(http.Header)
	header.Set("Content-Type", "text/html; charset=utf-8")
	return http.Response{
		StatusCode:    status,
		Body:          io.NopCloser(bytes.NewReader(body)),
		Proto:         "HTTP/1.1",
		Header:        header,
		ContentLength: int64(len(body)),
	}
}

// GetPageNotFound returns the 404 error page content
func GetPageNotFound() []byte {
	return htmlBytes
}
