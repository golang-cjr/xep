package main

import (
	"flag"
	"fmt"
	"github.com/ivpusic/golog"
	//	"github.com/skratchdot/open-golang/open"
	"github.com/kpmy/xep/luaexecutor"
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
	StatData struct {
		Total int
		Stat  []Stat
	}

	Stat struct {
		User  string
		Count float64
	}

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

var executor *luaexecutor.Executor

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

func doLua(script string) func(stream.Stream) error {
	return func(s stream.Stream) error {
		executor.Run(script)
		return nil
	}
}

func doLuaAndPrint(script string) func(stream.Stream) error {
	return doLua(fmt.Sprintf(`chat.send(%s)`, script))
}

func loadTpl(name string) (ret *template.Template, err error) {
	tn := filepath.Join("tpl", name+".tpl")
	if _, err = os.Stat(tn); err == nil {
		ret, err = template.ParseFiles(tn)
	}
	return
}

func bot(st stream.Stream) error {
	actors.With(st).Do(steps.PresenceTo(units.Bare2Full(ROOM, ME))).Run()
	executor = luaexecutor.NewExecutor(st)
	executor.Start()
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
						posts.Unlock()
					}
					if sender != ME {
						executor.NewMessage(luaexecutor.IncomingMessage{sender, e.Body})
						switch {
						/*case strings.EqualFold(strings.TrimSpace(e.Body), "пщ"):
						go func() {
							actors.With(st).Do(doReply(sender, e.Type)).Run()
						}() */
						case strings.HasPrefix(e.Body, "lua>"):
							go func(script string) {
								actors.With(st).Do(doLua(script)).Run()
							}(strings.TrimPrefix(e.Body, "lua>"))
						case strings.HasPrefix(e.Body, "say"):
							go func(script string) {
								actors.With(st).Do(doLuaAndPrint(script)).Run()
							}(strings.TrimSpace(strings.TrimPrefix(e.Body, "say")))
						}
					}
				}
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
		st := stream.New(s)
		if err := stream.Dial(st); err == nil {
			errHandler := func(err error) {
				log.Fatal(err)
			}
			neg := &steps.Negotiation{}
			actors.With(st).Do(steps.Starter, errHandler).Do(neg.Act(), errHandler).Run()
			if neg.HasMechanism("PLAIN") {
				auth := &steps.PlainAuth{Client: c, Pwd: pwd}
				neg := &steps.Negotiation{}
				bind := &steps.Bind{Rsrc: resource + strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(500))}
				actors.With(st).Do(auth.Act(), errHandler).Do(steps.Starter).Do(neg.Act()).Do(bind.Act()).Do(steps.Session).Run()
				actors.With(st).Do(steps.InitialPresence).Run()
				actors.With(st).Do(bot).Run()
			}
			wg.Done()
		} else {
			log.Fatal(err)
		}
	}()
	go neo_server(wg)
	go func() {
		time.Sleep(time.Duration(time.Millisecond * 200))
		//open.Start("http://localhost:3000")
		//open.Start("http://localhost:3000/stat")
	}()
	wg.Wait()
}
