package main

// Description: Defines connection handler, and connection listener.

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
)

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

// Defines a bundle of channels used for a connections communications
type cnchanbundle struct {
	conn   ch     // channel used to communicate with connection
	ingest ch     // channel used to connect to server log ingestor
	term   termch // used to coordinate connection/ingester shutdown routine.
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

// Connection Listener accepts and passes connections off to Connection Handler
func connListener(ip string) error {
	// Create Listener bound to socket.
	listener, err := net.Listen(netp, socket)
	if err != nil {
		// if we fail to create listener, we log the error and exit. This is a fatal error.
		log.Fatalf("%v: Failed to create listener: %q", connlistener.Name(), err)
	}
	// defer closing of listener until we escape from connection handler.
	defer func() {
		System.LogPayload(connlistener, "Closing Listener")
		if err := listener.Close(); err != nil {
			Error.LogError(connlistener, fmt.Sprintf("Failed to Close Listener %q", err.Error()))
		}
	}()
	// logs what socket the listener is bound to.
	System.LogPayload(connlistener, fmt.Sprintf("Listener bound to %v", listener.Addr()))
	// handles incoming connectons.
	for {
		// logs that we are waiting for a connection.
		System.LogPayload(connlistener, "Waiting for connection")
		// routine will hang here until a connection is accepted.
		conn, err := listener.Accept()
		if err != nil {
			// if we fail to accept connection, we log the error and continue.
			Error.LogError(connlistener, fmt.Sprintf("Failed to accept connection: %q", err.Error()))
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

// Connection Handler takes connections from listener, and processes read/writes
// TODO(jeanhaley): This function is a bit out of control, and needs to be refactored.
func connHandler(conn ConnectionHandler) {
	// Log new Connection.
	System.LogPayload(connhandler, fmt.Sprintf("New connection from %v", conn.GetConnectionId()))
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
		System.LogPayload(connhandler, fmt.Sprintf("Closing connection from %v", conn.GetConnectionId()))
		currentstate.RemoveConnection(conn.GetConnectionId())
		conn.Close()
	}()
	for {
		m, err := conn.ReadMsg() // read message from connection
		if err != nil {
			if err == io.EOF {
				System.LogPayload(connhandler, fmt.Sprintf("Received EOF from %v .", conn.GetConnectionId()))
				return
			} else {
				Error.LogError(connhandler, err.Error())
				return
			}
		}
		// Log message received
		System.LogPayload(connhandler, fmt.Sprintf("(%v)Received message:"+colorWrap(Purple, "%v"), conn.GetConnectionId(), string(m.GetPayload().String())))
		// Catch trigger words, and handle each one differently.
		switch {
		case HasString(m.GetPayload().String(), ":"):
			newPayload := parseAction(m)
			m, err := InitMsg(newPayload, Client, route{source: Global, destination: Global}, conn.GetConnColor())
			if err != nil {
				Error.Type().LogError(connhandler, err.Error())
				m.payload = []byte(err.Error())
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
