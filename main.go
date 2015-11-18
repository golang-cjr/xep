package main

import (
	"flag"
	"github.com/ivpusic/golog"
	"github.com/kpmy/xep/hookexecutor"
	"github.com/skratchdot/open-golang/open"
	"reflect"
	//	"github.com/kpmy/xep/jsexecutor"
	//	"github.com/kpmy/xep/luaexecutor"
	"github.com/kpmy/xep/muc"
	"github.com/kpmy/xippo/c2s/actors"
	"github.com/kpmy/xippo/c2s/actors/steps"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"github.com/kpmy/xippo/entity/dyn"
	"github.com/kpmy/xippo/units"
	"html/template"
	"log"
	"math/rand"
	"os"
	"path/filepath"
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

type (
	Post struct {
		User string
		Nick string
		Msg  string
	}

	Posts struct {
		data []Post
		sync.Mutex
	}
)

var posts *Posts

//var executor *luaexecutor.Executor
//var jsexec *jsexecutor.Executor
var hookExec *hookexecutor.Executor

func init() {
	flag.StringVar(&user, "u", "goxep", "-u=user")
	flag.StringVar(&server, "s", "xmpp.ru", "-s=server")
	flag.StringVar(&resource, "r", "go", "-r=resource")
	flag.StringVar(&pwd, "p", "GogogOg0", "-p=password")
	log.SetFlags(0)
	posts = new(Posts)
}

func (d *StatData) Len() int { return len(d.Stat) }

func (d *StatData) Less(i, j int) bool { return d.Stat[i].Count > d.Stat[j].Count }

func (d *StatData) Swap(i, j int) { d.Stat[i], d.Stat[j] = d.Stat[j], d.Stat[i] }

func doReply(sender string, typ entity.MessageType) func(stream.Stream) error {
	return func(s stream.Stream) error {
		m := entity.MSG(typ)
		if typ != entity.GROUPCHAT {
			m.To = units.Bare2Full(ROOM, sender)
		} else {
			m.To = ROOM
		}
		m.Body = "пщ"
		return s.Write(entity.Encode(dyn.NewMessage(m.Type, m.To, m.Body)))
	}
}

/* func doLua(script string) func(stream.Stream) error {
	return func(s stream.Stream) error {
		executor.Run(script)
		return nil
	}
}

func doJS(script string) func(stream.Stream) error {
	return func(s stream.Stream) error {
		jsexec.Run(script)
		return nil
	}
}

func doLuaAndPrint(script string) func(stream.Stream) error {
	return doLua(fmt.Sprintf(`chat.send(%s)`, script))
}
*/

func loadTpl(name string) (ret *template.Template, err error) {
	tn := filepath.Join("tpl", name+".tpl")
	if _, err = os.Stat(tn); err == nil {
		ret, err = template.ParseFiles(tn)
	}
	return
}

func bot(st stream.Stream) error {
	actors.With().Do(actors.C(steps.PresenceTo(units.Bare2Full(ROOM, ME), entity.CHAT, "ПЩ сюды: https://github.com/kpmy/xep"))).Run(st)
	//executor = luaexecutor.NewExecutor(st)
	//executor.Start()
	//jsexec = jsexecutor.NewExecutor(st)
	//jsexec.Start()
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
						posts.data = append(posts.data, Post{Nick: sender, User: user, Msg: e.Body})
						IncStat(user)
						IncStatLen(user, e.Body)
						posts.Unlock()
					}
					if sender != ME {
						/*
							executor.NewEvent(luaexecutor.IncomingEvent{"message",
														map[string]string{"sender": sender, "body": e.Body}})
													jsexec.NewEvent(jsexecutor.IncomingEvent{"message",
														map[string]string{"sender": sender, "body": e.Body}}) */
						hookExec.NewEvent(hookexecutor.IncomingEvent{"message",
							map[string]string{"sender": sender, "body": e.Body}})
						switch {
						case strings.EqualFold(strings.TrimSpace(e.Body), "пщ"):
							go func() {
								actors.With().Do(actors.C(doReply(sender, e.Type))).Run(st)
							}()
						}
						/*switch {
						case strings.HasPrefix(e.Body, "lua>"):
							go func(script string) {
								actors.With().Do(actors.C(doLua(script))).Run(st)
							}(strings.TrimPrefix(e.Body, "lua>"))
						case strings.HasPrefix(e.Body, "js>"):
							go func(script string) {
								actors.With().Do(actors.C(doJS(script))).Run(st)
							}(strings.TrimPrefix(e.Body, "js>"))
						case strings.HasPrefix(e.Body, "say"):
							go func(script string) {
								actors.With().Do(actors.C(doLuaAndPrint(script))).Run(st)
							}(strings.TrimSpace(strings.TrimPrefix(e.Body, "say")))
						} */
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
							//go func() { actors.With().Do(actors.C(doLuaAndPrint(`"` + user + `, насяльника..."`))).Run(st) }()
							/* executor.NewEvent(luaexecutor.IncomingEvent{"presence",
							map[string]string{"sender": sender, "user": user}}) */
							log.Println("ONLINE", user)
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
			log.Println(err)
			<-time.After(time.Second)
			dial(stream.New(s, redial))
		}

		redial(nil)
	}()
	go neo_server(wg)
	go func() {
		time.Sleep(time.Duration(time.Millisecond * 200))
		//open.Start("http://localhost:3000")
		open.Start("http://localhost:3000/stat")
	}()
	wg.Wait()
}
