package luaexecutor

import (
	"fmt"
	"github.com/kpmy/go-lua"
	"time"
	"xep/c2s/stream"
	"xep/entity"
)

const sleepDuration time.Duration = 1 * time.Second

// Executor executes Lua scripts in a shared Lua VM.
type Executor struct {
	incomingScripts chan string
	outgoingMsgs    chan string
	state           *lua.State
	xmppStream      stream.Stream
}

func NewExecutor(s stream.Stream) *Executor {
	e := &Executor{incomingScripts: make(chan string), outgoingMsgs: make(chan string)}
	e.xmppStream = s
	e.state = lua.NewState()
	lua.OpenLibraries(e.state)

	send := func(l *lua.State) int {
		str, _ := l.ToString(1)
		e.outgoingMsgs <- str
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
		if err != nil {
			fmt.Printf("lua fucking shit error: %s\n", err)
			m := entity.MSG(entity.GROUPCHAT)
			m.To = "golang@conference.jabber.ru"
			m.Body = err.Error()
			e.xmppStream.Write(entity.Produce(m))
		}
	}
}

func (e *Executor) sendingRoutine() {
	for msg := range e.outgoingMsgs {
		m := entity.MSG(entity.GROUPCHAT)
		m.To = "golang@conference.jabber.ru"
		m.Body = msg
		err := e.xmppStream.Write(entity.Produce(m))
		if err != nil {
			fmt.Printf("send error: %s", err)
		}
		time.Sleep(sleepDuration)
	}
}

func (e *Executor) Start() {
	go e.sendingRoutine()
	go e.execute()
}

func (e *Executor) Stop() {
	close(e.incomingScripts)
}

func (e *Executor) Run(script string) {
	e.incomingScripts <- script
}
