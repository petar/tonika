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

// TODO: 
//   (*) Add flag for file descriptor limit, currently 240
//   (*) Add flag for maximum body length, currently 40K
//   (*) Count unique users
//   (*) Count online users
//   (*) Print incoming IP
//   (*) If we are writing sequentially it may not make sense to have
//       asynchronous request accepts. Perhaps have a few concurrent file writers?
//   (*) Put build number in request header, and save it in log file

var (
	flagBind     = flag.String("bind", ":49494", "Daemon bind address")
	flagFile     = flag.String("file", "", "Log files prefix")
	//flagKey      = flag.String("key", "", "Daemon private key")
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

type RawRecord struct {
	UnixSec int64
	Cargo   []byte	
}

func process(p []byte) {
	tm := time.LocalTime()
	k := inc()
	fmt.Printf("tonika-mond-req: #%d, %s, len=%d\n", k, tm.String(), len(p))
	unixsec := time.Seconds()
	rec := &RawRecord{
		UnixSec: unixsec,
		Cargo:   p,
	}
	err := fw.Encode(rec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-mond: Error recording: %s\n", err)
	}
}

func serve(q *http.Query) {
	req := q.GetRequest()
	if req.Body != nil {
		buf := make([]byte, 4096*10)  // Max 40K
		n,err := req.Body.Read(buf)
		req.Body.Close()
		if n < len(buf) && (err == nil || err == os.EOF) {
			process(buf[0:n])	
		}
	}
	asc := q.Hijack()
	conn,_,_ := asc.Close()
	if conn != nil {
		conn.Close()
	}
}

func main() {
	fmt.Fprintf(os.Stderr, "Starting " + sys.Name + " Monitor Daemon, Build " + sys.Build + "\n")
	flag.Parse()

	// Setup file writing facility
	if *flagFile == "" {
		fmt.Fprintf(os.Stderr, "tonika-mond: You must specify a file prefix\n")
		os.Exit(1)
	}
	fw_, err := filewriter.MakeFileWriter(*flagFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-mod: Error creating log file: %s\n", err)
		os.Exit(1)
	}
	fw = fw_

	// Start listening
	l,err := net.Listen("tcp", *flagBind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-mond: Error binding: %s\n", err)
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
