[![Build Status](https://drone.io/github.com/kpmy/xep/status.png)](https://drone.io/github.com/kpmy/xep/latest)

# xep

[golang@c.j.r](https://github.com/golang-cjr) [c](http://d.ocsf.in)[h](http://d.ocsf.in/stat)[a](http://chatlogs.jabber.ru/golang@conference.jabber.ru/)[t](https://github.com/golang-cjr)-[b](http://d.ocsf.in)[o](http://d.ocsf.in/stat)[t](http://chatlogs.jabber.ru/golang@conference.jabber.ru/)

Bot provides a remote chat API, for sake of microservice flamewar.

The API is dead-simple. There are a socket you connect to and start receiving
messages. You may send messages back in same format - they will be said by bot.
When you recieve message with Type: "ping" please reply with message with Type:
"pong" or your client will be disconnected.

Message schema:

Type: string -- type of message. Currently contains "message" or "ping" or "pong" or "ack"
ID: int -- serial id of message
Error: string -- contains error message when some error occured
Data: map[string]string -- message payload
    body: string -- message body
    sender: string -- nickname of message sender

There are two supported wire protocols: binary msgpack-based and textual
json-based. They share a schema.

Binary: listen on port 1984. Every message starts with 2 byte big-endian unsigned integer, the length
of packed message. Then follows the byte blob of denotes size which contains
msgpack-encoded message.
Text: listen on port 1985. messages are encoded into compact json and delimeted with newlines.

Restrictions:

Bot limits the rate of messages (scored by lines) client can deliver to conference. There are hard
per-minute limit of lines and soft per-10-second limit. Current default
limit can be found in hookexecutor/executor.go. When client exceeds his quota
bot will reject messages. When bot rejects a message it sends it back to client
with Error field set to corresponding reason of delivery fail.
When client accepts the message, it will send the message with Type: "ack" and
copied ID to notify client that message was accepted to delivery.

Good luck writing your own client!
