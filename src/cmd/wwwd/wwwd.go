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


// Tonika Web Server

package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagBind = flag.String("http", ":80", "HTTP service address (e.g., ':80')")
	flagDir = flag.String("dir", "./", "Root directory of web server subtree")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: tonika-wwwd\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	fmt.Fprintf(os.Stderr, "tonika-wwwd: bind=%s, dir=%s\n", *flagBind, *flagDir)
	fmt.Fprintf(os.Stderr, "tonika-wwwd: Don't forget to set a high ulimit on this server!\n")

	s, err := MakeServer(*flagBind, *flagDir, 200)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-wwwd: Start server error: %s\n", err)
		os.Exit(2)
	}
	for {
		err := s.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tonika-wwwd: Accept error: %s\n", err)
			os.Exit(2)
		}
	}
}
