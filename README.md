What is this?\
 At it's simplest, and most utilitarian, TCP Serve acts as a very feature incomplete chat service. Multiple clients can connect to it 
 using either the accompanying client application, or a simple netcat command to the IP and port. It takes in raw TCP packets as bytes,
 and reflects them to every other client with an active connection.


 TCP Serve is a project I worked at during NYUs ITP Camp in 2023. 
 The purpose of this project was to learn some entry level server programming with Go and create an application that
 would have me utilize both GoRoutines and Channels from a micro-services centric approach.
 Although this program itself is monolithic, I attempted to approach the concept of parralel running goroutines as if they were seperate
 services communicating with and depending on each other, working together to accomplish a single objective. 
 
# TCPServ
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

## What the Client sees (NetCat connection)
![](client.gif)
