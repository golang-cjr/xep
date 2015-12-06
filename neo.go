package main

import (
	"bytes"
	"github.com/ivpusic/golog"
	"github.com/ivpusic/neo"
	"github.com/ivpusic/neo-cors"
	"github.com/ivpusic/neo/middlewares/logger"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type (
	TplData struct {
		Count *StatData
		Total *StatData
		Deads []*User
	}

	StatData struct {
		Total int
		Stat  []Stat
	}

	Stat struct {
		User  string
		Count int64
		Perc  float64
	}
)

var neo_log = golog.GetLogger("application")

func (d *StatData) Len() int { return len(d.Stat) }

func (d *StatData) Less(i, j int) bool { return d.Stat[i].Count > d.Stat[j].Count }

func (d *StatData) Swap(i, j int) { d.Stat[i], d.Stat[j] = d.Stat[j], d.Stat[i] }

func loadTpl(name string) (ret *template.Template, err error) {
	tn := filepath.Join("tpl", name+".tpl")
	if _, err = os.Stat(tn); err == nil {
		ret, err = template.ParseFiles(tn)
	}
	return
}

func neo_server(wg *sync.WaitGroup) {
	app := neo.App()
	app.Use(logger.Log)
	app.Use(cors.Init())
	//app.Templates("tpl/*.tpl") //кэширует в этом месте и далее не загружает с диска, сука
	app.Serve("/static", "static")
	app.Get("/favicon.ico", func(ctx *neo.Ctx) (int, error) {
		buf := bytes.NewBuffer(nil)
		ico(buf)
		io.Copy(ctx.Res, buf)
		return 200, nil
	})
	app.Get("/", func(ctx *neo.Ctx) (int, error) {
		data := struct {
			Posts []Post
		}{}
		room.Lock()
		for i := len(room.posts) - 1; i >= 0; i-- {
			p := room.posts[i]
			data.Posts = append(data.Posts, p)
		}
		room.Unlock()

		if t, err := loadTpl("log"); t != nil {
			//ctx.Res.Tpl("log.tpl", data)
			t.Execute(ctx.Res, data)
			return 200, nil
		} else {
			return 500, err
		}
	})
	app.Get("/stat", func(ctx *neo.Ctx) (int, error) {

		conv := func(t *CStatDoc) (td *StatData) {
			mm := t.Data
			total := t.Total
			td = &StatData{Total: total}
			for u, c := range mm {
				s := Stat{User: u}
				s.Count = int64(c)
				s.Perc = float64(c) / float64(total) * 100
				td.Stat = append(td.Stat, s)
			}
			sort.Stable(td)
			return
		}

		var t, c *CStatDoc
		var err error
		if t, err = GetStat(totalId); err == nil {
			if c, err = GetStat(countId); err == nil {

				data := &TplData{}
				data.Count = conv(c)
				data.Total = conv(t)
				room.Lock()
				for _, u := range room.users {
					if !u.Active {
						data.Deads = append(data.Deads, u)
					}
				}
				room.Unlock()
				var tpl *template.Template
				if tpl, err = loadTpl("stat"); tpl != nil {
					tpl.Execute(ctx.Res, data)
					return 200, nil
				}
			}
		}
		return 500, err
	})
	app.Start()
	wg.Done()
}
