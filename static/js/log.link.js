/**
 * Created by petry_000 on 05.12.2015.
 */
$(function(){
    var d = new Date();
    var day = d.getDate() <= 9 ? "0" + d.getDate() : d.getDate();
    $("#today").attr("href", "http://chatlogs.jabber.ru/golang@conference.jabber.ru/"+d.getFullYear()+"/"+(d.getMonth() + 1)+"/"+day+".html");
});