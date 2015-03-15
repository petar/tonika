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


// hhd: HTTP header dumper

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	httpaddr = flag.String("http", ":4949", "http listen address (e.g., ':4949')")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hhd\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func do(c net.Conn) {
	br := bufio.NewReader(c)
	clrf := 0
	for {
		b,err := br.ReadByte()
		if err != nil {
			//fmt.Printf("==== closed ====\n")
			break
		}
		fmt.Print(string(b))
		if b == '\r' || b == '\n' {
			clrf++
		} else {
			clrf = 0
		}
		if clrf == 4 {
			//fmt.Printf("==== *** ====\n")
			break
		}
	}
	c.Close()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	l,err := net.Listen("tcp", *httpaddr)
	if err != nil {
		log.Stderrf("error listening: %s", err)
		os.Exit(1)
	}

	for {
		c,err := l.Accept()
		if err != nil {
			log.Stderrf("error accepting: %s", err)
			os.Exit(1)
		}
		do(c)
	}
}
