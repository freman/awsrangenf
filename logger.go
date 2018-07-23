package main

import (
	"container/ring"
	"os"
)

type ringWriter struct {
	*ring.Ring
}

func (w *ringWriter) Write(b []byte) (int, error) {
	if w.Ring == nil {
		w.Ring = ring.New(100)
	}
	w.Ring.Value = string(b)
	w.Ring = w.Ring.Next()

	return os.Stdout.Write(b)
}
