<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8"/>
		<title>лог :: golang@c.j.r</title>
		<link rel="stylesheet" href="/static/css/default.css"/>
	</head>
	<body>
		<a href="/stat">стата</a>
		<h1>лог</h1>
		{{range .Posts}}<em>{{.Nick}}</em>: {{.Msg}}<br/>{{else}}ничего ._.{{end}}
	</body>
</html>
