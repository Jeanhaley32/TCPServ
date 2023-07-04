# theVoid
 Scream into the Void. A simple TCP server that reflects all received traffic to everyone connected.
 - Handles incoming TCP connections. 
 - Wraps received []byte into messages
 - parses those messages, and updates a screen buffer. 
 - writes screen buffer to all clients with composed and personalized banner.
 - Runs a TimeKeeper Routine that performs actions at specific intervals of time.
    - ATM, ingests msg.txt every 5 seconds, and update a list of messages.
    - ever 5 minutes, it uses a random int to pick a message to add to the global banner. 

## Example Server Side Logging
![](out.gif)
