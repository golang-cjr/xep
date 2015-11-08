os.exit = nil

chat.addEventHandler("message", "пщ", function(evt)
   if evt.body == "пщ" then chat.send("пщ") end
end)

chat.addEventHandler("message", "зига", function(evt)
   if evt.body == "o/" then chat.send("\\o,") end
end)
