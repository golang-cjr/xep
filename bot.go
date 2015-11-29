package main

import (
	"github.com/golang-cjr/xep/hookexecutor"
	"github.com/golang-cjr/xep/muc"
	"github.com/kpmy/xippo/c2s/actors"
	"github.com/kpmy/xippo/c2s/actors/steps"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"github.com/kpmy/xippo/entity/dyn"
	"github.com/kpmy/xippo/units"
	"log"
	"reflect"
	"strings"
)

var hookExec *hookexecutor.Executor

func bot(st stream.Stream) error {
	actors.With().Do(actors.C(steps.PresenceTo(units.Bare2Full(ROOM, ME), entity.CHAT, "http://d.ocsf.in/stat | https://github.com/golang-cjr/xep"))).Run(st)
	hookExec = hookexecutor.NewExecutor(st)
	hookExec.Start()
	for {
		st.Ring(conv(func(_e entity.Entity) {
			switch e := _e.(type) {
			case *entity.Message:
				if strings.HasPrefix(e.From, ROOM+"/") {
					sender := strings.TrimPrefix(e.From, ROOM+"/")
					um := muc.UserMapping()
					user := sender
					if u, ok := um[sender]; ok {
						user, _ = u.(string)
					}
					if e.Type == entity.GROUPCHAT {
						room.Lock()
						if sender != ME {
							IncStat(user)
							IncStatLen(user, e.Body)
							room.Active(user)
						}
						room.Grow(Post{Nick: sender, User: user, Msg: e.Body})
						room.Unlock()
					}
					if sender != ME {
						words.Lock()
						Stem(e.Body)
						words.Unlock()
					}
					if sender != ME {
						hookExec.NewEvent(hookexecutor.IncomingEvent{"message", map[string]string{"sender": sender, "body": e.Body}})
						switch {
						case strings.EqualFold(strings.TrimSpace(e.Body), "пщ"):
							go func() {
								actors.With().Do(actors.C(doReply(sender, e.Type, "пщ!"))).Run(st)
							}()
						case strings.HasPrefix(e.Body, "xep"):
							body := strings.TrimPrefix(e.Body, "xep")
							body = strings.TrimSpace(body)
							if body != "" {
								go func() {
									actors.With().Do(actors.C(doReply(sender, entity.GROUPCHAT, body))).Run(st)
								}()
							}
						}
					}
				}
			case dyn.Entity:
				switch e.Type() {
				case dyn.PRESENCE:
					if from := e.Model().Attr("from"); from != "" && strings.HasPrefix(from, ROOM+"/") {
						sender := strings.TrimPrefix(from, ROOM+"/")
						um := muc.UserMapping()
						user := sender
						if u, ok := um[sender]; ok {
							user, _ = u.(string)
						}
						actors.With().Do(R(func(r *Room) (e error) {
							r.User(user)
							return
						})).Run(room)
						if show := firstByName(e.Model(), "show"); e.Model().Attr("type") == "" && (show == nil || show.ChildrenCount() == 0) { //онлаен тип
							hookExec.NewEvent(hookexecutor.IncomingEvent{"presence", map[string]string{"sender": sender, "user": user}})
						}
					}
				}
			default:
				log.Println(reflect.TypeOf(e))
			}
		}), 0)
	}
}
