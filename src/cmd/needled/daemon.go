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
	"tonika/needle"
	"tonika/sys"
	// "tonika/util/filewriter"
)

var (
	flagNeedle = flag.String("needle", ":62077", "Bind address for needle UDP server")
	flagHttp   = flag.String("http", ":62070", "Bind address for HTTP API")
	flagLog    = flag.String("log", "needle-log", "Log files prefix")
)

func main() {
	fmt.Fprintf(os.Stderr, "Starting " + sys.Name + " Needle Daemon, Build " + sys.Build + "\n")
	flag.Parse()

	// Setup log writing facility
	if *flagLog == "" {
		fmt.Fprintf(os.Stderr, "tonika-needled: You must specify a log file prefix\n")
		os.Exit(1)
	}
	/*
	fw, err := filewriter.MakeFileWriter(*flagLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-needled: Error creating log file: %s\n", err)
		os.Exit(1)
	}
	*/

	// Resolve needle server address
	uaddr, err := net.ResolveUDPAddr(*flagNeedle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-needled: Error resolving needle bind address: %s\n", err)
		os.Exit(1)
	}

	// XXX: pass fw to needle server as logging facility

	// Start listening
	_, err = needle.MakeServer(uaddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-needled: Error starting: %s\n", err)
		os.Exit(1)
	}
	<-make(chan int)
}
