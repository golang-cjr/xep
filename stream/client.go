package stream

import (
	"bytes"
	"encoding/xml"
	"errors"
	"log"
	"net"
	"strings"
)

type Client struct {
	User, Server, Resource, Pwd string
	conn                        net.Conn
	buffer                      []byte
}

func resolve(server string) (addr, port string) {
	return server, "5222"
}

func (c *Client) Bare() string {
	return c.User + "@" + c.Server
}

func write(c net.Conn, b []byte) (int, error) {
	log.Println(string(b))
	return c.Write(b)
}

func read(c net.Conn, b []byte) (ret []byte, err error) {
	n := 0
	if n, err = c.Read(b); err == nil && n > 0 {
		ret = b[0:n]
		log.Println(string(ret))
	}
	return
}

func (c *Client) Status() {
	data, _ := xml.Marshal(&presence)
	write(c.conn, data)
}

func (c *Client) start2(next func() error) (err error) {
	s := &Stream{to: c.Server, from: c.Bare()}
	buf := bytes.NewBuffer([]byte(xml.Header))
	data, _ := xml.Marshal(s)
	pre := string(data)
	pre = strings.TrimSuffix(pre, "</stream:stream>")
	buf.Write([]byte(pre))
	write(c.conn, buf.Bytes())
	if data, err = read(c.conn, c.buffer); err == nil {
		post := string(data) + "</stream:stream>"
		data = []byte(post)
		xml.Unmarshal(data, s)
		read(c.conn, c.buffer) //blabla features
		b := &Bind{}
		*b = bind
		b.Init(c.Resource)
		biq := &Iq{}
		*biq = iq
		biq.Id = "0001"
		biq.Type = SET
		biq.Inner = b
		data, _ = xml.Marshal(biq)
		write(c.conn, data)
		if data, err = read(c.conn, c.buffer); err == nil {
			xml.Unmarshal(data, biq)
			log.Println(b.Jid)
			if b.Jid == "" {
				err = errors.New("resource not binded")
				return
			}
			siq := &Iq{}
			*siq = iq
			siq.Id = "0002"
			siq.Type = SET
			siq.Inner = &session
			data, _ = xml.Marshal(siq)
			write(c.conn, data)
			if data, err = read(c.conn, c.buffer); err == nil {
				xml.Unmarshal(data, siq)
				next()
			}
		}
	}
	return
}

func (c *Client) Start(next func() error) (err error) {
	s := &Stream{to: c.Server, from: c.Bare()}
	buf := bytes.NewBuffer([]byte(xml.Header))
	data, _ := xml.Marshal(s)
	pre := string(data)
	pre = strings.TrimSuffix(pre, "</stream:stream>")
	buf.Write([]byte(pre))
	write(c.conn, buf.Bytes())
	if data, err = read(c.conn, c.buffer); err == nil {
		post := string(data) + "</stream:stream>"
		data = []byte(post)
		xml.Unmarshal(data, s)
		s.features = &Features{}
		if data, err = read(c.conn, c.buffer); err == nil {
			xml.Unmarshal(data, s.features)
			if s.HasMechanism("PLAIN") {
				a := &PlainAuth{}
				*a = plainAuth
				a.Init(c.User, c.Pwd)
				data, _ = xml.Marshal(a)
				write(c.conn, data)
				if data, err = read(c.conn, c.buffer); err == nil {
					xml.Unmarshal(data, a)
					if a.Success {
						log.Println("success")
						return c.start2(next)
					} else {
						err = errors.New("failed")
					}
				}
			}
		}
	}
	return
}

func (c *Client) Do(s Stanza) (err error) {
	if data, e := xml.Marshal(s); e == nil {
		_, err = write(c.conn, data)
	} else {
		err = e
	}
	return
}

func (c *Client) Process() (err error) {
	read(c.conn, c.buffer)
	return
}

func (c *Client) Stop() (err error) {
	write(c.conn, []byte("</stream:stream>"))
	read(c.conn, c.buffer)
	return
}

func Dial(cli *Client) (ret *Client, err error) {
	host, port := resolve(cli.Server)
	if cli.conn, err = net.Dial("tcp", host+":"+port); err == nil {
		cli.buffer = make([]byte, 4096)
		ret = cli
	}
	return
}
