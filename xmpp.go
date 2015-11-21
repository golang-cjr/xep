package main

import (
	"github.com/kpmy/xippo/c2s/actors"
	"github.com/kpmy/xippo/c2s/actors/steps"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"github.com/kpmy/xippo/entity/dyn"
	"github.com/kpmy/xippo/units"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func doReply(sender string, typ entity.MessageType, body string) func(stream.Stream) error {
	return func(s stream.Stream) error {
		m := entity.MSG(typ)
		if typ != entity.GROUPCHAT {
			m.To = units.Bare2Full(ROOM, sender)
		} else {
			m.To = ROOM
		}
		m.Body = body
		return s.Write(entity.Encode(dyn.NewMessage(m.Type, m.To, m.Body)))
	}
}

func xmpp(wg *sync.WaitGroup) {
	s := &units.Server{Name: server}
	c := &units.Client{Name: user, Server: s}
	var redial func(error)

	dial := func(st stream.Stream) {
		log.Println("dialing ", s)

		if err := stream.Dial(st); err == nil {
			log.Println("dialed")
			neg := &steps.Negotiation{}
			actors.With().Do(actors.C(steps.Starter), redial).Do(actors.C(neg.Act()), redial).Run(st)
			if neg.HasMechanism("PLAIN") {
				auth := &steps.PlainAuth{Client: c, Pwd: pwd}
				neg := &steps.Negotiation{}
				bind := &steps.Bind{Rsrc: resource + strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(500))}
				actors.With().Do(actors.C(auth.Act()), redial).Do(actors.C(steps.Starter)).Do(actors.C(neg.Act())).Do(actors.C(bind.Act())).Do(actors.C(steps.Session)).Do(actors.C(steps.InitialPresence)).Run(st)

				actors.With().Do(actors.C(bot)).Run(st)
			}
			wg.Done()
		}
	}

	redial = func(err error) {
		if err != nil {
			log.Println(err)
		}
		<-time.After(time.Second)
		dial(stream.New(s, redial))
	}

	redial(nil)
}
