# theVoid
 Scream into the Void. A simple TCP server that reflects all received traffic to everyone connected.
 - Handles incoming TCP connections. 
 - Wraps received []byte into messages
 - parses those messages, and updates a screen buffer. 
 - writes screen buffer to all clients with composed and personalized banner. 