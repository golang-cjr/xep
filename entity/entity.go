package entity

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/kpmy/ypk/halt"
	"io"
	"reflect"
	"strings"
	"xep/tools/dom"
	"xep/units"
)

func ProduceStatic(i interface{}) *bytes.Buffer {
	if data, err := xml.Marshal(i); err == nil {
		return bytes.NewBuffer(data)
	} else {
		panic(err)
	}
}

func ConsumeStatic(b *bytes.Buffer) (ret Entity, err error) {
	d := &dumb{buf: b}
	if err = xml.NewDecoder(d.buf).Decode(d); err == nil {
		ret = d.x
	}
	return
}

func Decode(b *bytes.Buffer) (ret Entity, err error) {
	dd := &domm{buf: b}
	if err = dd.Unmarshal(); err == nil {
		ret = dd
	}
	return
}

func Encode(el dom.Element) *bytes.Buffer {
	dd := &domm{model: el}
	return dd.Produce()
}

type Entity interface {
	Model() dom.Element
	Produce() *bytes.Buffer
}

type domm struct {
	buf   *bytes.Buffer
	model dom.Element
}

func (x *domm) Model() dom.Element {
	return x.model
}

func (x *domm) Produce() (ret *bytes.Buffer) {
	if data, err := xml.Marshal(x); err == nil {
		ret = bytes.NewBuffer(data)
	} else {
		halt.As(100, ret)
	}
	return
}

func (x *domm) Unmarshal() (err error) {
	d := xml.NewDecoder(x.buf)
	var _t xml.Token
	var this dom.Element
	for stop := false; !stop && err == nil; {
		if _t, err = d.RawToken(); err == nil {
			switch t := _t.(type) {
			case xml.StartElement:
				el := dom.Elem(dom.ThisName(t.Name))
				if x.model == nil {
					x.model = el
					this = el
				} else {
					this.AppendChild(el)
					this = el
				}
				for _, a := range t.Attr {
					this.Attr(dom.ThisName(a.Name), a.Value)
				}
			case xml.CharData:
				if this != nil {
					this.AppendChild(dom.Txt(string(t)))
				} else {
					stop = true
				}
			case xml.EndElement:
				if this != nil {
					if p := this.Parent(); p != nil {
						this = p.(dom.Element)
					} else {
						stop = true
					}
				} else {
					stop = true
				}
			case nil:
			default:
				halt.As(100, reflect.TypeOf(t))
			}
		}
	}
	return
}

func (x *domm) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	start.Name.Local = x.model.Name()
	for k, v := range x.model.AttrAsMap() {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: k}, Value: v})
	}
	e.EncodeToken(start)
	for _, _l := range x.model.Children() {
		switch l := _l.(type) {
		case dom.Element:
			child := &domm{}
			child.model = l
			e.Encode(child)
		case dom.Text:
			e.EncodeToken(xml.CharData(l.Data()))
		default:
			halt.As(100, reflect.TypeOf(l))
		}
	}
	e.EncodeToken(start.End())
	return
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

func (d *dumb) Model() dom.Element {
	panic(126)
}

func Open(s *units.Server) Entity {
	buf := bytes.NewBuffer([]byte(xml.Header))
	st := stream
	st.To = s.Name
	data := ProduceStatic(st)
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
