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
	//"fmt"
	"io"
	"net"
	"os"
	"time"
	//"sync"
	"tonika/sys"
	"tonika/prof"
	//"tonika/util/term"
	"tonika/util/tube"
	"tonika/util/uptime"
)

type telephone struct {
	d        *Dialer0
	auth     sys.AuthRemote
	addr     string

	presence sys.Presence
	uptime   uptime.Uptime

	authing  map[*Conn]int
	conns    map[*Conn]int

	lk prof.Mutex
}

const (
	preConnect   = 3 // # of pre-connected
	maxPending   = 5 // max # of pending dials per telephone
	maxDialTries = 7
	replenishWait = 1e9
)

func makeTel(d *Dialer0, auth sys.AuthRemote, addr string) *telephone {
	return &telephone{
		d:        d,
		auth:     auth,
		addr:     addr,
		presence: sys.Presence{
			Id:          *auth.GetId(),
			MaybeOnline: true,
			Reachable:   true,
			Rating:      sys.InitialRating,
		},
		uptime:   uptime.Make(sys.Halflife, sys.RatingBound),
		authing:  make(map[*Conn]int),
		conns:    make(map[*Conn]int),
	}
}

func (t *telephone) GetAuth() sys.AuthRemote { return t.auth }

func (t *telephone) healthy() *Dialer0 {
	t.lk.Lock()
	defer t.lk.Unlock()
	return t.d
}

// Registers a connection with this telephone, if the telephone
// is still healthy.
func (t *telephone) register(conn *Conn) {
	for {
		t.lk.Lock()
		if t.d == nil {
			t.lk.Unlock()
			return
		}
		d := t.d
		t.conns[conn] = 1
		t.lk.Unlock()
		d.arrived(*t.auth.GetId())
		t.rebalance()

		subject, hrwc, err := conn.Poll()
		if err != nil {
			//fmt.Printf("d·tel[%#p] —— conn[%#p].Poll/err = %s\n", t, conn, err)
			break
		}
		if hrwc == nil {
			panic("hrwc == nil")
		}
		t.rebalance()
		if err = t.receive(subject, hrwc); err != nil {
			//fmt.Printf("d·tel[%#p].receive(%s, %#p),conn[%#p] —— err = %s\n",
			//	t, subject, hrwc, conn, err)
		}
		if conn.Error() != nil {
			break
		}
		t.rebalance()
	}
	t.killConn(conn)
	t.rebalance()
}

func (t *telephone) getReadyBusyCounts() (nr, nb int) {
	for c,_ := range t.conns {
		if c.GetRegime() == regimeReady {
			nr++
		} else {
			nb++
		}
	}
	return nr, nb
}

func (t *telephone) guessOnline() {
	t.lk.Lock()
	if t.d == nil {
		t.lk.Unlock()
		return
	}
	nrdy,_ := t.getReadyBusyCounts()
	p1 := nrdy > 0 
	p0 := t.presence.MaybeOnline
	t.presence.MaybeOnline = p1
	id := *t.auth.GetId()
	t.lk.Unlock()
	if p0 == p1 {
		return
	}
	t.d.announceOnline(id, p1)
}

func (t *telephone) rebalance() {

	t.lk.Lock()
	if t.d == nil {
		t.lk.Unlock()
		return
	}
	nrdy,_ := t.getReadyBusyCounts()
	short := preConnect - len(t.authing) - nrdy
	reachable := t.presence.Reachable
	t.lk.Unlock()

	if short > 0 && reachable {
		// TODO: Small bug. The connecting conns appear in t.conning a but after
		// the call to preconnect. So a busrst of rebalance() may not see that there
		// are connecting conns already.
		t.preconnect(short)
	}

	// Remove some connections if too many
	t.lk.Lock()
	nrdy,_ = t.getReadyBusyCounts()
	short = nrdy - 2*preConnect // # of conn's to kill
	for conn, _ := range t.conns {
		if conn.Error() != nil {
			t.conns[conn] = 0, false
			conn.Close() // this blocks but hopefully not for long
			continue
		}
		if short > 0 {
			kerr := conn.CloseIfReady()
			if kerr != os.EAGAIN {
				t.conns[conn] = 0, false
				short--
			}
		}
	}
	t.lk.Unlock()
	t.guessOnline()
}

func (t *telephone) preconnect(howmany int) {
	for i := 0; i < howmany; i++ {
		go t.connect()
	}
}

func (t *telephone) connect() {
	conn := MakeConn()
	t.lk.Lock()
	if t.d == nil {
		t.lk.Unlock()
		return
	}
	d := t.d
	addr := t.addr
	t.authing[conn] = 1
	t.lk.Unlock()

	err := conn.Connect(addr, &d.fdlim)
	if err != nil {
		t.lk.Lock()
		t.authing[conn] = 0, false
		t.lk.Unlock()
		conn.Close()
		t.rebalance()
		return
	}

	t.lk.Lock()
	if t.d == nil {
		t.authing[conn] = 0, false
		t.lk.Unlock()
		return
	}
	d = t.d
	auth := t.auth
	t.lk.Unlock()

	if auth.GetId() == nil {
		panic("d,tel: auth")
	}

	_, _, err = conn.Greet()

	if err == nil {
		// IMPORTANT: The call to d.getLocalAuth() cannot be in the lambda, because
		// then the d.Lock() will happen from inside Conn, which create a deadlock.
		// The only allowed locking order is Dialer->Tel->Conn->Handoff.
		localAuth := d.getLocalAuth()
		_, err = conn.Auth(func(tube tube.TubedConn) (sys.Id, tube.TubedConn, os.Error) {
				return sys.AuthConnect(localAuth, auth, tube)
			})
	}

	/*
	if err == nil {
		fmt.Printf(term.FgYellow+"connect--> authenticated\n"+term.Reset)
	} else {
		fmt.Printf(term.FgYellow+"connect--> auth err: %s\n"+term.Reset, err)
	}
	*/

	t.lk.Lock()
	t.authing[conn] = 0, false
	t.lk.Unlock()
	if err != nil {
		conn.Close()
		t.rebalance()
		return
	}

	t.register(conn)
}

func (t *telephone) kill() {
	t.lk.Lock()
	t.d = nil
	for conn,_ := range t.conns {
		t.conns[conn] = 0, false
		conn.Close()
	}
	t.lk.Unlock()
}

// Dialer methods about telephones

func (d *Dialer0) getTel(id sys.Id) *telephone {
	d.lk.Lock()
	t, ok := d.tels[id]
	d.lk.Unlock()
	if !ok {
		return nil
	}
	return t
}

func (d *Dialer0) Add(auth sys.AuthRemote, addr string) {
	d.lk.Lock()
	_, present := d.tels[*auth.GetId()]
	d.lk.Unlock()
	if present {
		return
	}
	t := makeTel(d, auth, addr)
	d.lk.Lock()
	d.tels[*auth.GetId()] = t
	d.dials[*auth.GetAcceptKey()] = t
	d.lk.Unlock()
	t.rebalance()
}

func (d *Dialer0) lookupTelAuth(dk *sys.DialKey) sys.AuthRemote {
	d.lk.Lock()
	t, ok := d.dials[*dk]
	d.lk.Unlock()
	if !ok {
		return nil
	}
	return t.GetAuth()
}

func (d *Dialer0) Update(id sys.Id, addr string) {
	t := d.getTel(id)
	if t == nil {
		return
	}
	t.lk.Lock()
	t.addr = addr
	t.presence.Reachable = true
	t.lk.Unlock()
	t.rebalance()
}

func (d *Dialer0) Revoke(id sys.Id) {
	d.lk.Lock()
	t, present := d.tels[id]
	d.tels[id] = nil, false
	if present {
		d.dials[*t.GetAuth().GetAcceptKey()] = nil, false
	}
	d.lk.Unlock()
	if !present {
		return
	}
	t.kill()
}

func (d *Dialer0) Presence(id sys.Id) *sys.Presence {
	t := d.getTel(id)
	if t == nil {
		return nil
	}
	t.lk.Lock()
	defer t.lk.Unlock()
	r := t.presence
	r.Rating = t.uptime.Rating
	return &r
}

func (d *Dialer0) Dial(id sys.Id, subject string) net.Conn {
	if subject == "" {
		return nil
	}
	t := d.getTel(id)
	if t == nil {
		return nil
	}
	rwc := t.dial(subject)
	if rwc == nil {
		return nil
	}
	return newDialerConn(rwc, *d.getLocalAuth().GetId(), id)
}

func (t *telephone) dial(subject string) io.ReadWriteCloser {
	for j := 0; j < maxDialTries; j++ {
		t.lk.Lock()
		if t.d == nil {
			t.lk.Unlock()
			return nil
		}
		rs := make([]*Conn,len(t.conns))
		i := 0
		for conn,_ := range t.conns {
			rs[i] = conn
			i++
		}
		t.lk.Unlock()
		
		for i = 0; i < len(rs); i++ {
			rwc, err := rs[i].Dial(subject)
			if err == nil {
				t.rebalance()
				return newRunOnClose(rwc, func(){ 
					if rs[i].Error() != nil {
						t.killConn(rs[i])
					}
					t.rebalance()
				})
			}
			if rs[i].Error() != nil {
				t.killConn(rs[i])
			}
		}
		time.Sleep(replenishWait)
	}
	return nil
}

func (t *telephone) killConn(conn *Conn) {
	if conn == nil {
		return
	}
	t.lk.Lock()
	t.conns[conn] = 0, false
	t.lk.Unlock()
	conn.Close()
}
