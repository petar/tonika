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


package fe

import (
	"strings"
	"tonika/http"
	"tonika/sys"
	"tonika/util/misc"
)

const (
	feWWWReq    = iota
	feAdminReq  = iota
	feTonikaReq = iota
	feBadReq    = iota
)

func getRequestType(req *http.Request) (int, sys.Id) {
	hostparts := strings.Split(req.Host, ".", -1)
	hostparts = misc.ReverseHost(hostparts)

	if len(hostparts) < 2 {
		return feWWWReq, 0
	}
	if hostparts[0] != "org" || hostparts[1] != "5ttt" {
		return feWWWReq, 0
	}
	hostparts = hostparts[2:]
	if len(hostparts) == 0 || hostparts[0] == "www" { // 5ttt.org, www.5ttt.org
		return feWWWReq, 0
	}
	if hostparts[0] == "a" {
		return feAdminReq, 0
	}
	id,err := sys.ParseId(hostparts[0])
	if err != nil {
		return feWWWReq, 0
	}
	hostparts = hostparts[1:]
	if len(hostparts) == 0 { // xxxxxxxxx.5ttt.org
		return feTonikaReq, id
	}
	if hostparts[0] == "a" { // a.xxxxxxxxx.5ttt.org
		return feAdminReq, id 
	}
	return feBadReq, 0
}
