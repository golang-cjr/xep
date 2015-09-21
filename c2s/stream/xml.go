package stream

import (
	"bytes"
	"encoding/xml"
	"github.com/kpmy/ypk/halt"
	"io"
	"reflect"
)

func spl1t(bunch io.Reader) (ret chan []byte) {
	ret = make(chan []byte)
	go func() {
		d := xml.NewDecoder(bunch)
		d.Strict = false
		var (
			_t  xml.Token
			err error
			buf *bytes.Buffer
			e   *xml.Encoder
		)
		init := func() {
			buf = bytes.NewBuffer(nil)
			e = xml.NewEncoder(buf)
		}
		flush := func() {
			e.Flush()
			if buf.Len() > 0 {
				ret <- buf.Bytes()
			}
			init()
		}
		join := func(n xml.Name) (ret string) {
			if n.Space != "" {
				ret = n.Space + ":"
			}
			ret = ret + n.Local
			return
		}
		depth := 0
		init()
		for stop := false; !stop && err == nil; {
			if _t, err = d.RawToken(); err == nil {
				switch t := _t.(type) {
				case xml.ProcInst:
					e.EncodeToken(t.Copy())
				case xml.StartElement:
					tt := t.Copy()
					tt.Name = xml.Name{Local: join(t.Name)}
					var tmp []xml.Attr
					for _, a := range tt.Attr {
						a.Name = xml.Name{Local: join(a.Name)}
						tmp = append(tmp, a)
					}
					tt.Attr = tmp
					e.EncodeToken(tt)
					if tt.Name.Local == "stream:stream" {
						depth--
						flush()
					}
					depth++
				case xml.EndElement:
					tt := t
					tt.Name = xml.Name{Local: join(t.Name)}
					e.EncodeToken(tt)
					depth--
					if depth == 0 {
						flush()
					}
				case xml.CharData:
					e.EncodeToken(t)
				default:
					halt.As(100, reflect.TypeOf(t))
				}
			} else {
				flush()
			}
		}
		close(ret)
	}()
	return
}
