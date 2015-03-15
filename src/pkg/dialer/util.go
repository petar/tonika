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


package dialer

import (
	"bufio"
	"io"
	"net"
	"os"
	"sync"
	"tonika/sys"
)

// net.Addr for Tonika
type dialerAddr string

func (da dialerAddr) Network() string { return "tonika" }

func (da dialerAddr) String() string { return string(da) }

// A net.Conn wrapper for io.ReadWriteCloser's returned by Dialer methods
type dialerConn struct {
	io.ReadWriteCloser
	la,ra dialerAddr
}

func newDialerConn(rwc io.ReadWriteCloser, local, remote sys.Id) *dialerConn {
	return &dialerConn{ rwc, dialerAddr(local.Eye()), dialerAddr(remote.Eye()) }
}

func (dconn *dialerConn) LocalAddr() net.Addr { return dconn.la }
func (dconn *dialerConn) RemoteAddr() net.Addr { return dconn.ra }
func (dconn *dialerConn) SetTimeout(nsec int64) os.Error { return nil }
func (dconn *dialerConn) SetReadTimeout(nsec int64) os.Error { return nil }
func (dconn *dialerConn) SetWriteTimeout(nsec int64) os.Error { return nil }

// Hack to wait until read is available on a bufio.Reader
func waitForRead(bufr *bufio.Reader) <-chan os.Error {
	ch := make(chan os.Error)
	go func() {
		_, err := bufr.ReadByte()
		if err != nil {
			ch <- err
			return
		}
		err = bufr.UnreadByte()
		if err != nil {
			ch <- err
			return
		}
		ch <- nil
	}()
	return ch
}

// runOnClose runs a user-supplied subroutine after the first invokation of Close.
type runOnClose struct {
	io.ReadWriteCloser
	run func()
}

func newRunOnClose(rwc io.ReadWriteCloser, f func()) *runOnClose {
	return &runOnClose{rwc, f}
}

func (roc *runOnClose) Close() os.Error {
	err := roc.ReadWriteCloser.Close()
	if roc.run != nil {
		roc.run()
		roc.run = nil
	}
	return err
}

// Pick a number in sequence 1,2,3,4,...
type counter struct {
	k  int64
	lk sync.Mutex
}

func (c *counter) Init() {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.k = 0
}

func (c *counter) Pick() int64 {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.k++
	return c.k
}
