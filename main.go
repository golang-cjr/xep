package main

import (
	"flag"
	//"github.com/skratchdot/open-golang/open"
	"log"

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
)

func init() {
	flag.StringVar(&user, "u", "goxep", "-u=user")
	flag.StringVar(&server, "s", "xmpp.ru", "-s=server")
	flag.StringVar(&resource, "r", "go", "-r=resource")
	flag.StringVar(&pwd, "p", "GogogOg0", "-p=password")
	log.SetFlags(0)
	posts = new(Posts)
}

func main() {
	flag.Parse()
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go xmpp(wg)
	go neo_server(wg)
	go func() {
		time.Sleep(time.Duration(time.Millisecond * 200))
		//open.Start("http://localhost:3000")
		//open.Start("http://localhost:3000/stat")
	}()
	wg.Wait()
}
