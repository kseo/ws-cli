package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	ws "github.com/gorilla/websocket"
)

var dialer = ws.Dialer{
	Subprotocols:    []string{"p1", "p2"},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func recv(conn *ws.Conn, wg *sync.WaitGroup) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			wg.Done()
			break
		}
		var message = string(p)
		fmt.Printf("< %s\n", message)
	}
}

func send(conn *ws.Conn) {
	reader := bufio.NewReader(os.Stdin)
	defer conn.Close()

	for {
		fmt.Print("> ")
		p, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("ReadBytes: %v", err)
		}
		// Remove the new line from the string.
		conn.WriteMessage(ws.TextMessage, (p[0 : len(p)-1]))
	}
}

func main() {
	var url = flag.String("url", "", "url")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(1)

	conn, _, err := dialer.Dial(*url, nil)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}

	go recv(conn, &wg)
	send(conn)

	wg.Wait()
}
