package steps

import (
	"bytes"
	"log"
	"reflect"
	"xep/c2s/stream"
	"xep/entity"
)

type Bind struct {
	Rsrc string
}

func (b *Bind) Act() func(s stream.Stream) error {
	return func(s stream.Stream) (err error) {
		bind := &entity.Bind{}
		*bind = entity.BindPrototype
		bind.Resource = b.Rsrc
		iq := entity.IQ(entity.SET, bind)
		if err = s.Write(entity.ProduceStatic(iq)); err == nil {
			s.Ring(func(b *bytes.Buffer) (done bool) {
				var _e entity.Entity
				if _e, err = entity.ConsumeStatic(b); err == nil {
					switch e := _e.(type) {
					case *entity.Iq:
						switch {
						case e.Id == iq.Id && e.Type == entity.RESULT:
							stream.Bind(s, e.Inner.(*entity.Bind).Jid)
							done = true
						}
					default:
						log.Println(reflect.TypeOf(e))
					}
				}
				return
			}, 0)
		}
		return
	}
}

func Session(s stream.Stream) (err error) {
	iq := entity.IQ(entity.SET, &entity.SessionPrototype)
	if err = s.Write(entity.ProduceStatic(iq)); err == nil {
		s.Ring(func(b *bytes.Buffer) (done bool) {
			var _e entity.Entity
			if _e, err = entity.ConsumeStatic(b); err == nil {
				switch e := _e.(type) {
				case *entity.Iq:
					switch {
					case e.Id == iq.Id && e.Type == entity.RESULT:
						done = true
					}
				default:
					log.Println(reflect.TypeOf(e))
				}
			}
			return
		}, 0)
	}
	return
}
