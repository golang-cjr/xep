package entity

import (
	"encoding/xml"
	"github.com/kpmy/ypk/fn"
	"math/rand"
	"strconv"
)

type IqType string

const (
	SET    IqType = "set"
	RESULT IqType = "result"
)

type MessageType string

const (
	GROUPCHAT MessageType = "groupchat"
)

type Iq struct {
	XMLName xml.Name
	Id      string      `xml:"id,attr,omitempty"`
	Type    IqType      `xml:"type,attr"`
	Inner   interface{} `xml:"iq"`
	dumbProducer
}

func (i *Iq) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	i.Id = getAttr(&start, "id")
	i.Type = IqType(getAttr(&start, "type"))
	var _t xml.Token
	for stop := false; !stop && err == nil; {
		_t, err = d.Token()
		switch t := _t.(type) {
		case xml.StartElement:
			if fact, ok := us[t.Name]; ok {
				i.Inner = fact()
			}
			if !fn.IsNil(i.Inner) {
				d.DecodeElement(i.Inner, &t)
			} else {
				//halt.As(100, t.Name)
			}
		case xml.EndElement:
			stop = t.Name == start.Name
		default:
			//halt.As(100, reflect.TypeOf(t))
		}
	}
	return
}

var iq = Iq{XMLName: xml.Name{Local: "iq"}}

func IQ(typ IqType, user interface{}) *Iq {
	i := &Iq{}
	*i = iq
	i.Type = typ
	i.Inner = user
	i.Id = strconv.FormatInt(int64(rand.Intn(0xffffff)), 16)
	return i
}

type Presence struct {
	XMLName xml.Name
	To      string `xml:"to,attr,omitempty"`
}

var PresencePrototype = Presence{XMLName: xml.Name{Local: "presence"}}

type Message struct {
	XMLName xml.Name
	dumbProducer
	Body string      `xml:"body,omitempty"`
	From string      `xml:"from,attr,omitempty"`
	To   string      `xml:"to,attr,omitempty"`
	Type MessageType `xml:"type,attr,omitempty"`
	Id   string      `xml:"id,attr,omitempty"`
}

var MessagePrototype = Message{XMLName: xml.Name{Local: "message"}}

func MSG(typ MessageType) *Message {
	m := &Message{}
	*m = MessagePrototype
	m.Type = typ
	m.Id = strconv.FormatInt(int64(rand.Intn(0xffffff)), 16)
	return m
}

var us map[xml.Name]func() interface{}

func init() {
	us = make(map[xml.Name]func() interface{})
	us[xml.Name{Space: "urn:ietf:params:xml:ns:xmpp-bind", Local: "bind"}] = func() interface{} { return &Bind{} }
}
