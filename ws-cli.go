package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/chzyer/readline"
	ws "github.com/gorilla/websocket"
)

func recv(conn *ws.Conn, rl *readline.Instance, wg *sync.WaitGroup, interrupt func()) {
	defer func() {
		interrupt()
		wg.Done()
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Fprintf(rl.Stdout(), "< %s\n", string(p))
	}
}

func send(conn *ws.Conn, rl *readline.Instance, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		wg.Done()
	}()

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

func dial(url string, origin string, subprotocol string) (*ws.Conn, error) {
	var subprotocols []string
	var header http.Header

	if subprotocol != "" {
		subprotocols = []string{subprotocol}
	}
	if origin != "" {
		header = http.Header{"Origin": {origin}}
	}

	dialer := ws.Dialer{
		Subprotocols:    subprotocols,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, _, err := dialer.Dial(url, header)
	return conn, err
}

func main() {
	var url = flag.String("url", "", "url")
	var origin = flag.String("origin", "", "optional origin")
	var subprotocol = flag.String("subprotocol", "", "optional subprotocol")
	flag.Parse()
	if *url == "" {
		flag.Usage()
		return
	}

	conn, err := dial(*url, *origin, *subprotocol)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}
	fmt.Println("connected (press CTRL+C to quit)")

	reader, writer := io.Pipe()
	interrupt := func() {
		fmt.Fprintf(writer, "%c", readline.CharInterrupt)
	}

	var mr MultiReader
	mr.Add(os.Stdin)
	mr.Add(reader)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
		Stdin:  &mr,
	})
	if err != nil {
		log.Fatalf("New: %v", err)
	}
	defer rl.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go recv(conn, rl, &wg, interrupt)
	go send(conn, rl, &wg)

	wg.Wait()

	fmt.Println("Disconnected")
}
