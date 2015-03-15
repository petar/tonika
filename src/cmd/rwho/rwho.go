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


// Dumps the address of the remote trying to connect here

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

var (
	flagAddr = flag.String("addr", ":80", "Address to listen to")
)

func main() {
	flag.Parse()

	l,err := net.Listen("tcp", *flagAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listening: %s", err)
		os.Exit(1)
	}

	for {
		c,err := l.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accepting: %s", err)
			os.Exit(1)
		}
		fmt.Printf("Connection from: %s\n", c.RemoteAddr())
		c.Close()
	}
}
