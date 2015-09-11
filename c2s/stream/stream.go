package stream

import (
	"bytes"
	"errors"
	"github.com/kpmy/ypk/halt"
	"hash/adler32"
	"net"
	"reflect"
	"time"
	"xep/tools/srv"
	"xep/units"
)

type Stream interface {
	Server() *units.Server
	Write(*bytes.Buffer) error
	Ring(func(*bytes.Buffer) bool, time.Duration)
	Id(...string) string
}

type wrapperStream struct {
	base Stream
}

func (w *wrapperStream) Write(b *bytes.Buffer) error { return w.base.Write(b) }

func (w *wrapperStream) Ring(fn func(*bytes.Buffer) bool, t time.Duration) {
	w.base.Ring(fn, t)
}

func (w *wrapperStream) Server() *units.Server { return w.base.Server() }

func (w *wrapperStream) Id(s ...string) string { return w.base.Id(s...) }

type dummyStream struct {
	to *units.Server
}

func (d *dummyStream) Ring(func(*bytes.Buffer) bool, time.Duration) { panic(126) }
func (d *dummyStream) Write(b *bytes.Buffer) error                  { panic(126) }
func (d *dummyStream) Server() *units.Server                        { return d.to }
func (d *dummyStream) Id(...string) string                          { return "" }

type xmppStream struct {
	to   *units.Server
	conn net.Conn
	ctrl chan bool
	data chan pack
	id   string
	jid  string
}

type pack struct {
	data []byte
	hash uint32
}

func (x *xmppStream) Id(s ...string) string {
	if len(s) > 0 {
		x.id = s[0]
	}
	return x.id
}
func (x *xmppStream) Server() *units.Server { return x.to }

func (x *xmppStream) Write(b *bytes.Buffer) (err error) {
	_, err = x.conn.Write(b.Bytes())
	return
}

func (x *xmppStream) Ring(fn func(*bytes.Buffer) bool, timeout time.Duration) {
	timed := make(chan bool)
	if timeout > 0 {
		go func() {
			<-time.NewTimer(timeout).C
			timed <- true
		}()
	}
	for stop := false; !stop; {
		select {
		case p := <-x.data:
			done := fn(bytes.NewBuffer(p.data))
			if !done {
				x.data <- pack{data: p.data, hash: p.hash}
			} else {
				stop = true
			}
		case <-timed:
			stop = true
		}
	}
}

func New(to *units.Server) Stream {
	return &wrapperStream{base: &dummyStream{to: to}}
}

func Bind(_s Stream, jid string) {
	switch w := _s.(type) {
	case *wrapperStream:
		switch s := w.base.(type) {
		case *xmppStream:
			s.jid = jid
		}
	}
}

func Dial(_s Stream) (err error) {
	switch w := _s.(type) {
	case *wrapperStream:
		switch s := w.base.(type) {
		case *dummyStream:
			x := &xmppStream{to: s.to}
			var (
				host, port string
			)
			if host, port, err = srv.Resolve(x.to); err == nil {
				if x.conn, err = net.Dial("tcp", host+":"+port); err == nil {
					x.ctrl = make(chan bool)
					x.data = make(chan pack, 256)
					go func(stream *xmppStream) {
						<-stream.ctrl
						stream.conn.Close()
					}(x)
					go func(stream *xmppStream) {
						var err error
						buf := make([]byte, 65535)
						for err == nil {
							n := 0
							n, err = stream.conn.Read(buf)
							if n > 0 && err == nil {
								data := make([]byte, n)
								copy(data, buf)
								for data := range spl1t(data) {
									stream.data <- pack{data: data, hash: adler32.Checksum(data)}
								}
							}
						}
					}(x)
					w.base = x
				}
			}
		default:
			err = errors.New("already connected")
		}
	default:
		halt.As(100, reflect.TypeOf(_s))
	}
	return
}