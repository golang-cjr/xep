<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8"/>
		<title>лог :: golang@c.j.r</title>
		<link rel="stylesheet" href="/static/css/default.css"/>
		<style>
			.message {
				white-space: pre-wrap
			}
			.user{
				color: grey
			}
		</style>
	</head>
	<body>
		<a href="/stat">стата</a>
		<h1>лог</h1>
		{{range .Posts}}<p class="message"><span class="user"><em>{{.Nick}}</em></span>: {{.Msg}}</p>{{else}}ничего ._.{{end}}
	</body>
</html>
