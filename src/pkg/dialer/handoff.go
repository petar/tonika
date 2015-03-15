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
	"bytes"
	//"fmt"
	"os"
	//"tonika/dbg"
	"tonika/prof"
	//"tonika/util/term"
	"tonika/util/tube"
)

// TODO:
//   (*) Timeout mechanism: If local or remote does not use the connection for a while,
//   close the connection gracefully and return control to underlying Conn.

var handoffCounter counter

// IMPORTANT: None of the handoff public methods (Read/Write/Close) can be called from
// inside a telephone lock.
type handoff struct {
	tag     int64
	session uint32
	y       *Conn
	rn, wn  int64        // # bytes read, # bytes written
	rk, wk  int64        // # read calls, # write calls
	buf     bytes.Buffer // session read-side buffer
	rclosed bool
	finfun  func()
	lk      prof.Mutex
}

const handoffUpFreq = 4

// Packaging of user data used by handoff. A zero-length cargo means EOF
type U_Cargo struct {
	Cargo []byte
}

func newHandoff(y *Conn, ses uint32, finfun func()) *handoff {
	h := &handoff{tag: handoffCounter.Pick(), session: ses, y: y, finfun: finfun}
	//fmt.Printf(term.FgYellow+"d·conn[%#p]·h[%#p]:%x —— handoff\n"+term.Reset, y,h,h.session)
	return h
}

func (h *handoff) GetTag() int64 { return h.tag }
func (h *handoff) GetSession() uint32 { return h.session }

func (h *handoff) Read(p []byte) (n int, err os.Error) {
	tube, err := h.checkForKill()
	if err != nil {
		return 0, err
	}
	h.lk.Lock()
	if h.rclosed {
		h.lk.Unlock()
		return 0, os.EOF
	}
	h.lk.Unlock()

	// No need to lock on h.buf, since it's only used here, and
	// concurrent read calls are not allowed.
	if h.buf.Len() == 0 {
		msg := &U_Cargo{}
		if err = tube.Decode(msg); err != nil {
			return 0, h.kill(err)
		}
		// A 0-length cargo is an indication of session EOF
		if msg.Cargo == nil || len(msg.Cargo) == 0 {
			h.lk.Lock()
			defer h.lk.Unlock()
			if h.y == nil {
				return 0, os.EBADF
			}
			h.rclosed = true
			return 0, os.EOF
		}
		n, _ = h.buf.Write(msg.Cargo)
		if n != len(msg.Cargo) {
			panic("d,conn,h: buf write")
		}
		h.lk.Lock()
		h.rn += int64(n)
		h.rk++
		h.lk.Unlock()
	}

	if h.buf.Len() == 0 {
		panic("handoff logic")
	}
	n, err = h.buf.Read(p)
	if err != nil {
		panic("d,conn,h: buf")
	}

	h.lk.Lock()
	defer h.lk.Unlock()
	if h.y == nil {
		return 0, os.EBADF
	}
	return n, nil
}

func (h *handoff) Write(p []byte) (n int, err os.Error) {
	tube, err := h.checkForKill()
	if err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}

	msg := &U_Cargo{p}
	if err = tube.Encode(msg); err != nil {
		return 0, h.kill(err)
	}
	h.lk.Lock()
	defer h.lk.Unlock()
	h.wn += int64(len(p))
	h.wk++
	if h.y == nil {
		return 0, os.EBADF
	}
	return len(p), nil
}

// Close terminates this handoff and returns control to the underlying Conn.
func (h *handoff) Close() (err os.Error) {
	h.lk.Lock()
	y := h.y
	h.y = nil
	if y == nil {
		h.lk.Unlock()
		return os.EBADF
	}
	rclosed := h.rclosed
	h.lk.Unlock()

	//fmt.Printf(term.FgCyan+"d·conn[%#p]·h[%#p]:%x —— close\n"+term.Reset, y,h,h.session)

	tube, err := y.getTube()
	if err != nil {
		return os.EIO
	}

	msg := &U_Cargo{}
	if err = tube.Encode(msg); err != nil {
		return h.kill(err)
	}

	if !rclosed {
		msg := &U_Cargo{}
		if err = tube.Decode(msg); err != nil {
			goto __End
		}
		if msg.Cargo == nil || len(msg.Cargo) == 0 {
			goto __End
		}
	}
__End:
	h.lk.Lock()
	ff := h.finfun
	h.finfun = nil
	h.lk.Unlock()
	if ff != nil {
		ff()
	}
	if err != nil {
		err = y.kill(err)
	}
	return err
}

// kill terminates this handoff as well as the underlying Conn.
func (h *handoff) kill(err os.Error) os.Error {
	h.lk.Lock()
	if h.y == nil {
		h.lk.Unlock()
		return os.EBADF
	}
	y := h.y
	//fmt.Printf(term.FgCyan+"d·conn[%#p]·h[%#p]:%x —— kill\n"+term.Reset, y,h,h.session)
	h.y = nil
	ff := h.finfun
	h.finfun = nil
	h.lk.Unlock()

	err = y.kill(err)
	if ff != nil {
		ff()
	}
	return err
}

func (h *handoff) checkForKill() (tube tube.TubedConn, err os.Error) {
	h.lk.Lock()
	if h.y == nil {
		h.lk.Unlock()
		return nil, os.EBADF
	}
	y := h.y
	h.lk.Unlock()

	tube, err = y.getTube()
	if err != nil {
		return nil, os.EIO
	}
	return tube, nil
}
