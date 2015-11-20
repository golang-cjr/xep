package main

import (
	"flag"
	"github.com/ivpusic/golog"
	"github.com/kpmy/xep/hookexecutor"
	//"github.com/skratchdot/open-golang/open"
	"github.com/kpmy/xep/muc"
	"github.com/kpmy/xippo/c2s/actors"
	"github.com/kpmy/xippo/c2s/actors/steps"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"github.com/kpmy/xippo/entity/dyn"
	"github.com/kpmy/xippo/units"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ROOM = "golang@conference.jabber.ru"
	ME   = "xep"
)

var (
	user     string
	pwd      string
	server   string
	resource string
	neo_log  = golog.GetLogger("application")
)

var hookExec *hookexecutor.Executor

func init() {
	flag.StringVar(&user, "u", "goxep", "-u=user")
	flag.StringVar(&server, "s", "xmpp.ru", "-s=server")
	flag.StringVar(&resource, "r", "go", "-r=resource")
	flag.StringVar(&pwd, "p", "GogogOg0", "-p=password")
	log.SetFlags(0)
	posts = new(Posts)
}

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

func bot(st stream.Stream) error {
	actors.With().Do(actors.C(steps.PresenceTo(units.Bare2Full(ROOM, ME), entity.CHAT, "http:/d.ocsf.in/stat | https://github.com/kpmy/xep"))).Run(st)
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
						posts.Lock()
						if sender != ME {
							IncStat(user)
							IncStatLen(user, e.Body)
						}
						posts.data = append(posts.data, Post{Nick: sender, User: user, Msg: e.Body})
						posts.Unlock()
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

func main() {
	flag.Parse()
	s := &units.Server{Name: server}
	c := &units.Client{Name: user, Server: s}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
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
					actors.With().Do(actors.C(auth.Act()), redial).Do(actors.C(steps.Starter)).Do(actors.C(neg.Act())).Do(actors.C(bind.Act())).Do(actors.C(steps.Session)).Run(st)
					actors.With().Do(actors.C(steps.InitialPresence)).Run(st)
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
	}()
	go neo_server(wg)
	go func() {
		time.Sleep(time.Duration(time.Millisecond * 200))
		//open.Start("http://localhost:3000")
		//open.Start("http://localhost:3000/stat")
	}()
	wg.Wait()
}
