package entity

import (
	"encoding/xml"
)

type Bind struct {
	XMLName  xml.Name
	Xmlns    string `xml:"xmlns,attr"`
	Resource string `xml:"resource"`
	Jid      string `xml:"jid"`
}

func (b *Bind) Init(rsrc string) {
	b.Resource = rsrc
}

var BindPrototype = Bind{XMLName: xml.Name{Local: "bind"}, Xmlns: "urn:ietf:params:xml:ns:xmpp-bind"}

type Session struct {
	XMLName xml.Name
	Xmlns   string `xml:"xmlns,attr"`
}

var SessionPrototype = Session{XMLName: xml.Name{Local: "session"}, Xmlns: "urn:ietf:params:xml:ns:xmpp-session"}
