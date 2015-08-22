package steps

import (
	"bytes"
	"log"
	"reflect"
	"xep/c2s/stream"
	"xep/entity"
	"xep/units"
)

type PlainAuth struct {
	Client *units.Client
	Pwd    string
}

func (p *PlainAuth) Act() func(stream.Stream) error {
	return func(s stream.Stream) (err error) {
		auth := &entity.PlainAuth{}
		*auth = entity.PlainAuthPrototype
		auth.Init(p.Client.Name, p.Pwd)
		if err = s.Write(entity.Produce(auth)); err == nil {
			s.Ring(func(b *bytes.Buffer) (ret *bytes.Buffer) {
				var _e entity.Entity
				if _e, err = entity.Consume(b); err == nil {
					switch e := _e.(type) {
					case *entity.Success:

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
