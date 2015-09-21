package entity

import (
	"bytes"
	"encoding/xml"
	"strings"
	"xep/tools/dom"
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

type dumbProducer struct{}

func (d *dumbProducer) Produce() *bytes.Buffer { panic(126) }

func (d *dumbProducer) Model() dom.Element {
	panic(126)
}
