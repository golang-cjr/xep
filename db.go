package main

import (
	"github.com/fjl/go-couchdb"
	"github.com/kpmy/ypk/halt"
)

const dbUrl = "http://127.0.0.1:5984"
const dbStatsName = "stats"
const dbWordsName = "words"

func init() {
	if client, err := couchdb.NewClient(dbUrl, nil); err == nil {
		statsDb, _ = client.CreateDB(dbStatsName)
		wordsDb, _ = client.CreateDB(dbWordsName)
	} else {
		halt.As(100, err)
	}
}
