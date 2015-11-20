package main

import (
	"sync"
)

type (
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
