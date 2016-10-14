package main

import (
	"fmt"
	"io"

	"github.com/chzyer/readline"
)

type interruptibleStdin struct {
	mr        *multiReader
	interrupt func()
}

func newInterruptibleStdin(rs ...io.ReadCloser) *interruptibleStdin {
	var mr multiReader
	for _, r := range rs {
		mr.Add(r)
	}
	reader, writer := io.Pipe()
	mr.Add(reader)

	return &interruptibleStdin{
		mr:        &mr,
		interrupt: func() { fmt.Fprintf(writer, "%c", readline.CharInterrupt) },
	}
}

func (i *interruptibleStdin) Read(b []byte) (int, error) {
	return i.mr.Read(b)
}
