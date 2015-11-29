package main

import (
	"github.com/fjl/go-couchdb"
	"log"
)

type CStatDoc struct {
	Total int
	Data  map[string]int
}

var statsDb *couchdb.DB

func GetStat(docId string) (ret *CStatDoc, err error) {
	ret = &CStatDoc{}
	if err = statsDb.Get(docId, ret, nil); err == nil {
		if ret.Data == nil {
			ret.Data = make(map[string]int)
		}
	} else if couchdb.NotFound(err) {
		if _, err = statsDb.Put(docId, ret, ""); err == nil {
			ret, err = GetStat(docId)
		}
	}
	return
}

func SetStat(docId string, old *CStatDoc) {
	if rev, err := statsDb.Rev(docId); err == nil {
		if _, err = statsDb.Put(docId, old, rev); err != nil {
			log.Println(err)
		}
	}
}

const countId = "count"
const totalId = "total"

func IncStat(user string) {
	if s, err := GetStat(totalId); err == nil {
		if _, ok := s.Data[user]; ok {
			s.Data[user] = s.Data[user] + 1
		} else {
			s.Data[user] = 1
		}
		s.Total++
		SetStat(totalId, s)
	}
}

func IncStatLen(user, msg string) {
	count := len([]rune(msg))
	if s, err := GetStat(countId); err == nil {
		if _, ok := s.Data[user]; ok {
			s.Data[user] = s.Data[user] + count
		} else {
			s.Data[user] = count
		}
		s.Total += count
		SetStat(countId, s)
	}
}
