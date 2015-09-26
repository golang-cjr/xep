package main

import (
	"bytes"
	"github.com/kpmy/xippo/entity"
	"github.com/kpmy/xippo/entity/dyn"
	"github.com/kpmy/xippo/tools/dom"
	"gopkg.in/xmlpath.v2"
	"log"
)

func conv(fn func(entity.Entity)) func(*bytes.Buffer) bool {
	delayed := func(msg dom.Element) bool {
		for _, _e := range msg.Children() {
			if e, ok := _e.(dom.Element); ok && e.Name() == "delay" {
				return true
			}
		}
		return false
	}

	return func(in *bytes.Buffer) (done bool) {
		done = true
		log.Println("IN")
		log.Println(string(in.Bytes()))
		log.Println()
		if p, err := xmlpath.Parse(bytes.NewBuffer(in.Bytes())); err == nil {
			log.Println("xpath", p.String())
		} else {
			log.Println(err)
		}
		if _e, err := entity.Decode(bytes.NewBuffer(in.Bytes())); err == nil {
			e := _e.Model()
			switch e.Name() {
			case dyn.MESSAGE:
				if !delayed(e) {
					if ent, err := entity.ConsumeStatic(in); err == nil {
						fn(ent)
					} else {
						log.Println(err)
					}
				}
			}
		} else {
			log.Println(err)
		}
		return
	}
}
