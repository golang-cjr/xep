package hookexecutor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/kpmy/xippo/c2s/stream"
	"github.com/kpmy/xippo/entity"
	"github.com/ugorji/go/codec"
)

const (
	DefaultAddr             = "0.0.0.0:1984"
	DefaultJSONAddr         = "0.0.0.0:1985"
	DefaultInboxBufferSize  = 4
	DefaultOutboxBufferSize = 4
	DefaultClientBufferSize = 8
	DefaultHeartbeatTrigger = 5 * time.Second
	DefaultHeartbeatTimeout = 10 * time.Second
	DefaultMessageLengthCap = 4 * 1024
)

type IncomingEvent struct {
	Type string
	Data map[string]string
}

type Message struct {
	*IncomingEvent
	ID int
}

type MessageWriter func(conn net.Conn, timeout time.Duration, msg *Message) error
type MessageReader func(conn net.Conn, timeout time.Duration) (*Message, error)

type clientReply struct {
	outbox chan *Message
	info   *clientInfo
}

type clientInfo struct {
	inbox chan *Message
	stop  chan struct{}
}

type Executor struct {
	listener     net.Listener
	jsonListener net.Listener
	xmppStream   stream.Stream
	logger       *log.Logger

	inbox          chan *IncomingEvent
	outbox         chan *Message
	cmdInbox       chan string
	clientRequests chan chan clientReply

	clients []*clientInfo
	counter int
}

func NewExecutor(s stream.Stream) *Executor {
	return &Executor{
		nil,
		nil,
		s,
		log.New(os.Stderr, "[hookexecutor] ", log.LstdFlags),
		make(chan *IncomingEvent, DefaultInboxBufferSize),
		make(chan *Message, DefaultOutboxBufferSize),
		make(chan string, DefaultInboxBufferSize),
		make(chan chan clientReply, DefaultInboxBufferSize),
		nil,
		0,
	}
}

func (exc *Executor) Start() {
	go exc.ListenAndServe(DefaultAddr)
	go exc.ListenAndServeJSON(DefaultJSONAddr)
	go exc.processEvents()
}

func (exc *Executor) Stop() {
	close(exc.inbox)
	close(exc.cmdInbox)
}

func (exc *Executor) Run(cmd string) {
	exc.cmdInbox <- cmd
}

func (exc *Executor) NewEvent(e IncomingEvent) {
	select {
	case exc.inbox <- &e:
	default:
		log.Println("executor is blocked")
	}
}

func stopPanic(exc *Executor, where string, callback func(err error)) {
	if err := recover(); err != nil {
		exc.logger.Printf("catched panic in %s: %s", where, err)
		if callback != nil {
			if realErr, ok := err.(error); ok {
				go callback(realErr)
			} else {
				go callback(fmt.Errorf("%v", err))
			}
		}
	}
}

func (exc *Executor) ListenAndServe(addr string) {
	defer stopPanic(exc, "listener", func(_ error) { exc.ListenAndServe(addr) })

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		exc.logger.Printf("failed to start listener, hooker disabled: %v", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			exc.logger.Printf("failed to accept new connection: %v", err)
			return
		}

		inbox, outbox := exc.createClient()
		stop := make(chan struct{})
		errors := make(chan error, 2)
		go exc.clientWriter(inbox, WriteMessage, conn, errors, stop)
		go exc.clientReader(outbox, ReadMessage, conn, errors, stop)
		go exc.stopOnError(stop, errors)
	}
}

func (exc *Executor) ListenAndServeJSON(addr string) {
	defer stopPanic(exc, "listener", func(_ error) { exc.ListenAndServeJSON(addr) })

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		exc.logger.Printf("failed to start listener, hooker disabled: %v", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			exc.logger.Printf("failed to accept new connection: %v", err)
			return
		}

		inbox, outbox := exc.createClient()
		stop := make(chan struct{})
		errors := make(chan error, 2)
		go exc.clientWriter(inbox, WriteJSONMessage, conn, errors, stop)
		go exc.clientReader(outbox, ReadJSONMessage, conn, errors, stop)
		go exc.stopOnError(stop, errors)
	}
}

func (exc *Executor) clientWriter(inbox chan *Message, write MessageWriter, conn net.Conn, errors chan error, stop chan struct{}) {
	defer stopPanic(exc, "clientWriter",
		func(err error) {
			exc.logger.Printf("catched panic in writer: %v", err)
			errors <- err
		})

	defer conn.Close()

	heartbeatTicker := time.NewTicker(DefaultHeartbeatTrigger)
	defer heartbeatTicker.Stop()

	for {
		select {
		case msg, ok := <-inbox:
			if !ok {
				close(stop)
				return
			}

			err := write(conn, DefaultHeartbeatTimeout, msg)
			if err != nil {
				exc.logger.Printf("failed to write message: %v", err)
				errors <- err
				return
			}
		case <-heartbeatTicker.C:
			ping := &Message{&IncomingEvent{"ping", nil}, -1}
			err := write(conn, DefaultHeartbeatTimeout, ping)
			if err != nil {
				exc.logger.Printf("failed to write ping message: %v", err)
				errors <- err
				return
			}
		case <-stop:
			return
		}
	}
}

func (exc *Executor) clientReader(outbox chan *Message, read MessageReader, conn net.Conn, errors chan error, stop chan struct{}) {
	defer stopPanic(exc, "clientReader",
		func(err error) {
			exc.logger.Printf("catched panic in reader: %v", err)
			errors <- err
		})
	defer conn.Close()

	for {
		msg, err := read(conn, DefaultHeartbeatTimeout)
		if err != nil {
			exc.logger.Printf("failed to read message: %v", err)
			errors <- err
			return
		}

		if msg.Type == "pong" {
			// ignore pongs, they are for resetting timeouts
			continue
		}

		select {
		case outbox <- msg:
		case <-stop:
			return
		}
	}
}

func (exc *Executor) stopOnError(stop chan struct{}, errors chan error) {
	defer stopPanic(exc, "stopper", nil)
	<-errors
	close(stop)
}

func ReadMessage(conn net.Conn, timeout time.Duration) (*Message, error) {
	conn.SetReadDeadline(time.Now().Add(DefaultHeartbeatTimeout))
	var lengthBuf [2]byte
	_, err := conn.Read(lengthBuf[:])
	if err != nil {
		return nil, err
	}

	length := int(binary.BigEndian.Uint16(lengthBuf[:]))
	if length > DefaultMessageLengthCap {
		return nil, errors.New("message is too long")
	}

	buf := bytes.NewBuffer(make([]byte, 0, length))
	_, err = io.CopyN(buf, conn, int64(length))
	if err != nil {
		return nil, err
	}

	var handle = &codec.MsgpackHandle{}
	var decoder = codec.NewDecoder(buf, handle)
	var result = &Message{}
	err = decoder.Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ReadJSONMessage(conn net.Conn, timeout time.Duration) (*Message, error) {
	conn.SetReadDeadline(time.Now().Add(DefaultHeartbeatTimeout))
	var handle = &codec.JsonHandle{}
	var decoder = codec.NewDecoder(conn, handle)
	var result = &Message{}
	err := decoder.Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func WriteMessage(conn net.Conn, timeout time.Duration, msg *Message) error {
	var handle = &codec.MsgpackHandle{}
	var buf []byte
	var encoder = codec.NewEncoderBytes(&buf, handle)
	err := encoder.Encode(msg)
	if err != nil {
		return err
	}

	length := len(buf)
	if length > DefaultMessageLengthCap {
		return errors.New("message is too long")
	}

	var lengthBuf [2]byte
	binary.BigEndian.PutUint16(lengthBuf[:], uint16(length))

	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = conn.Write(lengthBuf[:])
	if err != nil {
		return err
	}

	_, err = io.Copy(conn, bytes.NewBuffer(buf))
	return err
}

func WriteJSONMessage(conn net.Conn, timeout time.Duration, msg *Message) error {
	var handle = &codec.JsonHandle{}
	var buf []byte
	var encoder = codec.NewEncoderBytes(&buf, handle)
	err := encoder.Encode(msg)
	if err != nil {
		return err
	}

	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = io.Copy(conn, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte{'\n'})
	return err
}

func (exc *Executor) createClient() (inbox, outbox chan *Message) {
	reply := make(chan clientReply, 1)
	exc.clientRequests <- reply
	r := <-reply
	return r.info.inbox, r.outbox
}

func (exc *Executor) processEvents() {
	defer stopPanic(exc, "processEvents", func(_ error) { exc.processEvents() })

	for {
		select {
		case msg := <-exc.inbox:
			message := &Message{msg, exc.counter}
			exc.sendMessage(message)
			exc.counter++
		case cmd := <-exc.cmdInbox:
			// TODO(mechmind): handle cmds
			exc.logger.Printf("ignoring cmd: '%s'", cmd)
		case req := <-exc.clientRequests:
			outbox := exc.outbox

			info := &clientInfo{
				inbox: make(chan *Message, DefaultClientBufferSize),
				stop:  make(chan struct{}),
			}

			exc.clients = append(exc.clients, info)
			req <- clientReply{outbox, info}
		case msg := <-exc.outbox:
			exc.SendMessageToBot(msg)
		}
	}
}

func (exc *Executor) sendMessage(msg *Message) {
	deadClientIDs := []int{}

	for idx, ch := range exc.clients {
		select {
		case ch.inbox <- msg:
		default:
			deadClientIDs = append(deadClientIDs, idx)
		}
	}

	if len(deadClientIDs) == 0 {
		return
	}

	aliveClients := make([]*clientInfo, 0, len(exc.clients)-len(deadClientIDs))

	currentID := 0
	for idx, client := range exc.clients {
		if currentID < len(deadClientIDs) && idx == deadClientIDs[currentID] {
			// client is dead, drop him
			close(client.inbox)
			currentID++
		} else {
			// client alive, take him
			aliveClients = append(aliveClients, client)
		}
	}

	exc.clients = aliveClients
}

func (exc *Executor) SendMessageToBot(msg *Message) {
	m := entity.MSG(entity.GROUPCHAT)
	m.To = "golang@conference.jabber.ru"
	m.Body = msg.IncomingEvent.Data["body"]
	err := exc.xmppStream.Write(entity.ProduceStatic(m))
	if err != nil {
		exc.logger.Printf("failed to write message to xmpp stream: %v", err)
	}
}
