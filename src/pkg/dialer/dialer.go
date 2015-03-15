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


// The Dialer Subsystem is responsible for establishing a physical connection
// to a specified location. It is intended to use any means (UDP, TCP, uPNP,
// etc.) to get its hands on a reliable transport to the remote destination.
package dialer

import (
	"fmt"
	"io"
	//"log"
	"net"
	"os"
	//"sync"
	"time"
	"tonika/sys"
	"tonika/http"
	"tonika/prof"
	"tonika/util/tube"
	//"tonika/util/term"
)

type Dialer interface {
	Dial(id sys.Id, subject string) net.Conn
	Accept(subject string) (sys.Id, net.Conn)
	WaitForArrival() sys.Id
}

type Dialer0 struct {
	auth sys.AuthLocal

	ltcp    net.Listener
	tels    map[sys.Id]*telephone
	dials   map[sys.DialKey]*telephone
	unauthd map[*Conn]int
	listens map[string]chan *dialerRing

	fdlim http.FDLimiter
	lk    prof.Mutex
	err   os.Error

	arrivech chan sys.Id
	statusch chan *StatusUpdate
}

type dialerRing struct {
	id  sys.Id
	rwc io.ReadWriteCloser
}

type StatusUpdate struct {
	Id     sys.Id
	Online bool
}

func MakeDialer0(auth sys.AuthLocal, addr string, fdlim int) (d *Dialer0, err os.Error) {
	d = &Dialer0{
		tels:     make(map[sys.Id]*telephone),
		dials:    make(map[sys.DialKey]*telephone),
		unauthd:  make(map[*Conn]int),
		listens:  make(map[string]chan *dialerRing),
		arrivech: make(chan sys.Id, 5),
		statusch: make(chan *StatusUpdate, 5),
	}
	d.fdlim.Init(fdlim)
	if err := d.Bind(auth, addr); err != nil {
		return nil, err
	}
	go d.listen()
	return d, nil
}

func (d *Dialer0) getLocalAuth() sys.AuthLocal {
	d.lk.Lock()
	defer d.lk.Unlock()
	return d.auth
}

func (d *Dialer0) Bind(auth sys.AuthLocal, addr string) os.Error {
	l,err := net.Listen("tcp", addr)
	d.lk.Lock()	
	d.auth = auth
	d.ltcp, l = l, d.ltcp
	if err != nil {
		d.ltcp = nil
	} else {
		d.err = nil
	}
	d.lk.Unlock()
	if l != nil {
		l.Close()
	}
	return err
}

func (d *Dialer0) listen() os.Error {
	d.lk.Lock()
	fdlim := &d.fdlim
	d.lk.Unlock()
	for {
		d.lk.Lock()
		l := d.ltcp
		d.lk.Unlock()
		if l == nil {
			time.Sleep(10e9)
			continue
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
				d.setError(err)
				time.Sleep(10e9)
				continue
			}
			rwc := io.ReadWriteCloser(c)
			if fdlim != nil {
				rwc = newRunOnClose(c, func(){ fdlim.Unlock() })
			}
			go d.accept(rwc)
			continue
		}
		//log.Stderrf("%#p d —— fd starvation", d)
		fmt.Fprintf(os.Stderr, "tonika: warn: file descriptor starvation\n")
	}
	panic("unreach")
}

func (d *Dialer0) accept(rwc io.ReadWriteCloser) {
	conn := MakeConn()
	if err := conn.Attach(rwc); err != nil {
		rwc.Close()
		return
	}

	d.lk.Lock()
	d.unauthd[conn] = 1
	d.lk.Unlock()

	var err os.Error
	var remoteId sys.Id
	_, _, err = conn.Greet()

	if err == nil {
		// IMPORTANT: The call to d.getLocalAuth() cannot be in the lambda, because
		// then the d.Lock() will happen from inside Conn, which create a deadlock.
		// The only allowed locking order is Dialer->Tel->Conn->Handoff.
		localAuth := d.getLocalAuth()
		remoteId, err = conn.Auth(func(tube tube.TubedConn) (sys.Id, tube.TubedConn, os.Error) {
				return sys.AuthAccept(
					localAuth, 
					func(key *sys.DialKey) sys.AuthRemote { 
						return d.lookupTelAuth(key) 
					}, 
					tube)
			})
	}

	/*
	if err == nil {
		fmt.Printf(term.FgYellow+"accept--> authenticated\n"+term.Reset)
	} else {
		fmt.Printf(term.FgYellow+"accept--> auth err: %s\n"+term.Reset, err)
	}
	*/

	d.lk.Lock()
	d.unauthd[conn] = 0,false
	d.lk.Unlock()

	if err != nil {
		return
	}

	// TODO: Small. If tel changes (revoke+add) during authentication, we'd be
	// inserting a conn with stale authentication.
	t := d.getTel(remoteId)
	if t == nil {
		return
	}
	t.register(conn)
}

func (t *telephone) receive(subject string, rwc io.ReadWriteCloser) os.Error {
	t.lk.Lock()
	d := t.d
	auth := t.auth
	t.lk.Unlock()
	if d == nil {
		return os.EBADF
	}
	return d.receive(*auth.GetId(), subject, rwc)
}

func (d *Dialer0) receive(id sys.Id, subject string, rwc io.ReadWriteCloser) os.Error {
	d.lk.Lock()
	ch, ok := d.listens[subject]
	if !ok {
		d.lk.Unlock()
		return os.ErrorString("d: no listener for subject")
	}
	d.listens[subject] = nil, false
	d.lk.Unlock()
	ch <- &dialerRing{id, rwc}
	return nil
}

func (d *Dialer0) Accept(subject string) (remoteId sys.Id, conn net.Conn) {
	d.lk.Lock()
	if _, ok := d.listens[subject]; ok {
		panic("d: duplicate accept")
	}
	ch := make(chan *dialerRing)
	d.listens[subject] = ch
	d.lk.Unlock()
	dr := <-ch
	close(ch)
	if dr.rwc == nil {
		panic("d: accept err, rwc = nil")
	}
	return dr.id, newDialerConn(dr.rwc, *d.getLocalAuth().GetId(), dr.id)
}

func (d *Dialer0) WaitForArrival() sys.Id {
	return <-d.arrivech
}

func (d *Dialer0) arrived(id sys.Id) {
	_ = d.arrivech <- id
}

func (d *Dialer0) WaitForStatus() *StatusUpdate {
	return <-d.statusch
}

func (d *Dialer0) announceOnline(id sys.Id, v bool) {
	_ = d.statusch <- &StatusUpdate{id,v}
}

func (d *Dialer0) Error() os.Error {
	d.lk.Lock()
	defer d.lk.Unlock()
	return d.err
}

func (d *Dialer0) setError(err os.Error) {
	d.lk.Lock()
	defer d.lk.Unlock()
	d.err = err
}
