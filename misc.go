package main

import (
	"bytes"
	"github.com/kpmy/xippo/entity"
	"github.com/kpmy/xippo/entity/dyn"
	"github.com/kpmy/ypk/dom"
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
		//log.Println("IN")
		//log.Println(string(in.Bytes()))
		//log.Println()
		if _, err := xmlpath.Parse(bytes.NewBuffer(in.Bytes())); err == nil {

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
			case dyn.PRESENCE:
				fn(_e)
			}
		} else {
			log.Println(err)
		}
		return
	}
}

func firstByName(root dom.Element, name string) (ret dom.Element) {
	for _, x := range root.Children() {
		if e, ok := x.(dom.Element); ok && e.Name() == name {
			ret = e
			break
		}
	}
	return
}
