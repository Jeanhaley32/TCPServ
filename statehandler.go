package main

// File: statehandler.go
import "fmt"

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
