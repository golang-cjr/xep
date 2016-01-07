/**
 * Created by petry_000 on 05.12.2015.
 */
$(function(){
    var d = new Date();
	var month = d.getMonth() < 9 ? "0" + (d.getMonth() + 1) : d.getMonth() + 1;
    var day = d.getDate() <= 9 ? "0" + d.getDate() : d.getDate();
    console.log(month);
    $("#today").attr("href", "http://chatlogs.jabber.ru/golang@conference.jabber.ru/"+d.getFullYear()+"/"+month+"/"+day+".html");
});