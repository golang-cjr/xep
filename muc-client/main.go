package main

import (
	"bytes"
	"flag"
	"github.com/ivpusic/golog"
	"github.com/ivpusic/neo"
	"github.com/ivpusic/neo-cors"
	"github.com/ivpusic/neo/middlewares/logger"
	"github.com/skratchdot/open-golang/open"
	"html/template"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"xep/c2s/actors"
	"xep/c2s/actors/steps"
	"xep/c2s/stream"
	"xep/entity"
	"xep/units"
)

var (
	user     string
	pwd      string
	server   string
	resource string
	neo_log  = golog.GetLogger("application")
)

const tpl = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8"/>
	</head>
	<body>
		<h1>лог:</h1>
		{{range .Posts}}<em>{{.User}}</em>: {{.Msg}}<br/>{{else}}ничего ._.{{end}}
	</body>
</html>
`

type Post struct {
	User string
	Msg  string
}

type Posts struct {
	data []Post
	sync.Mutex
}

var posts *Posts

func init() {
	flag.StringVar(&user, "u", "goxep", "-u=user")
	flag.StringVar(&server, "s", "xmpp.ru", "-s=server")
	flag.StringVar(&resource, "r", "go", "-r=resource")
	flag.StringVar(&pwd, "p", "GogogOg0", "-p=password")
	log.SetFlags(0)
	posts = new(Posts)
}

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
				bind := &steps.Bind{Rsrc: resource + strconv.Itoa(rand.Intn(500))}
				actors.With(st).Do(auth.Act(), errHandler).Do(steps.Starter).Do(neg.Act()).Do(bind.Act()).Do(steps.Session).Run()
				actors.With(st).Do(steps.InitialPresence).Run()
				actors.With(st).Do(func(st stream.Stream) error {
					actors.With(st).Do(steps.PresenceTo("golang@conference.jabber.ru/xep")).Run()
					for {
						st.Ring(conv(func(_e entity.Entity) {
							switch e := _e.(type) {
							case *entity.Message:
								if strings.HasPrefix(e.From, "golang@conference.jabber.ru/") {
									posts.Lock()
									posts.data = append(posts.data, Post{User: strings.TrimPrefix(e.From, "golang@conference.jabber.ru/"),
										Msg: e.Body})
									posts.Unlock()
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
		t, _ := template.New("log").Parse(tpl)
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
			buf := bytes.NewBuffer(nil)
			t.Execute(buf, data)
			ctx.Res.Raw(buf.Bytes())
			return 200, nil
		})
		app.Start()
		wg.Done()
	}()
	go func() {
		time.Sleep(time.Duration(time.Millisecond * 200))
		open.Start("http://localhost:3000")
	}()
	wg.Wait()
}
