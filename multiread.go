package main

import (
	"errors"
	"io"
	"sync"
)

var errClosedIO = errors.New("This I/O has been closed")

type multiReader struct {
	sync.RWMutex
	r       *io.PipeReader
	w       *io.PipeWriter
	sources []io.ReadCloser
	closed  bool
}

func (mr *multiReader) init() error {
	mr.RLock()
	if mr.closed {
		mr.RUnlock()
		return errClosedIO
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

func (mr *multiReader) Read(buf []byte) (int, error) {
	if err := mr.init(); err != nil {
		return -1, err
	}
	return mr.r.Read(buf)
}

func (mr *multiReader) Close() error {
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

func (mr *multiReader) Add(src io.ReadCloser) error {
	if err := mr.init(); err != nil {
		return err
	}
	mr.Lock()
	mr.sources = append(mr.sources, src)
	mr.Unlock()
	go io.Copy(mr.w, src)
	return nil
}
