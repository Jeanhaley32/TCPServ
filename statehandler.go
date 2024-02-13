package main

// File: statehandler.go
import "fmt"

// State is used to derive over-all state of connections
// TODO(jeanhaley) For better performance, this should be a map of Connection ID to *connection.
// There is some work in making this change, as is we'd have to deal with some form of mutex handling
// for multiple sources tring to access this map at the same time. This may wont if we make the state
// handler it's own function that alone negotiates communications to each connection. This maybe stunts scalability,
// which is not really necessary for the scope of this project but may be a useful thing to consider in the spirit
// of good practice. (at least what I think that spirit may be)
//
// A better way of handling this would be to allow for multiple goroutines to be able to access a map, but either way
// The bottleneck becomes the map itself. Even if we have multiple go routines working on this connection state struct, we would
// still need to negotiate usage to allow for one at a time to read or modify it. So, maybe multi-threading this is not really helpful.
// unless we can divide the struct into slices, and hand them off to individual processes that work on their sliced portions.  Work can then
// be divided amongst those goroutines based on what slice they have authority over; This sounds complicated but potentially worth it.
//
// The next step for this though is to turn it into a map, and create a singleton goroutine that's job is to handle reads and modifications to
// this state struct.
//
// This could look like
// - broker reaches out to state handler with a connection ID it is looking to communicate with.
// - state handler grabs communication channel from state map.
// - state handler returns channel.
// - broker then sends it's intended message to that channel.
//
//   Considering this, there may be a few issues. We may need to add some form of marker, maybe a variable that takes on the valie of the NID for the job currently writing to that
//   Connection, this would also require that the node working with that connection send a communication back to release this channel. We'd also need to create an alternative plan in
//   case that node fails early and cannot communicate that it's finished. Maybe another process that can return a check on any specific node. This does sound like something a
//   "state handler" would do.

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
		RandomFacts := fmt.Sprintf("\"%v\"", SpecialMessage)
		newScreen := []byte(fmt.Sprintf("%v%v\n%v\n%v%v\n", clearScreen, coloredBranding, splash, ClientMessage, printWithBorder(RandomFacts)))
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
