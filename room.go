package main

import (
	"errors"
	"sync"
)

type (
	Post struct {
		User string
		Nick string
		Msg  string
	}

	User struct {
		Nick   string
		Active bool
	}

	Room struct {
		posts []Post
		users map[string]*User
		sync.Mutex
	}
)

var room *Room

func (r *Room) New() *Room {
	r = new(Room)
	r.users = make(map[string]*User)
	return r
}
func (r *Room) Grow(p Post) {
	r.posts = append(r.posts, p)
}

func (r *Room) User(user string) (u *User) {
	ok := false
	if u, ok = r.users[user]; !ok {
		u = &User{Nick: user}
		r.users[user] = u
	}
	return
}

func (r *Room) Active(user string) {
	r.User(user).Active = true
}

func R(fn func(*Room) error) func(interface{}) (interface{}, error) {
	return func(x interface{}) (ret interface{}, err error) {
		if r, ok := x.(*Room); ok {
			ret = r
			r.Lock()
			err = fn(r)
			r.Unlock()
		} else {
			err = errors.New("unknown room")
		}
		return
	}
}
