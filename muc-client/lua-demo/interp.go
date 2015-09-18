package main

import (
	"bufio"
	"fmt"
	"github.com/kpmy/go-lua"
)

func main() {
	l := lua.NewState()
	lua.OpenLibraries(l)
	if err := lua.DoString(l, `print("Hello World")`); err != nil {
		panic(err)
	}
	l.Out.Seek(0, 0)
	rd := bufio.NewReader(l.Out)
	var err error
	for r := ' '; err == nil; {
		r, _, err = rd.ReadRune()
		fmt.Println(r)
	}
}
