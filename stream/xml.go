package stream

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"strings"
)

func setAttr(start *xml.StartElement, name, value string) {
	a := xml.Attr{}
	a.Name.Local = name
	a.Value = value
	start.Attr = append(start.Attr, a)
}

func getAttr(start *xml.StartElement, name string) (value string) {
	for _, v := range start.Attr {
		if strings.ToLower(v.Name.Local) == strings.ToLower(name) {
			value = v.Value
			break
		}
	}
	return
}

type Stream struct {
	to, from string
	id       string
	features *Features
}

func (s *Stream) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	start.Name = xml.Name{Local: "stream:stream"}
	setAttr(&start, "version", "1.0")
	setAttr(&start, "xmlns", "jabber:client")
	setAttr(&start, "xml:lang", "en")
	setAttr(&start, "xmlns:stream", "http://etherx.jabber.org/streams")
	setAttr(&start, "to", s.to)
	setAttr(&start, "from", s.from)
	e.EncodeToken(start)
	e.EncodeToken(start.End())
	return
}

func (s *Stream) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	s.id = getAttr(&start, "id")
	return
}

func (s *Stream) HasMechanism(mech string) (ok bool) {
	if s.features != nil {
		for _, v := range s.features.Mechanisms {
			if strings.ToLower(v) == strings.ToLower(mech) {
				ok = true
				break
			}
		}
	}
	return
}

type Features struct {
	Mechanisms []string `xml:"mechanisms>mechanism"`
}

type PlainAuth struct {
	XMLName   xml.Name
	Xmlns     string `xml:"xmlns,attr"`
	Mechanism string `xml:"mechanism,attr"`
	Data      string `xml:",chardata"`
	Success   bool
}

func (p *PlainAuth) Init(user, pwd string) {
	data := bytes.NewBuffer([]byte(user + "@" + pwd))
	data.WriteByte(0)
	data.Write([]byte(user))
	data.WriteByte(0)
	data.Write([]byte(pwd))
	p.Data = base64.StdEncoding.EncodeToString(data.Bytes())
}

func (p *PlainAuth) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	p.Success = start.Name.Local == "success"
	return
}

var plainAuth = PlainAuth{XMLName: xml.Name{Local: "auth"}, Xmlns: "urn:ietf:params:xml:ns:xmpp-sasl", Mechanism: "PLAIN"}

type Bind struct {
	XMLName  xml.Name
	Xmlns    string `xml:"xmlns,attr"`
	Resource string `xml:"resource"`
	Jid      string `xml:"jid"`
}

func (b *Bind) Init(rsrc string) {
	b.Resource = rsrc
}

var bind = Bind{XMLName: xml.Name{Local: "bind"}, Xmlns: "urn:ietf:params:xml:ns:xmpp-bind"}

type Session struct {
	XMLName xml.Name
	Xmlns   string `xml:"xmlns,attr"`
}

var session = Session{XMLName: xml.Name{Local: "session"}, Xmlns: "urn:ietf:params:xml:ns:xmpp-session"}
