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
		<script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
	</head>
	<body>
		<a id="today">golang@c.j.r. сегодня</a> <a href="/stat">стата</a>
		<h1>лог</h1>
		{{range .Posts}}<p class="message"><span class="user"><em>{{.Nick}}</em></span>: {{.Msg}}</p>{{else}}ничего ._.{{end}}
		<script>
			$(function(){
				var d = new Date();
				$("#today").attr("href", "http://chatlogs.jabber.ru/golang@conference.jabber.ru/"+d.getFullYear()+"/"+(d.getMonth() + 1)+"/"+d.getDate()+".html");
			})
		</script>
	</body>
</html>
