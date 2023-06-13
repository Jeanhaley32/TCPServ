package main

// TODO(jeanhaley) - The following items need to be addressed:
//		- Rethink the "state handler" and how it is going to be used. maybe make it a "connection manager".
//		- Create a flow chart that shows the flow of how a client []byte is wrapped in a "msg", and routed through
//	          the system. I feel that at the moment there is no real rhyme or reason for this, and this needs to be
//		  codified.
// 		- enable the acceptance of CLI arguments to set the IP, Port, and Buffer size.
//		- re-factor code. Breakdown individual, independent functions into their own files.
//              - More Long term
//			- Break off individual components into seperate "micro-services" using GRPC for communication.
// 			- use Bubbtletea to create a CLI interface for the server.

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/google/uuid"
)

const (
	corgi = "⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⣧⣼⣧⠀⠀⠀⠀⠀⠀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣭⣭⣤⣄⠀⠀⠀⠀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⣿⣿⣿⣿⣷⣤⣤⡄\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣼⣿⣮⣍⣉⣉⣀⣀⠀⠀⠀\n" +
		"⠀⠀⣠⣶⣶⣶⣶⣶⣶⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⠀\n" +
		"⣴⣿⣿⣿⣿⣿⣯⡛⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀\n" +
		"⠉⣿⣿⣿⣿⣿⣿⣷⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠏⠀⠀\n" +
		"⠀⣿⣿⣿⣿⣿⣿⡟⠸⠿⠿⠿⣿⣿⣿⣿⣿⣿⣿⠿⠋⠀⠀⠀\n" +
		"⠀⠘⢿⣿⣿⠿⠋⠀⠀⠀⠀⠀⠀⠉⠉⣿⣿⡏⠁⠀⠀⠀⠀⠀\n" +
		"⠀⠀⢸⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⡇⠀⠀⠀⠀⠀⠀\n"
)

// ___ Global Variables ___

var (
	ip, netp, port, banner                                                            string
	buffersize, logerTime, clientchannelbuffer, logchannelbuffer, systemchannelbuffer int
	clientChan, logChan, sysChan                                                      ch // Global Channels
	currentstate                                                                      state
)

var (
	// creating a blank global branding variable.
	// this needs to be done, because the type figure.figure is not exported.
	branding = figure.NewColorFigure("", "nancyj-fancy", "Blue", true)
)

// uses init function to set set up global flag variables, and channels.
func init() {
	// setting Global Flags
	flag.StringVar(&ip, "ip", "127.0.0.1", "IP for server to listen on")
	flag.StringVar(&netp, "netp", "tcp", "Network protocol to use")
	flag.StringVar(&port, "port", "6000", "Port for server to listen on")
	flag.IntVar(&buffersize, "bufferSize", 1024, "Message Buffer size.")
	flag.IntVar(&logerTime, "logerTime", 120, "time in between server status check, in seconds.")
	flag.StringVar(&banner, "banner", "TheVoid", "Banner to display on startup")
	flag.IntVar(&clientchannelbuffer, "clientchannelbuffer", 20, "size of client channel buffer")
	flag.IntVar(&logchannelbuffer, "logchannelbuffer", 20, "size of log channel buffer")
	flag.IntVar(&systemchannelbuffer, "systemchannelbuffer", 20, "size of system channel buffer")
	flag.Parse()
	// instantiating global channels.
	clientChan = make(chan msg, clientchannelbuffer)
	logChan = make(chan msg, logchannelbuffer)
	sysChan = make(chan msg, systemchannelbuffer)
	branding = figure.NewColorFigure(banner, "nancyj-fancy", "Blue", true) // sets banner to value passed by terminal flags.

}

// ___ Global Channel Variables ___
// create a channel type with blank interface
type ch chan msg

// create a Termination channel type with blank interface
type termch chan interface{}

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

// Reads from Channel.
func (m MsgEnumType) ReadFromChannel() interface{} {
	return <-m.GetChannel()
}

// ___ End Global Channel Variables ___

// defining Color Enums
type Color int64

const (
	Red Color = iota
	Green
	Yellow
	Blue
	Purple
	Cyan
	Gray
	White
)

// Returns color as a string
func (c Color) Color() string {
	switch c {
	case Red:
		return "\033[31m"
	case Green:
		return "\033[32m"
	case Yellow:
		return "\033[33m"
	case Blue:
		return "\033[34m"
	case Purple:
		return "\033[35m"
	case Cyan:
		return "\033[36m"
	case Gray:
		return "\033[37m"
	case White:
		return "\033[97m"
	}
	return ""
}

type timestamp string

// Defining UID types.
// Node is any node that can send or receive messages, be it a client, server, or goroutine.
// Message is any message sent or received by a node.
type (
	UID   uint32
	NID   UID // Node ID
	MsgID UID // Message I
)

// Returns message type as string
func (c NID) IdType() string {
	return "node"
}

func (m MsgID) IdType() string {
	return "message"
}

// Define Enum for noClient UID
const (
	Global NID = 0
)

// Defines a bundle of channels used for a connections communications
type cnchanbundle struct {
	conn   ch     // channel used to communicate with connection
	ingest ch     // channel used to connect to channgels log ingestor
	term   termch // used to coordinate connectin/ingester shutdown routine.
}

// Connection object, represent a connection to the server.
type connection struct {
	messageHistory []msg        // Message History
	connectionId   NID          // Unique Identifier for connection
	conn           net.Conn     // connection objct
	chbundle       cnchanbundle // bundle of channels used for connections communications.
	startTime      time.Time    // Time of connection starting
}

// Defines interface needed for connection handler
type ConnectionHandler interface {
	ReadMsg() (msg, error)        // reads from connection, and returns a constructed message
	Write(msg) (n int, err error) // Writes to Connection handler Channel
	Close() error                 // Exposes net.Conn Close method
	LastMessage() msg             // Returns last message bundled in messageHistory
	AppendHistory(msg)            // Appends message to message history
	GetConnectionId() NID         // exposes ConnectionId
}

// initializes connection object
func initConnection(c *connection, conn net.Conn) {
	c.startTime = time.Now()
	c.chbundle.conn = make(chan msg, 20)
	c.chbundle.ingest = make(chan msg, 20)
	c.chbundle.term = make(chan interface{})
	c.conn = conn
	c.generateUid()
}

// Returns last message bundled in messageHistory
func (c connection) LastMessage() msg {
	return c.messageHistory[len(c.messageHistory)-1]
}

// reads from connection, and returns a constructed message
// Take []byte from connection, and creates a message object.
// uses initMsg to initialize message object, which sets
// message type, message id, and message timestamp.

func (c connection) ReadMsg() (msg, error) {
	var buf = make([]byte, buffersize)

	// route is a struct used to define a message route.
	route := struct{ source, destination NID }{
		source:      c.GetConnectionId(),
		destination: Global}
	n, err := c.conn.Read(buf)
	if err != nil {
		return msg{payload: buf}, err
	}
	m, err := InitMsg(buf[:n], Client, route)
	if err != nil {
		return m, err
	}
	c.AppendHistory(m)
	return m, nil
}

// Writes to Connection handler Channel
func (c *connection) Write(m msg) (int, error) {
	return c.conn.Write(m.GetPayload())
}

// Exposes net.Conn Close method
func (c *connection) Close() error {
	return c.conn.Close()
}

// Appends message to message history
func (c *connection) AppendHistory(m msg) {
	c.messageHistory = append(c.messageHistory, m)
}

// exposes ConnectionId
func (c connection) GetConnectionId() NID {
	if c.connectionId == 0 {
		return NID(0)
	}
	return c.connectionId
}

// generates a unique connection id
func (c *connection) generateUid() {
	c.connectionId = NID(uuid.New().ID())
}

// State is used to derive over-all state of connections
type state struct {
	connections []*connection
}

// defined state handler interface
type StateHandler interface {
	ActiveConnections() int     // Returns number of active connections
	RemoveConnection(NID)       // Removes connection from state 'connections
	AddConnection(*connection)  // Appends a new connection to state 'connections
	WriteMessage(message) error // Writes message to connections based on message destination
}

func (s *state) WriteMessage(m msg) error {
	// if message destination is global, we write to all connections
	if m.GetDestination() == Global {
		for _, c := range s.connections {
			_, err := c.Write(m)
			if err != nil {
				return err
			}
		}
	} else {
		for _, c := range s.connections {
			if c.GetConnectionId() == m.GetDestination() {
				_, err := c.Write(m)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *state) ActiveConnections() int {
	return len(s.connections)
}

func (s *state) RemoveConnection(cn NID) {
	for i, c := range s.connections {
		if c.connectionId == cn {
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
		}
	}
}

// Appends a new connection to state 'connections
func (s *state) AddConnection(c *connection) {
	s.connections = append(s.connections, c)
}

// Defines payload type
type payload []byte

// Payload method to return payload as a string.
func (p payload) String() string {
	return string(p)
}

// Message "object"
// individual message received from connection.
type msg struct {
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
	switch m.msgType {
	case Client:
		newPayload = payload(Green.Color() + string(m.payload) + reset)
	case Error:
		newPayload = payload(Red.Color() + string(m.payload) + reset)
	case System:
		newPayload = payload(Yellow.Color() + string(m.payload) + reset)
	}
	return string(newPayload)
}

// sets message type
func (m *msg) SetType(t MsgEnumType) {
	m.msgType = t
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
func InitMsg(b []byte, t MsgEnumType, route struct{ source, destination NID }) (msg, error) {
	// if destination is not 0, set destination to destination
	// If source is not provided, return an error.
	if route.source == 0 {
		return msg{}, fmt.Errorf("source not provided")
	}
	destination := Global
	if route.destination != 0 {
		destination = route.destination
	}

	m := msg{}
	m.SetSource(route.source)
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

// sets message payload
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

func main() {

	for _, v := range branding.Slicify() {
		fmt.Println(colorWrap(Blue, v))
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print(corgi)
	var wg sync.WaitGroup
	wg.Add(2) // adding two goroutines
	go func() {
		MessageBroker() // starting the Event Handler go routine
		wg.Done()       // decrementing the counter when done
	}()
	go func() {
		connListener(ip)
		wg.Done() // decrementing the counter when done
	}()
	wg.Wait() // waiting for all goroutines to finish
}

// Connection Listener accepts and passes connections off to Connection Handler
func connListener(ip string) error {
	// Create Listener bound to socket.
	listener, err := net.Listen(netp, net.JoinHostPort(ip, port))
	if err != nil {
		// if we fail to create listener, we log the error and exit. This is a fatal error.
		log.Fatalf("Failed to create listener: %q", err)
	}
	// defer closing of listener until we escape from connection handler.
	defer func() {
		m := msg{
			payload: []byte("Listener Closed"),
			msgType: System,
		}
		if err := listener.Close(); err != nil {
			m.payload = []byte(fmt.Sprintf("Failed to Close Listener %q", err.Error()))
			m.msgType = Error
		}
		m.msgType.WriteToChannel(m)
	}()
	// logs what socket the listener is bound to.
	System.WriteToChannel(msg{payload: []byte(fmt.Sprintf("Listener bound to %v", listener.Addr()))})
	// handles incoming connectons.
	for {
		// logs that we are waiting for a connection.
		System.WriteToChannel(msg{payload: []byte("Waiting for connection"), msgType: System})
		// routine will hang here until a connection is accepted.
		conn, err := listener.Accept()
		if err != nil {
			// if we fail to accept connection, we log the error and continue.
			Error.WriteToChannel(msg{payload: []byte(fmt.Sprintf("Failed to accept connection: %q", err.Error()))})
		}
		// initializing connection object
		newConn := connection{}
		initConnection(&newConn, conn)
		// Add connection directory to state. This is used to track active connections.
		currentstate.AddConnection(&newConn)
		// kicking off connection handler
		go connHandler(&newConn)
	}

}

// Connection Handler takes connections from listener, and processes read/writes
// TODO(jeanhaley): This function is a bit out of control, and needs to be refactored.
func connHandler(conn ConnectionHandler) {
	conn.Write(msg{payload: payload(branding.ColorString())}) // writes branding to connection
	System.WriteToChannel(msg{
		payload: []byte(fmt.Sprintf("New connection from %v", conn.GetConnectionId())),
		msgType: System,
	}) // logs start of new session

	// defering closing function until we escape from session handler.
	defer func() {
		System.WriteToChannel(msg{payload: []byte(fmt.Sprintf("Closing connection from %v", conn.GetConnectionId()))})
		// TODO(JeanHaley) Create a state handler(manager?) that can close this for us.
		// we should send a signal through an explicit connection channel to
		// the state handler that then tells it to close this connection and
		// pops it from the list of active connections.
		currentstate.RemoveConnection(conn.GetConnectionId())
		conn.Close()
	}()
	for {
		m, err := conn.ReadMsg() // read message from connection
		if err != nil {
			if err == io.EOF {
				System.WriteToChannel(msg{
					payload: []byte(fmt.Sprintf("Received EOF from %v .", conn.GetConnectionId())),
				})
				return
			} else {
				Error.WriteToChannel(msg{payload: []byte(err.Error())})
				return
			}
		}
		// Logs message received
		System.WriteToChannel(msg{
			payload: payload(fmt.Sprintf(
				"(%v)Received message: "+
					colorWrap(Purple, "%v"),
				conn.GetConnectionId(),
				string(m.GetPayload().String()))),
		})
		// Respond to message object
		switch {
		case m.GetPayload().String() == "corgi":
			m.SetPayload(payload(corgi))
			Client.WriteToChannel(m)
		case m.GetPayload().String() == "ping":
			m.SetPayload(payload("pong"))
			Client.WriteToChannel(m)
		case strings.Split(string(m.GetPayload().String()), ":")[0] == "ascii":
			m.SetPayload(
				payload(
					figure.NewColorFigure(
						// remove trailing newline
						strings.Split(m.GetPayload().String(), ":")[1][:len(strings.Split(m.GetPayload().String(), ":")[1])-1],
						"nancyj-fancy",
						"Green", true).ColorString())) // sets payload to ascii art
			Client.WriteToChannel(m) // write message to Client Channel
		}
	}
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
		case msg := <-clientChan:
			currentstate.WriteMessage(msg)
			logger.Printf("(%v)Received payload: %v", msg.GetSource(), msg.ColorWrap())
		case msg := <-sysChan:
			logger.Println(msg.ColorWrap())
		case msg := <-logChan:
			logger.Println(msg.ColorWrap())
		case <-time.After(time.Second * time.Duration(logerTime)):
			// Log a message that no errors have occurred for loggerTime seconds
			msg := msg{
				payload: []byte(fmt.Sprintf(
					"No errors for %v seconds, %v active connections",
					logerTime,
					currentstate.ActiveConnections())),
				msgType:     System,
				destination: Global,
			}
			logger.Println(msg.ColorWrap())
		}
	}
}

// colorWrap wraps a string in a color
func colorWrap(c Color, m string) string {
	const Reset = "\033[0m"
	return c.Color() + m + Reset
}
