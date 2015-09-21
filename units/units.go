package units

import (
	"github.com/kpmy/ypk/assert"
)

type Server struct {
	Name string
}

type Client struct {
	Name   string
	Server *Server
}

func Bare2Full(bare string, rsrc string) string {
	assert.For(bare != "", 20)
	assert.For(rsrc != "", 21)
	return bare + "/" + rsrc
}
