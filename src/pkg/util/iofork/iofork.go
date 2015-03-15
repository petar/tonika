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


// Duplicate all info read from a reader into a second reader
package iofork

import (
	"bytes"
	"io"
	"net"
	"os"
)

func NewReader(r io.Reader, lim int) (r1 io.Reader, r2 io.Reader) {
	fr1 := &forkReader1{
		Reader: r,
		lim: lim,
		buf: bytes.NewBuffer(nil),
	}
	return fr1, (*forkReader2)(fr1)
}

type forkReader1 struct {
	io.Reader
	lim int // panic of limit is reached
	buf *bytes.Buffer
}

func (fr1 *forkReader1) Read(p []byte) (n int, err os.Error) {
	n,err = fr1.Reader.Read(p)
	if n > 0 {
		if m,_ := fr1.buf.Read(p[0:n]); m != n {
			panic("iofork")
		}
	}
	return
}

type forkReader2 forkReader1

func (fr2 *forkReader2) Read(p []byte) (n int, err os.Error) {
	return fr2.buf.Read(p)
}

// (===) connection snooping

func NewConn(c net.Conn) (c1 net.Conn, r2 io.Reader) {
	r1, r2 := NewReader(c, 512)
	return &forkConn1{c, r1}, r2
}

type forkConn1 struct {
	net.Conn
	r1 io.Reader
}

func (fc1 *forkConn1) Read(b []byte) (n int, err os.Error) {
	return fc1.r1.Read(b)
}

func (fc1 *forkConn1) Write(b []byte) (n int, err os.Error) {
	return fc1.Conn.Write(b)
}

func (fc1 *forkConn1) Close() os.Error { return fc1.Conn.Close() }

func (fc1 *forkConn1) LocalAddr() net.Addr { return fc1.Conn.LocalAddr() }

func (fc1 *forkConn1) RemoteAddr() net.Addr { return fc1.Conn.RemoteAddr() }

func (fc1 *forkConn1) SetTimeout(nsec int64) os.Error {
	return fc1.Conn.SetTimeout(nsec)
}

func (fc1 *forkConn1) SetReadTimeout(nsec int64) os.Error {
	return fc1.Conn.SetReadTimeout(nsec)
}

func (fc1 *forkConn1) SetWriteTimeout(nsec int64) os.Error {
	return fc1.Conn.SetWriteTimeout(nsec)
}
