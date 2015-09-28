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
const callbacksLocation string = "chatclbks"

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
		// get callbacks table
		l.PushString(callbacksLocation)
		l.Table(lua.RegistryIndex)
		// set callback
		l.PushValue(1)
		l.PushValue(2)
		l.SetTable(-3)
		l.SetTop(0)
		return 0
	}

	listClbks := func(l *lua.State) int {
		clbkNames := []string{}
		// get callbacks table
		l.PushString(callbacksLocation)
		l.Table(lua.RegistryIndex)
		// loop
		l.PushNil()
		for l.Next(-2) {
			key, ok := l.ToString(-2)
			// ignore non-string shit
			if ok {
				clbkNames = append(clbkNames, key)
			}
			l.Pop(1)
		}
		l.SetTop(0)
		l.NewTable()
		// build a list from callback names
		for i, key := range clbkNames {
			l.PushInteger(i + 1)
			l.PushString(key)
			l.SetTable(-3)
		}
		return 1
	}

	var chatLibrary = []lua.RegistryFunction{
		lua.RegistryFunction{"send", send},
		lua.RegistryFunction{"onmessage", registerClbk},
		lua.RegistryFunction{"listmsghandlers", listClbks},
	}

	lua.NewLibrary(e.state, chatLibrary)
	e.state.SetGlobal("chat")
	// set up callbacks table
	e.state.PushString(callbacksLocation)
	e.state.NewTable()
	e.state.SetTable(lua.RegistryIndex)
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
		// get callbacks table
		e.state.PushString(callbacksLocation)
		e.state.Table(lua.RegistryIndex)
		// loop over callbacks
		e.state.PushNil()
		for e.state.Next(-2) {
			if e.state.IsFunction(-1) {
				e.state.PushString(msg.Sender)
				e.state.PushString(msg.Body)
				err := e.state.ProtectedCall(2, 0, 0)
				if err != nil {
					m := entity.MSG(entity.GROUPCHAT)
					m.To = "golang@conference.jabber.ru"
					m.Body, _ = e.state.ToString(-1)
					e.xmppStream.Write(entity.ProduceStatic(m))
					e.state.Pop(1)
				}
			} else {
				e.state.Pop(1)
			}
		}
		// pop callbacks table
		e.state.Pop(1)
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
	e.incomingMsgs <- msg
}
