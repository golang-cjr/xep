package entity

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type Stream struct {
	XMLName xml.Name
	Xmlns   string `xml:"xmlns,attr,omitempty"`
	Version string `xml:"version,attr,omitempty"`
	Stream  string `xml:"xmlns:stream,attr,omitempty"`
	To      string `xml:"to,attr,omitempty"`
	Id      string `xml:"-"`
	dumbProducer
}

var stream Stream = Stream{XMLName: xml.Name{Local: "stream:stream"}, Version: "1.0", Xmlns: "jabber:client", Stream: "http://etherx.jabber.org/streams"}

func (s *Stream) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	s.Id = getAttr(&start, "id")
	var _t xml.Token
	for stop := false; err == nil && !stop; {
		if _t, err = d.Token(); err == nil {
			switch t := _t.(type) {
			case xml.EndElement:
				stop = (start.Name == t.Name)
			default:
				err = errors.New(fmt.Sprint("unhandled ", t))
			}
		}
	}
	return
}
