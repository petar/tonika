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
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"
	"tonika/sys"
	"tonika/util/signal"
	"tonika/util/iopipe"
	"tonika/util/term"
)

func authfunc(EncodeDecoder) (sys.Id, os.Error)  {
	return 0, nil
}

const msg = "Hello World!"

func lead(u io.ReadWriteCloser, q int, t *testing.T, done chan int) {
	c := MakeConn()
	if err := c.Attach(u); err != nil {
		t.Fatalf("attach[%d]: %s", q, err)
	}
	if _, err := c.Auth(nil, authfunc); err != nil {
		t.Fatalf("auth[%d]: %s", q, err)
	}
	if q == 0 {
		go dialAndRead(c,t, done)
	}
	for {
		_, rwc, err := c.Poll()
		if err != nil {
			t.Fatalf("poll[%d]: %s", q, err)
		}
		if q == 1 {
			go acceptAndWrite(rwc,t)
		}
	}
}

func acceptAndWrite(rwc io.ReadWriteCloser, t *testing.T) {
	// write
	buf := []byte(msg)
	for len(buf) > 0 {
		n,err := rwc.Write(buf)
		if err != nil {
			t.Fatalf("write: %s", err)
		}
		buf = buf[n:]
	}

	// read
	buf2 := make([]byte, len(msg))
	free := buf2
	for len(free) > 0 {
		n,err := rwc.Read(free)
		if err != nil {
			t.Fatalf("write/read: %s", err)
		}
		free = free[n:]
	}
	if string(buf2) != msg {
		t.Fatalf("write/read crc fail")
	}

	// close
	if err := rwc.Close(); err != nil {
		t.Fatalf("write/close: %s", err)
	}
}

func dialAndRead(c *Conn, t *testing.T, done chan int) {
	for k := 0; k < 10; k++ {
		rwc, err := c.Dial("sss")
		if err != nil {
			t.Fatalf("dial: %s", err)
		}

		// read
		buf := make([]byte, len(msg))
		free := buf
		for len(free) > 0 {
			n,err := rwc.Read(free)
			if err != nil {
				t.Fatalf("read: %s", err)
			}
			free = free[n:]
		}
		if string(buf) != msg {
			t.Fatalf("read crc fail")
		}

		// write
		buf2 := []byte(msg)
		for len(buf2) > 0 {
			n,err := rwc.Write(buf2)
			if err != nil {
				t.Fatalf("read/write: %s", err)
			}
			buf2 = buf2[n:]
		}

		// close
		if err := rwc.Close(); err != nil {
			t.Fatalf("read/close: %s", err)
		}
		fmt.Printf(term.FgRed+"READ/WRITE OK %d\n"+term.Reset, k+1)
		time.Sleep(1e9)
	}
	done <- 1
}

func TestConn(t *testing.T) {
	signal.InstallCtrlCPanic()
	r1,r2 := iopipe.MakePipe()
	done := make(chan int)
	go lead(r1, 0, t, done)
	go lead(r2, 1, t, nil)
	<-done
}

func TestConnOverTCP(t *testing.T) {
	signal.InstallCtrlCPanic()

	go func() {
		l,err := net.Listen("tcp",":44000")
		if err != nil {
			t.Fatalf("listen: %s", err)
		}
		c,err := l.Accept()
		if err != nil {
			t.Fatalf("accept: %s", err)
		}
		go lead(c, 1, t, nil)
	}()

	c,err := net.Dial("tcp","",":44000")
	if err != nil {
		t.Fatalf("dial: %s", err)
	}
	done := make(chan int)
	go lead(c, 0, t, done)
	<-done
}
