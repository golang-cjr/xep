package luaexecutor

import (
	"fmt"
	"github.com/Shopify/go-lua"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"path/filepath"
	"sync"
	"time"
)

const sleepDuration time.Duration = 1 * time.Second
const callbacksLocation string = "clbks"

// An utility struct for incoming events.
type IncomingEvent struct {
	Type string
	Data map[string]string
}

// Executor executes Lua scripts in a shared Lua VM.
type Executor struct {
	incomingScripts chan string
	outgoingMsgs    chan string
	incomingEvents  chan IncomingEvent
	stateMutex      sync.Mutex
	state           *lua.State
	xmppStream      stream.Stream
}

func NewExecutor(s stream.Stream) *Executor {
	e := &Executor{
		incomingScripts: make(chan string),
		outgoingMsgs:    make(chan string),
		incomingEvents:  make(chan IncomingEvent),
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
		// get events table
		l.PushString(callbacksLocation)
		l.Table(lua.RegistryIndex)
		// get callbacks table
		evtName := lua.CheckString(l, 1)
		l.PushString(evtName)
		l.Table(-2)
		// create new table if one doesn't exist
		if l.IsNil(-1) {
			l.Pop(1)
			// create and add table to the events table
			l.PushString(evtName)
			l.NewTable()
			l.SetTable(-3)
			// and get it back
			l.PushString(evtName)
			l.Table(-2)
		}
		// set callback
		l.PushValue(2)
		l.PushValue(3)
		l.SetTable(-3)
		l.SetTop(0)
		return 0
	}

	listClbks := func(l *lua.State) int {
		clbkNames := []string{}
		// get events table
		l.PushString(callbacksLocation)
		l.Table(lua.RegistryIndex)
		// get callbacks for the event
		evtName := lua.CheckString(l, 1)
		l.PushString(evtName)
		l.Table(-2)
		if !l.IsNil(-1) {
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
		} else {
			l.SetTop(0)
			l.PushNil()
		}
		return 1
	}

	var chatLibrary = []lua.RegistryFunction{
		lua.RegistryFunction{"send", send},
		lua.RegistryFunction{"addEventHandler", registerClbk},
		lua.RegistryFunction{"listEventHandlers", listClbks},
	}

	lua.NewLibrary(e.state, chatLibrary)
	e.state.SetGlobal("chat")
	// set up callbacks table
	e.state.PushString(callbacksLocation)
	e.state.NewTable()
	e.state.SetTable(lua.RegistryIndex)
	lua.DoFile(e.state, filepath.Join("static", "bootstrap.lua"))
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

func (e *Executor) processIncomingEvents() {
	for evt := range e.incomingEvents {
		e.stateMutex.Lock()
		// get events table
		e.state.PushString(callbacksLocation)
		e.state.Table(lua.RegistryIndex)
		// get callbacks table for the specific event
		e.state.PushString(evt.Type)
		e.state.Table(-2)
		// loop over callbacks
		if !e.state.IsNil(-1) {
			e.state.PushNil()
			for e.state.Next(-2) {
				if e.state.IsFunction(-1) {
					// create the table which will be passed to the handler
					e.state.NewTable()
					// loop over the event data, populating the table
					for k, v := range evt.Data {
						e.state.PushString(k)
						e.state.PushString(v)
						e.state.SetTable(-3)
					}
					err := e.state.ProtectedCall(1, 0, 0)
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
		}
		// pop callbacks table or nil value
		e.state.Pop(1)
		e.stateMutex.Unlock()
	}
}

func (e *Executor) Start() {
	go e.sendingRoutine()
	go e.execute()
	go e.processIncomingEvents()
}

func (e *Executor) Stop() {
	close(e.incomingScripts)
	close(e.incomingEvents)
	close(e.outgoingMsgs)
}

func (e *Executor) Run(script string) {
	e.incomingScripts <- script
}

// Call this on every event - it's required for event handlers to work
func (e *Executor) NewEvent(evt IncomingEvent) {
	e.incomingEvents <- evt
}
