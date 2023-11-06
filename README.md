What is this?
 At it's simplest, and most utilitarian. TCP Serve acts as a very feature incomplete chat service. Multiple peope can connect to it using 
 the accompanying client application, or a simple netcat command to the IP and port. It takes in raw TCP packets as bytes, and reflects 
 them to every other connection already connected to it.

 
 TCP Serve is a project I worked at during NYUs ITP Camp in 2023. 
 The purpose of this project was to learn some entry level server programming with go, create an application that
 would have me utilize both GoRoutines and Channels from a approach that attempted to put a microservice concepts into practice.
 Although this program itself i monolithic, I attempted to approach the concept of parralel running goroutines as if they were seperate
 services communicating, and depending on each other and work together to produce a single objective. 
 
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
