package main

// TODO(jeanhaley) - The following items need to be addressed
// 0. Generate a random color for each user access the Void.
// 1. Create a logging system, that will handle all logging for this server.
// 2. maybe breakup message broker into a log handler? also, why are we distinguishing between log and system?
//    Maybe they should be treated the same?
// 2. Create a routine that handles global messages, and sends a screen state to all clients.
// 		1. How about we create a unique banne for each user, that pulls from unique user stats, and
// 			 references certain global variables.
// 		2. Add cat-facts into the banner.
// 3. Create a system to handle the state of the server, and the state of the clients.
// 4. Create a better method of wrapping errors in error messages, to make routing them easier.
import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
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
	clearScreen = "\033[H\033[2J"
)

// ___ Global Variables ___

var (
	ip, netp, port, banner, banmsgfp, SpecialMessage                                  string
	buffersize, logerTime, clientchannelbuffer, logchannelbuffer, systemchannelbuffer int
	ClientMessageCount                                                                int // Sets limit for messages show to client
	clientChan, logChan, sysChan                                                      ch  // Global Channels
	currentstate                                                                      state
	globalState                                                                       []msg
	banmessages                                                                       []string // holds banner messages
	ServerStartTime                                                                   time.Time
)

var (
	// creating a blank global branding variable.
	// this needs to be done, because the type figure.figure is not exported.
	branding = figure.NewColorFigure("", "nancyj-fancy", "Blue", true)
)

// uses init function to set set up global flag variables, and channels.
func init() {
	ServerStartTime = time.Now()
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
	flag.IntVar(&ClientMessageCount, "ClientMessageCount", 20, "Number of messages to show to client")
	flag.StringVar(&banmsgfp, "banmsgfp", "msg.txt", "Banner Message File Path")
	flag.Parse()

	globalState = make([]msg, 0, ClientMessageCount)
	// instantiating global channels.
	clientChan = make(chan msg, clientchannelbuffer)
	logChan = make(chan msg, logchannelbuffer)
	sysChan = make(chan msg, systemchannelbuffer)
	branding = figure.NewColorFigure(banner, "nancyj-fancy", "Blue", true) // sets banner to value passed by terminal flags.

}

// ___ Global Channel Variables ___
// create a channel type with blank interface
type ch chan msg

// define route type.
type route struct {
	source      NID
	destination NID
}

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
	Black
	LightGray
	DarkGray
	LightRed
	LightGreen
	LightYellow
	LightBlue
	LightMagenta
	LightCyan
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
	case Black:
		return "\033[30m"
	case LightGray:
		return "\033[37m"
	case DarkGray:
		return "\033[90m"
	case LightRed:
		return "\033[91m"
	case LightGreen:
		return "\033[92m"
	case LightYellow:
		return "\033[93m"
	case LightBlue:
		return "\033[94m"
	case LightMagenta:
		return "\033[95m"
	case LightCyan:
		return "\033[96m"
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
	connColor      Color        // color used to represent connection
	messageHistory []msg        // Message History
	connectionId   NID          // Unique Identifier for connection
	conn           net.Conn     // connection objct
	chbundle       cnchanbundle // bundle of channels used for connections communications.
	startTime      time.Time    // Time of connection starting
}

// Defines interface needed for connection handler
type ConnectionHandler interface {
	GetSplashScreen() string      // returns connections unique splash screen
	ChooseConnColor()             // Chooses a color for connection
	GetConnColor() Color          // returns connection color
	ReadMsg() (msg, error)        // reads from connection, and returns a constructed message
	Write(msg) (n int, err error) // Writes to Connection handler Channel
	Close() error                 // Exposes net.Conn Close method
	LastMessage() msg             // Returns last message bundled in messageHistory
	AppendHistory(msg)            // Appends message to message history
	GetConnectionId() NID         // Gets ConnectionId
	GetStartTime() time.Time      // Gets startTime
}

// Gets startTime
func (c connection) GetStartTime() time.Time {
	return c.startTime
}

// initializes connection object
func initConnection(c *connection, conn net.Conn) {
	c.startTime = time.Now()
	c.chbundle.conn = make(chan msg, 20)
	c.chbundle.ingest = make(chan msg, 20)
	c.chbundle.term = make(chan interface{})
	c.ChooseConnColor()
	c.conn = conn
	c.generateUid()
}

// returns a unique splash screen for the connection.
func (c connection) GetSplashScreen() string {
	r := time.Since(c.startTime)
	h := int(r.Hours())
	m := int(r.Minutes()) % 60
	s := int(r.Seconds()) % 60
	connectionTime := fmt.Sprintf("Session Length: %02d:%02d:%02d", h, m, s)
	connectionId := fmt.Sprintf("ConnID:%v", c.GetConnectionId())
	return colorWrap(Purple, fmt.Sprintf("%v\t\t\t\t%v\n", connectionId, connectionTime))
}

// randomly chooses a color to represent this connection
func (c *connection) ChooseConnColor() {
	rand.NewSource(time.Now().UnixNano())
	// this is used to color code messages from this connection.
	c.connColor = Color(rand.Intn(17))
}

// returns connections color
func (c connection) GetConnColor() Color {
	return c.connColor
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
	cutnewline := len([]byte("\n"))
	// Removes trailing newline from message.
	m, err := InitMsg(buf[:n-cutnewline], Client, route, c.GetConnColor())
	if err != nil {
		return m, err
	}
	c.AppendHistory(m)
	return m, nil
}

// Writes to Connection handler Channel
func (c *connection) Write(m msg) (int, error) {
	return c.conn.Write([]byte(m.ColorWrap()))
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
	WriteScreen(message) error  // Updates all connections with new screen
	WriteMessage(message) error // Writes message to connections based on message destination
}

// Updates all connections with new screen
func (s *state) WriteScreen() {
	coloredBranding := colorWrap(Red, branding.ColorString())
	splash := splashScreen()
	for _, c := range s.connections {
		ClientMessage := c.GetSplashScreen()
		// RandomFacts := fmt.Sprintf("\"%v\"", SpecialMessage)
		newScreen := []byte(fmt.Sprintf("%v%v\n%v\n%v\n", clearScreen, coloredBranding, splash, ClientMessage))
		// newScreen := []byte(fmt.Sprintf("%v%v\n%v\n%v%v\n", clearScreen, coloredBranding, splash, ClientMessage, printWithBorder(RandomFacts)))

		for _, m := range globalState {
			newScreen = append(newScreen, m.GetPayload()...)
		}
		c.Write(msg{payload: payload(newScreen), destination: Global})
	}
}

// Writes message to connections based on message destination
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
	// go func() {
	// 	FactGenerator()
	// 	wg.Done()
	// }()
	wg.Wait() // waiting for all goroutines to finish
}

// Connection Listener accepts and passes connections off to Connection Handler
func connListener(ip string) error {
	// Create Listener bound to socket.
	listener, err := net.Listen(netp, net.JoinHostPort(ip, port))
	if err != nil {
		// if we fail to create listener, we log the error and exit. This is a fatal error.
		log.Fatalf("ConnListener: Failed to create listener: %q", err)
	}
	// defer closing of listener until we escape from connection handler.
	defer func() {
		m := msg{
			payload: []byte("Listener Closed"),
			msgType: System,
		}
		if err := listener.Close(); err != nil {
			m.payload = []byte(fmt.Sprintf("ConnListener: Failed to Close Listener %q", err.Error()))
			m.msgType = Error
		}
		m.msgType.WriteToChannel(m)
	}()
	// logs what socket the listener is bound to.
	System.WriteToChannel(msg{payload: []byte(fmt.Sprintf("ConnListener: Listener bound to %v", listener.Addr()))})
	// handles incoming connectons.
	for {
		// logs that we are waiting for a connection.
		System.WriteToChannel(msg{payload: []byte("ConnListener: Waiting for connection"), msgType: System})
		// routine will hang here until a connection is accepted.
		conn, err := listener.Accept()
		if err != nil {
			// if we fail to accept connection, we log the error and continue.
			Error.WriteToChannel(msg{payload: []byte(fmt.Sprintf("ConnListener: Failed to accept connection: %q", err.Error()))})
			continue
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
	// Log new Connection.
	System.WriteToChannel(msg{
		payload: []byte(fmt.Sprintf("New connection from %v", conn.GetConnectionId())),
		msgType: System,
	})
	// Send message to clearn new connection screen.
	conn.Write(
		msg{
			payload: payload(clearScreen),
		})
	// Display server branding message.
	conn.Write(
		msg{
			payload: payload(branding.ColorString() +
				"\n" + splashScreen()),
		}) // writes branding to connection

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
		// Log message received
		System.WriteToChannel(msg{
			payload: payload(fmt.Sprintf(
				"(%v)Received message: "+
					colorWrap(Purple, "%v"),
				conn.GetConnectionId(),
				string(m.GetPayload().String()))),
			msgType: System,
		})
		// Catch trigger words, and handle each one differently.
		switch {
		case HasString(m.GetPayload().String(), ":"):
			newPayload := parseAction(m)
			m, err := InitMsg(newPayload, Client, route{source: Global, destination: Global}, conn.GetConnColor())
			if err != nil {
				Error.WriteToChannel(msg{payload: []byte(err.Error())})
				return
			}
			Client.WriteToChannel(m)
			continue
		case HasString(m.GetPayload().String(), "corgi"):
			Client.WriteToChannel(msg{
				payload: payload(corgi),
				msgType: System,
			})
		}
		m.SetPayload(
			payload(
				fmt.Sprintf("(%v) %v",
					conn.GetConnectionId(),
					string(m.ColorWrap()))))
		Client.WriteToChannel(m) // write message to Client Channel
	}
}

// returns payload based on action preceeding ":"
func parseAction(m msg) payload {
	switch strings.Split(m.GetPayload().String(), ":")[0] {
	case "ascii":
		return payload(figure.NewColorFigure(strings.Split(m.GetPayload().String(), ":")[1], "nancyj-fancy", "Green", true).ColorString())
	default:
		return payload(fmt.Sprintf("Invalid action: %v", strings.Split(m.GetPayload().String(), ":")[0]))
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

//

// Returns Splash Screen elements.
func splashScreen() string {
	welcome := "Welcome to the Void!"
	activeconn := colorWrap(
		Green, fmt.Sprintf(
			"There are currently %v active connections.", currentstate.ActiveConnections()))
	directions := colorWrap(Purple, "Type 'ascii:' before your message to display ascii art")
	splashmessage := fmt.Sprintf("\t\t%v\n\t  %v\n  %v\v", welcome, activeconn, directions)
	return splashmessage
}

// colorWrap wraps a string in a color
func colorWrap(c Color, m string) string {
	const Reset = "\033[0m"
	return c.Color() + m + Reset
}

// HasString returns true if a string contains another string
func HasString(str, match string) bool {
	bool, _ := regexp.MatchString(match, str)
	return bool
}

// // Sets Strings that are attached to the Voids Banner.
// func FactGenerator() {
// 	// Switch statement that takes in the current time and performs actions based on the time.
// 	var BannerMessages []string
// 	defer func() {
// 		msg := msg{
// 			payload: []byte("Fact Generator Closed"),
// 			msgType: System,
// 		}
// 		msg.msgType.WriteToChannel(msg)
// 	}()
// 	FileContents, err := ioutil.ReadFile(banmsgfp)
// 	if err != nil {
// 		Error.WriteToChannel(msg{payload: []byte(err.Error())})
// 	} else {
// 		BannerMessages = strings.Split(string(FileContents), "\n")
// 		SpecialMessage = BannerMessages[rand.Intn(len(BannerMessages))]
// 	}
// 	for {
// 		time.AfterFunc(1*time.Minute, func() {
// 			FileContents, err = ioutil.ReadFile(banmsgfp)
// 			if err != nil {
// 				Error.WriteToChannel(msg{payload: []byte(err.Error())})
// 				return
// 			}
// 			BannerMessages = strings.Split(string(FileContents), "\n")
// 		})

// 		time.AfterFunc(3*time.Minute, func() {
// 			SpecialMessage = BannerMessages[rand.Intn(len(BannerMessages))]
// 		})
// 	}
// }

// func printWithBorder(text string) string {
// 	horizontalBorder := "+" + strings.Repeat("-", len(text)+2) + "+"
// 	return fmt.Sprintf("%v\n| %v |\n%v", horizontalBorder, text, horizontalBorder)
// }
