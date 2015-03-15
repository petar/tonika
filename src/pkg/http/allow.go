// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net"
	"strings"
)

type AllowHosts struct {
	hosts map[string]int
}

// s is a comma-separated list of host names.
// TODO: If the allow list is long, or there is no network connectivity,
// this function may take too long.
func MakeAllowHosts(s string) *AllowHosts {
	ah := &AllowHosts{ make(map[string]int) }
	parts := strings.Split(s, ",", -1)
	for _,p := range parts {
		ip := net.ParseIP(p)
		var ips []string
		if ip == nil {
			_, addrs, err := net.LookupHost(p)
			if err != nil {
				continue
			}
			ips = addrs
		} else {
			ips = make([]string, 1)
			ips[0] = ip.String()
		}
		for _,p := range ips {
			ah.hosts[p] = 1
		}
	}
	return ah
}

// NOTE: IsAllowedAddr currently allows only TCP addresses.
func (ah *AllowHosts) IsAllowedAddr(addr net.Addr) bool {
	tcpaddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return false
	}
	_,ok = ah.hosts[tcpaddr.IP.String()]
	return ok
}
