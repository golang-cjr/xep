package entity

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"xep/units"
)

func Produce(i interface{}) *bytes.Buffer {
	if data, err := xml.Marshal(i); err == nil {
		return bytes.NewBuffer(data)
	} else {
		panic(err)
	}
}

func Consume(b *bytes.Buffer) (ret Entity, err error) {
	d := &dumb{buf: b}
	if err = xml.NewDecoder(d.buf).Decode(d); err == nil {
		ret = d.x
	}
	return
}

type Entity interface {
	Produce() *bytes.Buffer
}

type dumb struct {
	buf *bytes.Buffer
	x   Entity
}

func (x *dumb) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	if fact, ok := ns[start.Name]; ok {
		x.x = fact(x.buf)
		d.DecodeElement(x.x, &start)
	} else {
		err = errors.New(fmt.Sprint("unknown entity ", start.Name))
	}
	return
}

func (d *dumb) Produce() *bytes.Buffer {
	return d.buf
}

func Open(s *units.Server) Entity {
	buf := bytes.NewBuffer([]byte(xml.Header))
	st := stream
	st.To = s.Name
	data := Produce(st)
	pre := strings.TrimSuffix(string(data.Bytes()), "</stream:stream>")
	io.Copy(buf, bytes.NewBufferString(pre))
	return &dumb{buf: buf}
}

var ns map[xml.Name]func(*bytes.Buffer) Entity

func init() {
	ns = make(map[xml.Name]func(*bytes.Buffer) Entity)

	ns[xml.Name{Space: "http://etherx.jabber.org/streams", Local: "stream"}] = func(buf *bytes.Buffer) Entity {
		s := stream
		buf.WriteString("</stream:stream>")
		return &s
	}

	ns[xml.Name{Space: "stream", Local: "features"}] = func(*bytes.Buffer) Entity { return &Features{} }

	ns[xml.Name{Space: "urn:ietf:params:xml:ns:xmpp-sasl", Local: "success"}] = func(*bytes.Buffer) Entity { return &Success{} }

	ns[xml.Name{Local: "iq"}] = func(*bytes.Buffer) Entity { return &Iq{} }

	ns[xml.Name{Local: "message"}] = func(*bytes.Buffer) Entity { return &Message{} }
}
