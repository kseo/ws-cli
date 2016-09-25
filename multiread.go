package main

import (
	"errors"
	"io"
	"sync"
)

var ErrClosedIO = errors.New("This I/O has been closed")

type MultiReader struct {
	sync.RWMutex
	r       *io.PipeReader
	w       *io.PipeWriter
	sources []io.ReadCloser
	closed  bool
}

func (mr *MultiReader) init() error {
	mr.RLock()
	if mr.closed {
		mr.RUnlock()
		return ErrClosedIO
	}
	init := false
	if mr.r == nil || mr.w == nil {
		init = true
	}
	mr.RUnlock()
	if init {
		mr.Lock()
		mr.r, mr.w = io.Pipe()
		mr.sources = []io.ReadCloser{}
		mr.Unlock()
	}
	return nil
}

func (mr *MultiReader) Read(buf []byte) (int, error) {
	if err := mr.init(); err != nil {
		return -1, err
	}
	return mr.r.Read(buf)
}

func (mr *MultiReader) Close() error {
	mr.Lock()
	defer mr.Unlock()

	var err error

	if e1 := mr.w.Close(); e1 != nil {
		err = e1
	}
	for _, src := range append(mr.sources, mr.r) {
		if e1 := src.Close(); e1 != nil {
			err = e1
		}
	}

	mr.r, mr.w = nil, nil
	mr.closed = true
	return err
}

func (mr *MultiReader) Add(src io.ReadCloser) error {
	if err := mr.init(); err != nil {
		return err
	}
	mr.Lock()
	mr.sources = append(mr.sources, src)
	mr.Unlock()
	go io.Copy(mr.w, src)
	return nil
}
