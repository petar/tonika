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
	"net"
	"strings"
	"tonika/http"
	"tonika/sys"
)

func askForGreenLight() bool {
	url,err := http.ParseURL(sys.TangraServerURL)
	if err != nil {
		panic("tangra, bad URL")
	}
	req := &http.Request{
		Method: "GET",
		URL: url,
		Close: true,
		UserAgent: sys.Name + "-Client-Tangra",
	}
	req.Header = make(map[string]string)
	req.Header["Tonika-Build"] = sys.Build
	req.Header["Tonika-Released"] = sys.Released

	// Connect
	conn,err := net.Dial("tcp", "", url.Host)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return true
	}
	cc := http.NewClientConn(conn,nil)

	// Send request
	err = cc.Write(req)
	if err != nil {
		cc.Close()
		conn.Close()
		return true
	}
	resp,err := cc.Read()
	if err != nil {
		cc.Close()
		conn.Close()
		return true
	}
	cc.Close()
	conn.Close()

	if resp.Header == nil {
		return true
	}
	green, ok := resp.Header["Green"]
	if !ok {
		return true
	}
	green = strings.TrimSpace(green)
	if green == "Halt" {
		return false
	}
	return true
}
