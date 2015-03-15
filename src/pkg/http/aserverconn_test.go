// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net"
	"os/signal"
	"testing"
)

func sigh() {
	for {
		<-signal.Incoming
		panic()
	}
}

func serveLoop(asc *AsyncServerConn) {
	for {
		req, err := asc.Read()
		if err != nil {
			c, _, _ := asc.Close()
			if c != nil {
				c.Close()
			}
			fmt.Printf("r-closed conn\n")
			return
		}
		fmt.Printf("req\n")
		if req.Body != nil {
			req.Body.Close()
		}
		err = asc.Write(req, newRespServiceUnavailable())
		if err != nil {
			c, _, _ := asc.Close()
			if c != nil {
				c.Close()
			}
			fmt.Printf("w-closed conn\n")
			return
		}
	}
}

func TestAsyncServerConn(t *testing.T) {
	go sigh()
	l, err := net.Listen("tcp", ":4949")
	if err != nil {
		t.Fatalf("listen")
	}
	for {
		c, err := l.Accept()
		if err != nil {
			t.Fatalf("accept")
		}
		fmt.Printf("accepted\n")
		asc, err := NewAsyncServerConn(10e9, c)
		if err != nil {
			t.Fatalf("asc")
		}
		go serveLoop(asc)
	}
}
