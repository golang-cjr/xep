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
		if err = s.Write(entity.Produce(iq)); err == nil {
			s.Ring(func(b *bytes.Buffer) (ret *bytes.Buffer) {
				var _e entity.Entity
				if _e, err = entity.Consume(b); err == nil {
					switch e := _e.(type) {
					case *entity.Iq:
						switch {
						case e.Id == iq.Id && e.Type == entity.RESULT:
							stream.Bind(s, e.Inner.(*entity.Bind).Jid)
						default:
							ret = b //pass
						}
					default:
						log.Println(reflect.TypeOf(e))
						ret = b //pass
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
	if err = s.Write(entity.Produce(iq)); err == nil {
		s.Ring(func(b *bytes.Buffer) (ret *bytes.Buffer) {
			var _e entity.Entity
			if _e, err = entity.Consume(b); err == nil {
				switch e := _e.(type) {
				case *entity.Iq:
					switch {
					case e.Id == iq.Id && e.Type == entity.RESULT:
					default:
						ret = b
					}
				default:
					log.Println(reflect.TypeOf(e))
					ret = b //pass
				}
			}
			return
		}, 0)
	}
	return
}
