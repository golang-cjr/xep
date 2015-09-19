package luaexecutor

import (
	"fmt"
	"github.com/kpmy/go-lua"
	"xep/c2s/stream"
	"xep/entity"
)

// Executor executes Lua scripts in a shared Lua VM.
type Executor struct {
	incomingScripts chan string
	state           *lua.State
	xmppStream      stream.Stream
}

func NewExecutor(s stream.Stream) *Executor {
	e := &Executor{incomingScripts: make(chan string)}
	e.xmppStream = s
	e.state = lua.NewState()
	lua.OpenLibraries(e.state)

	send := func(l *lua.State) int {
		m := entity.MSG(entity.GROUPCHAT)
		m.To = "golang@conference.jabber.ru"
		str, _ := l.ToString(1)
		m.Body = str
		err := e.xmppStream.Write(entity.Produce(m))
		if err != nil {
			fmt.Printf("lua shit error: %s", err)
			l.Error()
		}
		return 0
	}

	var chatLibrary = []lua.RegistryFunction{
		lua.RegistryFunction{"send", send},
	}

	lua.NewLibrary(e.state, chatLibrary)
	e.state.SetGlobal("chat")
	return e
}

func (e *Executor) execute() {
	for script := range e.incomingScripts {
		err := lua.DoString(e.state, script)
		fmt.Printf("lua fucking shit error: %s\n", err)
		if err != nil {
			m := entity.MSG(entity.GROUPCHAT)
			m.To = "golang@conference.jabber.ru"
			m.Body = err.Error()
			e.xmppStream.Write(entity.Produce(m))
		}
	}
}

func (e *Executor) Start() {
	go e.execute()
}

func (e *Executor) Stop() {
	close(e.incomingScripts)
}

func (e *Executor) Run(script string) {
	e.incomingScripts <- script
}
