package steps

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"xep/c2s/stream"
	"xep/entity"
)

func Starter(s stream.Stream) (err error) {
	if err = s.Write(entity.Open(s.Server()).Produce()); err == nil {
		s.Ring(func(b *bytes.Buffer) (done bool) {
			_e, e := entity.ConsumeStatic(b)
			if _e != nil {
				switch e := _e.(type) {
				case *entity.Stream:
					s.Id(e.Id)
					done = true
				default:
					log.Println(reflect.TypeOf(e))
				}
			} else if e == nil {
				err = errors.New(fmt.Sprint("unknown entity ", string(b.Bytes())))
			} else {
				err = e
			}
			return
		}, 0)
	}
	return
}

type Negotiation struct {
	AuthMechanisms []string
}

func (n *Negotiation) Act() func(stream.Stream) error {
	return func(s stream.Stream) (err error) {
		s.Ring(func(b *bytes.Buffer) (done bool) {
			var _e entity.Entity
			if _e, err = entity.ConsumeStatic(b); err == nil {
				switch e := _e.(type) {
				case *entity.Features:
					n.AuthMechanisms = e.Mechanisms
					done = true
				default:
					log.Println(reflect.TypeOf(e))
				}
			}
			return
		}, 0)
		return
	}
}

func (n *Negotiation) HasMechanism(mech string) (ok bool) {
	if n.AuthMechanisms != nil {
		for _, v := range n.AuthMechanisms {
			if strings.ToLower(v) == strings.ToLower(mech) {
				ok = true
				break
			}
		}
	}
	return
}
