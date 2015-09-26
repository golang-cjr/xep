package main

import (
	"github.com/ivpusic/neo"
	"github.com/ivpusic/neo-cors"
	"github.com/ivpusic/neo/middlewares/logger"
	"sort"
	"sync"
)

func neo_server(wg *sync.WaitGroup) {
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
}
