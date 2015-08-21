package stream

import (
	"encoding/xml"
	"fmt"
	"reflect"
)

const (
	SET = "set"
)

type Stanza interface {
	Prepare()
}

type Iq struct {
	XMLName xml.Name
	Id      string      `xml:"id,attr"`
	Type    string      `xml:"type,attr"`
	Inner   interface{} `xml:"iq"`
}

func (i *Iq) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var _t xml.Token
	for stop := false; !stop && err == nil; {
		_t, err = d.Token()
		switch t := _t.(type) {
		case xml.StartElement:
			d.DecodeElement(i.Inner, &t)
		case xml.EndElement:
			stop = t.Name == start.Name
		default:
			panic(fmt.Sprintln(reflect.TypeOf(t)))
		}
	}
	return
}

func (i *Iq) Prepare() {
	*i = iq
}

var iq = Iq{XMLName: xml.Name{Local: "iq"}}

type Presence struct {
	XMLName xml.Name
	To      string `xml:"to,attr,omitempty"`
}

func (p *Presence) Prepare() {
	*p = presence
}

var presence = Presence{XMLName: xml.Name{Local: "presence"}}
