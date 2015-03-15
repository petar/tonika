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

var req = &Request{
	Method: "GET",
	RawURL: "http://www.google.com/",
	//URL: ,
	Proto:      "HTTP/1.1",
	ProtoMajor: 1,
	ProtoMinor: 1,
	//Header: map[string]string{},
	Close:     false,
	Host:      "www.google.com",
	Referer:   "",
	UserAgent: "Fake",
	Form:      map[string][]string{},
}

func TestAsyncClientConn(t *testing.T) {
	go sigh()
	c, err := net.Dial("tcp", "", "www.google.com:80")
	if err != nil {
		t.Fatalf("dial")
	}
	acc, err := NewAsyncClientConn(10e9, c)
	if err != nil {
		t.Fatalf("acc")
	}
	resp, err := acc.Fetch(req)
	if err != nil {
		t.Fatalf("fetch")
	}
	d, err := DumpResponse(resp, true)
	if err != nil {
		t.Fatalf("dump")
	}
	fmt.Printf("RESP:\n%s\n", string(d))
	acc.Close()
}
