package main

import (
	"log"
	"net"
	"testing"
)

func TestConnection(t *testing.T){
	if c, err:=net.Dial("tcp", "jabber.ru:5222"); err==nil{
		c.Write([]byte(`<?xml version='1.0'?>
   <stream:stream
       from='uni0n@xmpp.ru'
       to='xmpp.ru'
       version='1.0'
       xml:lang='en'
       xmlns='jabber:client'
       xmlns:stream='http://etherx.jabber.org/streams'>`))
		buf:=make([]byte, 256)
		c.Read(buf);
		log.Println(string(buf))
		c.Close()
	}
}