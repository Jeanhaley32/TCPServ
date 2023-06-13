// For this to work, we need to spin off the read element of this program into a goroutine.
// otherwise, the program will only read X amount of lines. There may be a loop way of doing this
// checking for EOF. But I'm not sure how to do that yet. I have an idea, but i'm lazy.
// i'll do it later.
package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	stringTest := []byte("hello world")
	asciiTest := []byte("ascii:test")
	corgiTest := []byte("corgi")
	_, err = conn.Write(stringTest)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Println("string response:")
	fmt.Printf("%s \n", response)

	_, err = conn.Write(asciiTest)
	if err != nil {
		panic(err)
	}
	response, err = reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Println("ascii response:")
	fmt.Printf("%s \n", response)

	_, err = conn.Write(corgiTest)
	if err != nil {
		panic(err)
	}
	response, err = reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	fmt.Println("corgi response:")
	fmt.Printf("%v \n", response)
}
