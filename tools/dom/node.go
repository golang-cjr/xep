package dom

import (
	"encoding/xml"
	"fmt"
	"github.com/kpmy/ypk/assert"
)

type (
	Leaf interface {
		Parent(...Node) Node
	}

	Text interface {
		Leaf
		Data() string
	}

	Node interface {
		Leaf
		AppendChild(Leaf)
		Children() []Leaf
	}

	Element interface {
		Attr(name string, set ...string) (get string)
		AttrAsMap() map[string]string
		Node
		Name() string
	}

	elem struct {
		a map[string]string
		l []Leaf
		n string
		p Node
	}

	txt struct {
		d string
		p Node
	}
)

func (e *elem) String() string {
	return fmt.Sprint("<", e.n, " ", e.a, ">", e.l, "</", e.n, ">")
}

func (e *elem) AppendChild(l Leaf) {
	e.l = append(e.l, l)
	l.Parent(e)
}

func (e *elem) Attr(name string, set ...string) (get string) {
	assert.For(name != "", 20)
	if e.a == nil {
		e.a = make(map[string]string)
	}
	if len(set) > 0 {
		e.a[name] = set[0]
	}
	return e.a[name]
}

func (e *elem) Name() string { return e.n }

func (e *elem) AttrAsMap() map[string]string { return e.a }

func (e *elem) Children() []Leaf { return e.l }

func (e *elem) Parent(p ...Node) Node {
	if len(p) > 0 {
		e.p = p[0]
	}
	return e.p
}

func Elem(name string) Element {
	return &elem{n: name}
}

func (t *txt) String() string { return t.d }
func (t *txt) Data() string   { return t.d }

func (t *txt) Parent(p ...Node) Node {
	if len(p) > 0 {
		t.p = p[0]
	}
	return t.p
}

func Txt(data string) Text {
	return &txt{d: data}
}

func ThisName(n xml.Name) string {
	if n.Space != "" {
		return n.Space + ":" + n.Local
	} else {
		return n.Local
	}
}
