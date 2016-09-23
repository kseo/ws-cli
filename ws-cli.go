package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/chzyer/readline"
	ws "github.com/gorilla/websocket"
)

var dialer = ws.Dialer{
	Subprotocols:    []string{"p1", "p2"},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func recv(conn *ws.Conn, rl *readline.Instance, wg *sync.WaitGroup) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			wg.Done()
			break
		}
		buf := new(bytes.Buffer)
		buf.WriteString("< ")
		buf.Write(p)
		buf.WriteRune('\n')

		rl.Stdout().Write(buf.Bytes())
	}
}

func send(conn *ws.Conn, rl *readline.Instance) {
	defer conn.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("ReadLine: %v", err)
		}

		// Remove the new line from the string.
		conn.WriteMessage(ws.TextMessage, []byte(line))
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
	fmt.Println("connected (press CTRL+C to quit)")

	rl, err := readline.New("> ")
	if err != nil {
		log.Fatalf("New: %v", err)
	}
	defer rl.Close()

	go recv(conn, rl, &wg)
	send(conn, rl)

	wg.Wait()
}
