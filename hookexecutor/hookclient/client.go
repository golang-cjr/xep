package hookclient

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/golang-cjr/xep/hookexecutor"
)

const (
	DefaultClientInboxSize  = 4
	DefaultClientOutboxSize = 4
)

type Client struct {
	addr string
	conn net.Conn

	prefixHandlers []stringMatchHandler
	substrHandlers []stringMatchHandler

	logger *log.Logger
	stop   chan struct{}
}

type Handler interface {
	HandleMessage(msg *hookexecutor.Message) (*hookexecutor.Message, error)
}

type HandlerFunc func(msg *hookexecutor.Message) (*hookexecutor.Message, error)

func (hf HandlerFunc) HandleMessage(msg *hookexecutor.Message) (*hookexecutor.Message, error) {
	return hf(msg)
}

type stringMatchHandler struct {
	prefix  string
	handler Handler
}

func NewClient(addr string) *Client {
	return &Client{
		addr,
		nil,
		nil,
		nil,
		log.New(os.Stderr, "[hookclient] ", log.LstdFlags),
		nil,
	}
}

func (c *Client) Start() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}

	c.conn = conn
	c.stop = make(chan struct{})

	go c.run()
	return nil
}

func (c *Client) Stop() {
	close(c.stop)
}

func (c *Client) Wait() {
	<-c.stop
}

func (c *Client) run() {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Println("panic recovered: %v", err)
		}
	}()

	inbox := make(chan *hookexecutor.Message, DefaultClientInboxSize)
	outbox := make(chan *hookexecutor.Message, DefaultClientOutboxSize)
	errors := make(chan error, 2)
	go c.reader(inbox, errors, c.stop)
	go c.writer(outbox, errors, c.stop)
	go c.stopOnError(c.stop, errors)

	for {
		select {
		case msg, ok := <-inbox:
			if !ok {
				return
			}

			if msg.Type == "ping" {
				pong := &hookexecutor.Message{&hookexecutor.IncomingEvent{"pong", nil}, -1}
				outbox <- pong
				continue
			}

			handlers := c.selectHandlers(msg)
			if len(handlers) > 0 {
				go c.executeHandlers(handlers, msg, outbox)
			}

		case <-c.stop:
			return
		}
	}
}

func (c *Client) selectHandlers(msg *hookexecutor.Message) []Handler {
	handlers := []Handler{}

	text := msg.Data["body"]
	for _, prefixHandler := range c.prefixHandlers {
		if strings.HasPrefix(text, prefixHandler.prefix) {
			handlers = append(handlers, prefixHandler.handler)
		}
	}

	for _, substrHandler := range c.substrHandlers {
		if strings.Contains(text, substrHandler.prefix) {
			handlers = append(handlers, substrHandler.handler)
		}
	}

	return handlers
}

func (c *Client) executeHandlers(handlers []Handler, msg *hookexecutor.Message, outbox chan *hookexecutor.Message) {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Println("panic recovered in handler executer: %v", err)
		}
	}()

	for _, handler := range handlers {
		reply, err := handler.HandleMessage(msg)
		if err != nil {
			c.logger.Printf("handler failed to handle message '%s': %v", msg, err)
			continue
		}

		if reply != nil {
			outbox <- reply
		}
	}
}

func (c *Client) reader(inbox chan *hookexecutor.Message, errors chan error, stop chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Println("panic recovered in reader: %v", err)
		}
	}()
	defer close(inbox)

	for {
		msg, err := hookexecutor.ReadMessage(c.conn, hookexecutor.DefaultHeartbeatTimeout)
		if err != nil {
			c.logger.Printf("reader failed to read message: %v", err)
			errors <- err
			return
		}

		inbox <- msg
	}
}

func (c *Client) writer(outbox chan *hookexecutor.Message, errors chan error, stop chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Println("panic recovered in writer: %v", err)
		}
	}()

	for {
		select {
		case msg := <-outbox:
			err := hookexecutor.WriteMessage(c.conn, hookexecutor.DefaultHeartbeatTimeout, msg)
			if err != nil {
				c.logger.Printf("writer failed to write message: %v", err)
				errors <- err
				return
			}
		case <-stop:
			return
		}
	}
}

func (c *Client) stopOnError(stop chan struct{}, errors chan error) {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Println("panic recovered in stopper: %v", err)
		}
	}()
	<-errors
	close(stop)
}

func (c *Client) HandlePrefix(prefix string, handler Handler) {
	c.prefixHandlers = append(c.prefixHandlers, stringMatchHandler{prefix, handler})
}

func (c *Client) HandleSubstr(needle string, handler Handler) {
	c.substrHandlers = append(c.substrHandlers, stringMatchHandler{needle, handler})
}
