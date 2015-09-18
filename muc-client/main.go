package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/ivpusic/golog"
	"github.com/ivpusic/neo"
	"github.com/ivpusic/neo-cors"
	"github.com/ivpusic/neo/middlewares/logger"
	"github.com/kpmy/go-lua"
	"github.com/skratchdot/open-golang/open"
	"html/template"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"xep/c2s/actors"
	"xep/c2s/actors/steps"
	"xep/c2s/stream"
	"xep/entity"
	"xep/muc-client/muc"
	"xep/units"
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

func conv(fn func(entity.Entity)) func(*bytes.Buffer) bool {
	return func(in *bytes.Buffer) (done bool) {
		done = true
		log.Println("IN")
		log.Println(string(in.Bytes()))
		log.Println()
		if _e, err := entity.Consume(in); err == nil {
			switch e := _e.(type) {
			case *entity.Message:
				fn(e)
			}
		} else {
			log.Println(err)
		}
		return
	}
}

func doReply(s stream.Stream) error {
	m := entity.MSG(entity.GROUPCHAT)
	m.Body = "пщ"
	m.To = "golang@conference.jabber.ru"
	return s.Write(entity.Produce(m))
}

func doLua(script string) func(stream.Stream) error {
	return func(s stream.Stream) error {
		m := entity.MSG(entity.GROUPCHAT)
		m.To = "golang@conference.jabber.ru"
		defer func() {
			if r := recover(); r != nil {
				m.Body = fmt.Sprint(r)
				s.Write(entity.Produce(m))
			}
		}()
		l := lua.NewState()
		lua.OpenLibraries(l)
		if err := lua.DoString(l, script); err == nil {
			l.Out.Seek(0, 0)
			rd := bufio.NewReader(l.Out)
			var err error
			var rr []rune
			for r := ' '; err == nil; {
				if r, _, err = rd.ReadRune(); err == nil {
					rr = append(rr, r)
				}
			}
			m.Body = string(rr)
		} else {
			m.Body = fmt.Sprint(err)
		}
		return s.Write(entity.Produce(m))
	}
}

func loadTpl(name string) (ret *template.Template, err error) {
	tn := filepath.Join("tpl", name+".tpl")
	if _, err = os.Stat(tn); err == nil {
		ret, err = template.ParseFiles(tn)
	}
	return
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
				actors.With(st).Do(func(st stream.Stream) error {
					actors.With(st).Do(steps.PresenceTo("golang@conference.jabber.ru/xep")).Run()
					for {
						st.Ring(conv(func(_e entity.Entity) {
							switch e := _e.(type) {
							case *entity.Message:
								if strings.HasPrefix(e.From, "golang@conference.jabber.ru/") {
									sender := strings.TrimPrefix(e.From, "golang@conference.jabber.ru/")
									um := muc.UserMapping()
									user := sender
									if u, ok := um[sender]; ok {
										user, _ = u.(string)
									}
									posts.Lock()
									posts.data = append(posts.data, Post{Nick: sender, User: user, Msg: e.Body})
									posts.Unlock()
									if sender != "xep" {
										switch {
										case strings.EqualFold(e.Body, "пщ"):
											go func() {
												actors.With(st).Do(doReply).Run()
											}()
										case strings.HasPrefix(e.Body, "lua>"):
											go func(script string) {
												actors.With(st).Do(doLua(script)).Run()
											}(strings.TrimPrefix(e.Body, "lua>"))
										}
									}
								}
							}
						}), 0)
					}
				}).Run()
			}
			wg.Done()
		} else {
			log.Fatal(err)
		}
	}()
	go func() {
		app := neo.App()
		app.Use(logger.Log)
		app.Use(cors.Init())
		//app.Templates("tpl/*.tpl") //кэширует в этом месте и далее не загружает с диска, сука
		app.Serve("/static", "static")
		app.Get("/", func(ctx *neo.Ctx) (int, error) {
			posts.Lock()
			data := struct {
				Posts []Post
			}{}
			for i := len(posts.data) - 1; i >= 0; i-- {
				p := posts.data[i]
				data.Posts = append(data.Posts, p)
			}
			posts.Unlock()

			if t, err := loadTpl("log"); t != nil {
				//ctx.Res.Tpl("log.tpl", data)
				t.Execute(ctx.Res, data)
				return 200, nil
			} else {
				return 500, err
			}
		})
		app.Get("/stat", func(ctx *neo.Ctx) (int, error) {
			mm := make(map[string]int)
			total := 0
			posts.Lock()
			for _, p := range posts.data {
				n := 0
				if old, ok := mm[p.User]; ok {
					n = old + 1
				} else {
					n = 1
				}
				mm[p.User] = n
			}
			total = len(posts.data)
			posts.Unlock()
			data := &StatData{Total: total}
			for u, c := range mm {
				s := Stat{User: u}
				s.Count = float64(c) / float64(total) * 100
				data.Stat = append(data.Stat, s)
			}
			sort.Stable(data)
			if t, err := loadTpl("stat"); t != nil {
				t.Execute(ctx.Res, data)
				return 200, nil
			} else {
				return 500, err
			}
		})
		app.Start()
		wg.Done()
	}()
	go func() {
		time.Sleep(time.Duration(time.Millisecond * 200))
		open.Start("http://localhost:3000")
		//open.Start("http://localhost:3000/stat")
	}()
	wg.Wait()
}
