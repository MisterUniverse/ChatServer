package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string // an outgoing message channel

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // all incoming client messages
)

func broadcast() {
	clients := make(map[client]bool) // all connected clients

	for {
		select {
		case msg := <-messages:
			// Broadcast the message to all connected clients
			// clients outgoin message channel
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // Need to handle network errors
	}
}

func handleConnection(conn net.Conn) {
	ch := make(chan string) // outgoing client message
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ": " + input.Text()
	}
	// Need to handle errors from input.Err()

	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func main() {
	listener, err := net.Listen("tcp", "localhost:6969")
	if err != nil {
		log.Fatal(err)
	}

	go broadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go handleConnection(conn)
	}
}
