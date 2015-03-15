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


// Chan interfaces to various net interfaces
package netch

import (
	"log"
	"net"
	"os"
	"tonika/http"
)

// Channel returns net.Conn or os.Error. It receives anything to mean close.
func Listen(addr string, fdlim *http.FDLimiter) (chan interface{}, os.Error) {
	l,err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	cc := make(chan interface{})
	go func() {
		for {
			_,sig := <-cc
			if sig || closed(cc) {
				l.Close()
				return
			}
			if fdlim == nil || fdlim.LockOrTimeout(10e9) == nil {
				c,err := l.Accept()
				if err != nil {
					if c != nil {
						c.Close()
					}
					if fdlim != nil {
						fdlim.Unlock()
					}
					l.Close()
					cc<- err
					return
				}
				if fdlim != nil {
					cc <- http.NewConnRunOnClose(c, func(){ fdlim.Unlock() })
				} else {
					cc <- c
				}
			} else {
				log.Stderrf("accept, fd starvation\n")
			}
		} // for
	}()
	return cc,nil
}
