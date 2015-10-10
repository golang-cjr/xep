<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8"/>
		<title>стата :: golang@c.j.r</title>
		<link rel="stylesheet" href="/static/css/default.css"/>

		<script src="http://code.highcharts.com/adapters/standalone-framework.js"></script>
		<script src="http://code.highcharts.com/highcharts.js"></script>
	</head>
	<body>
		<a href="/">логи</a>
		<div class="clearfix"></div>
		<div class="container">
			<div class="col" id="col-userlist">
				<h1>стата</h1>
				<p><em>всего сообщений</em>: {{.Total}}</p>
				<table>
					{{range .Stat}}
					<tr>
						<td><em>{{.User}}</em></td>
						<td>{{.Count}}</td>
						<td>{{printf "%.2f" .Perc}}%</td>
					</tr>
					{{else}}<trd><tr>ничего ._.</tr></td>{{end}}
				</table>
			</div>
			<div id="chart-container" class="col">
			</div>
		</div>
	</body>
	<script type="text/javascript">
		var data = [];
		{{range .Stat}}
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
				text: "Стата"
			},
			series: [{
				name: "Пиздливость",
				data: data
				}]
			});
	</script>
</html>
