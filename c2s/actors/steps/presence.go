package steps

import (
	"xep/c2s/stream"
	"xep/entity"
)

func InitialPresence(s stream.Stream) (err error) {
	err = s.Write(entity.ProduceStatic(&entity.PresencePrototype))
	return
}

func PresenceTo(jid string) func(stream.Stream) error {
	return func(s stream.Stream) error {
		pr := &entity.Presence{}
		*pr = entity.PresencePrototype
		pr.To = jid
		return s.Write(entity.ProduceStatic(pr))
	}
}
