<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8"/>
		<title>стата :: golang@c.j.r</title>
		<link rel="stylesheet" href="/static/css/default.css"/>

		<script src="http://code.highcharts.com/adapters/standalone-framework.js"></script>
		<script src="http://code.highcharts.com/highcharts.js"></script>
		<script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
	</head>
	<body>
		<a id="today">golang@c.j.r. сегодня</a> <a href="/">логи</a>
		<div class="clearfix"></div>
		<div class="container">
			<div class="col" id="col-userlist">
				<h1>стата</h1>
				<p><strong>Всего</strong> (сообщений/символов): {{.Total.Total}}/{{.Count.Total}}</p>
				<table>
					<tr>
						<td><strong>Сообщения</strong></td>
					</tr>
					{{range .Total.Stat}}
					<tr>
						<td><em>{{.User}}</em></td>
						<td>{{.Count}}</td>
						<td>{{printf "%.2f" .Perc}}%</td>
					</tr>
					{{else}}<trd><tr>ничего ._.</tr></td>{{end}}
					<tr><td/></tr>
					<tr>
						<td><strong>Символы</strong></td>
					</tr>
					{{range .Count.Stat}}
					<tr>
						<td><em>{{.User}}</em></td>
						<td>{{.Count}}</td>
						<td>{{printf "%.2f" .Perc}}%</td>
					</tr>
					{{end}}
				</table>
			</div>
			<div id="chart-container" class="col">
			</div>
			<div id="deads-container">
				<h2>трупы</h2>
				<img src="/static/img/deads.png" /><br/>
				{{range .Deads}}
					<span><em>{{.Nick}}</em></span><br/>
				{{end}}
			</div>
		</div>
	</body>
	<script type="text/javascript">
		var data = [];
		{{range .Total.Stat}}
			data.push({
				name: "{{.User}}",
				y: parseFloat({{printf "%.2f" .Perc}})
				});
		{{end}}
		var item = document.querySelector("#chart-container");
		item.style.height = document.querySelector("#col-userlist").clientHeight + "px";
		var chart = new Highcharts.Chart({
			chart: {
				renderTo: 'chart-container',
				type: 'pie'
			},
			title: {
				text: "Сообщения"
			},
			series: [{
				name: "Пиздливость",
				data: data
				}]
			});
	</script>
	<script>
		$(function(){
				var d = new Date();
				$("#today").attr("href", "http://chatlogs.jabber.ru/golang@conference.jabber.ru/"+d.getFullYear()+"/"+(d.getMonth() + 1)+"/"+d.getDate()+".html");
			})
	</script>
</html>
