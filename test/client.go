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
