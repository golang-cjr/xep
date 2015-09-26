package luaexecutor

import (
	"fmt"
	"github.com/kpmy/go-lua"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"sync"
	"time"
)

const sleepDuration time.Duration = 1 * time.Second
const callbackLocation string = "chatclbk"

// An utility struct for incoming messages.
type IncomingMessage struct {
	Sender string
	Body   string
}

// Executor executes Lua scripts in a shared Lua VM.
type Executor struct {
	incomingScripts chan string
	outgoingMsgs    chan string
	incomingMsgs    chan IncomingMessage
	stateMutex      sync.Mutex
	state           *lua.State
	xmppStream      stream.Stream
}

func NewExecutor(s stream.Stream) *Executor {
	e := &Executor{
		incomingScripts: make(chan string),
		outgoingMsgs:    make(chan string),
		incomingMsgs:    make(chan IncomingMessage),
	}
	e.xmppStream = s
	e.state = lua.NewState()
	lua.OpenLibraries(e.state)

	send := func(l *lua.State) int {
		str, _ := l.ToString(1)
		e.outgoingMsgs <- str
		return 0
	}

	registerClbk := func(l *lua.State) int {
		l.PushString(callbackLocation)
		l.PushValue(-2)
		l.SetTable(lua.RegistryIndex)
		return 0
	}

	var chatLibrary = []lua.RegistryFunction{
		lua.RegistryFunction{"send", send},
		lua.RegistryFunction{"onmessage", registerClbk},
	}

	lua.NewLibrary(e.state, chatLibrary)
	e.state.SetGlobal("chat")
	return e
}

func (e *Executor) execute() {
	for script := range e.incomingScripts {
		e.stateMutex.Lock()
		err := lua.DoString(e.state, script)
		if err != nil {
			fmt.Printf("lua fucking shit error: %s\n", err)
			m := entity.MSG(entity.GROUPCHAT)
			m.To = "golang@conference.jabber.ru"
			m.Body = err.Error()
			e.xmppStream.Write(entity.ProduceStatic(m))
		}
		e.stateMutex.Unlock()
	}
}

func (e *Executor) sendingRoutine() {
	for msg := range e.outgoingMsgs {
		m := entity.MSG(entity.GROUPCHAT)
		m.To = "golang@conference.jabber.ru"
		m.Body = msg
		err := e.xmppStream.Write(entity.ProduceStatic(m))
		if err != nil {
			fmt.Printf("send error: %s", err)
		}
		time.Sleep(sleepDuration)
	}
}

func (e *Executor) processIncomingMsgs() {
	for msg := range e.incomingMsgs {
		e.stateMutex.Lock()
		e.state.PushString(callbackLocation)
		e.state.Table(lua.RegistryIndex)
		if e.state.IsFunction(-1) {
			e.state.PushString(msg.Sender)
			e.state.PushString(msg.Body)
			err := e.state.ProtectedCall(2, 0, 0)
			if err != nil {
				m := entity.MSG(entity.GROUPCHAT)
				m.To = "golang@conference.jabber.ru"
				m.Body, _ = e.state.ToString(-1)
				e.xmppStream.Write(entity.ProduceStatic(m))
			}
		} else {
			e.state.Pop(1)
		}
		e.stateMutex.Unlock()
	}
}

func (e *Executor) Start() {
	go e.sendingRoutine()
	go e.execute()
	go e.processIncomingMsgs()
}

func (e *Executor) Stop() {
	close(e.incomingScripts)
	close(e.incomingMsgs)
	close(e.outgoingMsgs)
}

func (e *Executor) Run(script string) {
	e.incomingScripts <- script
}

// Call this on every incoming message - it's required for
// chat.onmessage to work.
func (e *Executor) NewMessage(msg IncomingMessage) {
	select {
	case e.incomingMsgs <- msg:
	default:
	}
}
