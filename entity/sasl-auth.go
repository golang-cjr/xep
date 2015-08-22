package entity

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
)

type PlainAuth struct {
	XMLName   xml.Name
	Xmlns     string `xml:"xmlns,attr"`
	Mechanism string `xml:"mechanism,attr"`
	Data      string `xml:",chardata"`
}

type Success struct {
	dumbProducer
}

func (p *PlainAuth) Init(user, pwd string) {
	data := bytes.NewBuffer([]byte(user + ":" + pwd))
	data.WriteByte(0)
	data.Write([]byte(user))
	data.WriteByte(0)
	data.Write([]byte(pwd))
	p.Data = base64.StdEncoding.EncodeToString(data.Bytes())
}

var PlainAuthPrototype = PlainAuth{XMLName: xml.Name{Local: "auth"}, Xmlns: "urn:ietf:params:xml:ns:xmpp-sasl", Mechanism: "PLAIN"}
