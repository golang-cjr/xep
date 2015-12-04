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
		<script src="https://cdn.rawgit.com/gregjacobs/Autolinker.js/master/dist/Autolinker.min.js"></script>
		<script src="/static/js/log.link.js"></script>
	</head>
	<body>
		<a id="today">golang@c.j.r.</a> <a href="/stat">стата</a>
		<h1>лог</h1>
		{{range .Posts}}<p class="message"><span class="user"><em>{{.Nick}}</em></span>: <span class="content">{{.Msg}}</span></p>{{else}}ничего ._.{{end}}
		<script>
			$(function(){
				$(".content").each(function(i, e){
					var content = $(e).text();
					$(e).empty();
					$(e).html(Autolinker.link(content, {
						newWindow: true,
						stripPrefix: true
					}));
				})
			})
		</script>
	</body>
</html>
