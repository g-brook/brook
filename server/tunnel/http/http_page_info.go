package http

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

func Get404Info() []byte {
	return []byte(html)
}
