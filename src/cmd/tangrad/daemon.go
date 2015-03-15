// Tonika: A distributed social networking platform
// Copyright (C) 2010 Petar Maymounkov <petar@5ttt.org>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.


package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
	"tonika/http"
	"tonika/sys"
	"tonika/util/filewriter"
)

var (
	flagBind     = flag.String("bind", ":37373", "Daemon bind address")
	flagLog      = flag.String("log", "tangralog", "Log files prefix")
)

var (
	fw    *filewriter.FileWriter
	pings int64
	lk    sync.Mutex
)

func inc() int64 {
	lk.Lock()
	defer lk.Unlock()
	pings++
	return pings
}

func process(req *http.Request, raddr net.Addr) *http.Response {
	tm := time.LocalTime()
	k := inc()
	
	build := "?.?.?.?"
	if req.Header != nil {
		b,ok := req.Header["Tonika-Build"]
		if ok {
			build = b
		}
	}
	green := "OK"
	fmt.Printf("tonika-tangrad-req: #%d, Time: %s, Build: %s, Green: %s\n", 
		k, tm.String(), build, green)

	resp := &http.Response{
		Status: "OK",
		StatusCode: 200,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RequestMethod: "GET",
		Close: true,
	}
	resp.Header = make(map[string]string)
	resp.Header["Green"] = green

	var w bytes.Buffer
	fmt.Fprintf(&w, "Time: %s, From: %s, Build: %s, Green: %s\n",
		time.LocalTime().Format(time.UnixDate), raddr.String(), build, green)
	_,err := fw.Write(w.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-tangrad: Error writing log: %s\n", err)
	}

	return resp
}

func serve(q *http.Query) {
	req := q.GetRequest()
	if req.Body != nil {
		req.Body.Close()
	}
	resp := process(req, q.RemoteAddr())
	asc := q.Hijack()
	asc.Write(req, resp)
	conn,_,_ := asc.Close()
	if conn != nil {
		conn.Close()
	}
}

func main() {
	fmt.Fprintf(os.Stderr, "Starting " + sys.Name + " Tangra Daemon, Build " + sys.Build + "\n")
	flag.Parse()

	// Setup log writing facility
	if *flagLog == "" {
		fmt.Fprintf(os.Stderr, "tonika-tangrad: You must specify a log file prefix\n")
		os.Exit(1)
	}
	fw_, err := filewriter.MakeFileWriter(*flagLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-tangrad: Error creating log file: %s\n", err)
		os.Exit(1)
	}
	fw = fw_

	// Start listening
	l,err := net.Listen("tcp", *flagBind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-tangrad: Error binding: %s\n", err)
		os.Exit(1)
	}
	s := http.NewAsyncServer(l,10e9,240)
	for {
		q,err := s.Read()
		if err == nil {
			go serve(q)
		}
	}
}
