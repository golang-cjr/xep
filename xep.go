package main

import (
	"log"
	"xep/stream"
)

func init() {
	log.SetFlags(0)
}

func main() {
	if cli, err := stream.Dial(&stream.Client{User: "goxep", Server: "xmpp.ru", Resource: "go", Pwd: "GogogOg0"}); err == nil {
		log.Println("connected")
		for err := cli.Start(func() error {
			cli.Status()
			p := &stream.Presence{}
			p.Prepare()
			p.To = "golang@conference.jabber.ru/xep"
			cli.Do(p)
			return nil
		}); err == nil; {
			err = cli.Process()
		}
		defer cli.Stop()
	} else {
		log.Fatal(err)
	}
}
