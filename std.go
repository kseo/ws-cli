package main

import (
	"fmt"
	"io"

	"github.com/chzyer/readline"
)

type InterruptibleStdin struct {
	mr        *MultiReader
	interrupt func()
}

func NewInterruptibleStdin(rs ...io.ReadCloser) *InterruptibleStdin {
	var mr MultiReader
	for _, r := range rs {
		mr.Add(r)
	}
	reader, writer := io.Pipe()
	mr.Add(reader)

	return &InterruptibleStdin{
		mr:        &mr,
		interrupt: func() { fmt.Fprintf(writer, "%c", readline.CharInterrupt) },
	}
}

func (i *InterruptibleStdin) Read(b []byte) (int, error) {
	return i.mr.Read(b)
}
