package jsexecutor

import (
	"fmt"
	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"github.com/robertkrimen/otto"
	"sync"
	"time"
)

const sleepDuration time.Duration = 1 * time.Second

// An utility struct for incoming events.
type IncomingEvent struct {
	Type string
	Data map[string]string
}

// Executor executes JS scripts in a shared JS VM.
type Executor struct {
	incomingScripts chan string
	outgoingMsgs    chan string
	incomingEvents  chan IncomingEvent
	stateMutex      sync.Mutex
	eventHandlers   map[string]map[string]otto.Value
	vm              *otto.Otto
	xmppStream      stream.Stream
}

func NewExecutor(s stream.Stream) *Executor {
	e := &Executor{
		incomingScripts: make(chan string),
		outgoingMsgs:    make(chan string),
		incomingEvents:  make(chan IncomingEvent),
		eventHandlers:   make(map[string]map[string]otto.Value),
	}
	e.xmppStream = s
	e.vm = otto.New()

	send := func(call otto.FunctionCall) otto.Value {
		str, _ := call.Argument(0).ToString()
		e.outgoingMsgs <- str
		return otto.UndefinedValue()
	}

	addHandler := func(call otto.FunctionCall) otto.Value {
		evtName, err := call.Argument(0).ToString()
		handlerName, err := call.Argument(1).ToString()
		if err != nil {
			return otto.FalseValue()
		}
		val := call.Argument(2)
		if !val.IsFunction() {
			return otto.FalseValue()
		}
		handlers, ok := e.eventHandlers[evtName]
		if !ok {
			e.eventHandlers[evtName] = map[string]otto.Value{handlerName: val}
		} else {
			handlers[handlerName] = val
		}
		return otto.TrueValue()
	}

	listHandlers := func(call otto.FunctionCall) otto.Value {
		evtName, err := call.Argument(0).ToString()
		if err != nil {
			return otto.UndefinedValue()
		}
		list := []string{}
		for handlerName := range e.eventHandlers[evtName] {
			list = append(list, handlerName)
		}
		val, err := e.vm.ToValue(list)
		if err != nil {
			return otto.UndefinedValue()
		} else {
			return val
		}
	}

	chatLibrary, _ := e.vm.Object("Chat = {};")
	chatLibrary.Set("send", send)
	chatLibrary.Set("addEventHandler", addHandler)
	chatLibrary.Set("listEventHandlers", listHandlers)
	return e
}

func (e *Executor) execute() {
	for script := range e.incomingScripts {
		e.stateMutex.Lock()
		_, err := e.vm.Run(script)
		if err != nil {
			fmt.Printf("js fucking shit error: %s\n", err)
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
		obj, _ := e.vm.Object("({})")
		for key, value := range evt.Data {
			obj.Set(key, value)
		}
		for _, handler := range e.eventHandlers[evt.Type] {
			_, err := handler.Call(obj.Value(), obj.Value())
			if err != nil {
				fmt.Printf("js fucking shit error: %s\n", err)
				m := entity.MSG(entity.GROUPCHAT)
				m.To = "golang@conference.jabber.ru"
				m.Body = err.Error()
				e.xmppStream.Write(entity.ProduceStatic(m))
			}
		}
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
