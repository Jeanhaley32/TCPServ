package main

// Description: Defines message types and message broker.
import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

// Defining UID types.
// Node is any node that can send or receive messages, be it a client,or goroutine.
// Msg is any message sent or received by a node.
type (
	UID   uint32
	NID   UID // Node ID
	MsgID UID // Message I
)

// Defines payload type
type payload []byte

// Payload method to return payload as a string.
func (p payload) String() string {
	return string(p)
}

// Returns message type as string
func (c NID) IdType() string {
	return "node"
}

func (m MsgID) IdType() string {
	return "message"
}

// Define Enums for node types
const (
	Global NID = iota
	connlistener
	connhandler
	timekeeper
	messagebroker
)

// Returns name of node
func (c NID) Name() string {
	switch c {
	case Global:
		return "Global"
	case connlistener:
		return "listener"
	case connhandler:
		return "Handler"
	case timekeeper:
		return "Timekeeper"
	case messagebroker:
		return "MessageBroker"
	default:
		return fmt.Sprintf("Node%v", c)
	}
}

// create a Termination channel type with blank interface
type termch chan interface{}

// create a channel type with blank interface
type ch chan msg

// define route type.
type route struct {
	source      NID
	destination NID
}

// Message "object"
// individual message received from connection.
type msg struct {
	MsgColor     Color       // color used to represent message
	destination  NID         // destination of message
	source       NID         // source of message
	ID           MsgID       // Unique Identifier for message
	payload      payload     // msg payload as a byte array
	timeReceived time.Time   // Time Message was received
	msgType      MsgEnumType // Message type. Used to define message route.
}

// Defines interface needed for message handler
type message interface {
	ColorWrap() string       // returns color wraped payload
	SetSource(NID)           // sets message source
	SetDesitnation(NID)      // sets message destination
	SetType(MsgEnumType)     // sets message type
	SetPayload(payload)      // sets message payload
	GetDestination() NID     // returns destination
	GetPayload() payload     // returns payload
	GetTimestamp() timestamp // returns timestamp
	GetId() MsgID            // returns message id
	GetMsgType() MsgEnumType // returns message type
	GetSource() NID          // returns message source
}

// returns source of message
func (m msg) GetSource() NID {
	return m.source
}

// Sets msg destination
func (m *msg) SetDestination(n NID) {
	m.destination = n
}

// wraps payload in color based on message type
func (m msg) ColorWrap() string {
	var newPayload []byte
	const reset = "\033[0m"
	newPayload = payload(m.GetMsgColor().Color() + string(m.payload) + reset)
	return string(newPayload)
}

// sets message type
func (m *msg) SetType(t MsgEnumType) {
	m.msgType = t
}

// sets color used to represent message.
func (m *msg) SetMsgColor(c *Color) {
	if c != nil {
		m.MsgColor = *c
		return
	}
	if m.msgType == Client {
		m.MsgColor = Green
	}
	if m.msgType == Error {
		m.MsgColor = Red
	}
	if m.msgType == System {
		m.MsgColor = Yellow
	}
}

// gets msg color
func (m msg) GetMsgColor() Color {
	return m.MsgColor
}

// sets message source
func (m *msg) SetSource(n NID) {
	m.source = n
}

// initializes a message object
//
//	generates a unique message id
//	sets time to current time
//	sets payload to byte array
//	sets message type.
func InitMsg(b []byte, t MsgEnumType, r route, c Color) (msg, error) {
	// if destination is not 0, set destination to destination
	// else destination is considered global.
	destination := Global
	if r.destination != 0 {
		destination = r.destination
	}
	m := msg{}
	m.SetMsgColor(&c)
	m.SetSource(r.source)
	m.SetDestination(destination)
	m.generateUid()
	m.setTime()
	m.SetPayload(payload(b))
	m.SetType(t)
	return m, nil
}

// Get msg destionation
func (m msg) GetDestination() NID {
	if m.destination == 0 {
		return NID(0)
	}
	return m.destination
}

// sets t to current time
func (m *msg) setTime() {
	m.timeReceived = time.Now()
}

// Return payload
func (m msg) GetPayload() payload {
	return m.payload
}

// Sets payload
func (m *msg) SetPayload(p payload) {
	m.payload = p
}

// Returns Timestamp in Month/Day/Year Hour:Minute:Second format
func (m msg) GetTimestamp() timestamp {
	return timestamp(m.timeReceived.Format("2006-01-02T15:04:05Z"))
}

// return message Id
func (m msg) GetId() MsgID {
	return m.ID
}

// return message type
func (m msg) GetMsgType() MsgEnumType {
	return m.msgType.Type()
}

// Generates a unique message id
func (m *msg) generateUid() {
	m.ID = MsgID(uuid.New().ID())
}

// Defining type used to define a message route and purpose
type MsgEnumType int64

const (
	Client MsgEnumType = iota
	Error
	System
)

// Returns msg type
func (m MsgEnumType) Type() MsgEnumType {
	switch m {
	case Client:
		return Client
	case Error:
		return Error
	case System:
		return System
	}
	return System
}

// returns MsgEnumType as a string
func (m MsgEnumType) String() string {
	switch m {
	case Client:
		return "Client"
	case Error:
		return "Error"
	case System:
		return "System"
	}
	return "System"
}

// Returns Global message routing channel based on msg type
func (m MsgEnumType) GetChannel() ch {
	switch m {
	case Client:
		return clientChan
	case Error:
		return logChan
	case System:
		return sysChan
	}
	return sysChan
}

// Writes message to channel based on msg type.
func (m MsgEnumType) WriteToChannel(a msg) {
	a.SetType(m)
	m.GetChannel() <- a
}

// Logs Payload, does not write to Client.
// Used for logging purposes.
func (m MsgEnumType) LogPayload(s NID, p string) {
	System.WriteToChannel(
		msg{
			payload: payload(fmt.Sprintf("%v: %v", s.Name(), p)),
			source:  s,
		},
	)
}

// Error Logging Payload, does not write to Client.
func (m MsgEnumType) LogError(s NID, p string) {
	Error.WriteToChannel(
		msg{
			payload: payload(fmt.Sprintf("%v:%v", s.Name(), p)),
			source:  s,
		},
	)
}

// Reads from Channel.
func (m MsgEnumType) ReadFromChannel() interface{} {
	return <-m.GetChannel()
}

// MessageBroker is used to handle messages from the three global channels.
// It is responsible for writing messages to the appropriate connections.
// It is also responsible for logging messages to the console.
func MessageBroker() {
	// Create a custom logger
	logger := log.New(os.Stdout, "", log.LstdFlags)
	// defering exit routine for eventHandler.
	defer func() { logger.Printf(colorWrap(Red, "Exiting Error Logger")) }()
	for {
		select {
		case m := <-clientChan:
			// if globalState is full, pop first element and append new message.
			newPayload := payload(fmt.Sprintf("%v\n", m.GetPayload()))
			m.SetPayload(newPayload)
			if len(globalState) == ClientMessageCount {
				globalState = globalState[1:]
				globalState = append(globalState, m)
			} else {
				globalState = append(globalState, m)
			}
			currentstate.WriteScreen()
		case m := <-sysChan:
			logger.Print(m.ColorWrap())
		case m := <-logChan:
			logger.Print(m.ColorWrap())
		}
	}
}
