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


// Profiling primitives
package prof

import (
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// ReadWriteCloser
type ReadWriteCloser struct {
	io.ReadWriteCloser
	in,out int64
	begin,end int64 // ns
	lk sync.Mutex
}

func NewReadWriteCloser(rwc io.ReadWriteCloser) *ReadWriteCloser {
	return &ReadWriteCloser{ ReadWriteCloser: rwc, begin: time.Nanoseconds() }
}

func (r *ReadWriteCloser) InTraffic() int64 { 
	r.lk.Lock()
	defer r.lk.Unlock()
	return r.in 
}

func (r *ReadWriteCloser) OutTraffic() int64 { 
	r.lk.Lock()
	defer r.lk.Unlock()
	return r.out 
}

func (r *ReadWriteCloser) Duration() int64 {
	r.lk.Lock()
	defer r.lk.Unlock()
	if r.end == 0 {
		return time.Nanoseconds() - r.begin
	}
	return r.end - r.begin
}

func (r *ReadWriteCloser) Read(p []byte) (n int, err os.Error) {
	n,err = r.ReadWriteCloser.Read(p)
	r.lk.Lock()
	r.in += int64(n)
	r.lk.Unlock()
	return n,err
}

func (r *ReadWriteCloser) Write(p []byte) (n int, err os.Error) {
	n,err = r.ReadWriteCloser.Write(p)
	r.lk.Lock()
	r.out += int64(n)
	r.lk.Unlock()
	return n,err
}

func (r *ReadWriteCloser) Close() (err os.Error) {
	err = r.ReadWriteCloser.Close()
	r.lk.Lock()
	if r.end != 0 {
		r.end = time.Nanoseconds()
	}
	r.lk.Unlock()
	return
}

// Conn
type Conn struct {
	net.Conn
	in,out int64
	begin,end int64 // ns
	lk sync.Mutex
}

func NewConn(c net.Conn) *Conn {
	return &Conn{ Conn: c, begin: time.Nanoseconds() }
}

func (c *Conn) InTraffic() int64 { 
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.in 
}

func (c *Conn) OutTraffic() int64 { 
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.out 
}

func (c *Conn) Duration() int64 {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.end == 0 {
		return time.Nanoseconds() - c.begin
	}
	return c.end - c.begin
}

func (c *Conn) Read(p []byte) (n int, err os.Error) {
	n,err = c.Conn.Read(p)
	c.lk.Lock()
	c.in += int64(n)
	c.lk.Unlock()
	return n, err
}

func (c *Conn) Write(p []byte) (n int, err os.Error) {
	n,err = c.Conn.Write(p)
	c.lk.Lock()
	c.out += int64(n)
	c.lk.Unlock()
	return n, err
}

func (c *Conn) Close() (err os.Error) {
	err = c.Conn.Close()
	c.lk.Lock()
	if c.end == 0 {
		c.end = time.Nanoseconds()
	}
	c.lk.Unlock()
	return
}
