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
	"fmt"
	"io"
	//"json"
	//"log"
	"net"
	"rand"
	"os"
	"sync"
	//"tonika/dbg"
	"tonika/prof"
	"tonika/http"
	"tonika/sys"
	"tonika/util/alarm"
	"tonika/util/backoff"
	//"tonika/util/term"
	"tonika/util/misc"
	trand "tonika/util/rand"
	"tonika/util/tube"
)

var connCounter counter
var sessionCounter counter

type Conn struct {
	tag    int64
	id     *sys.Id
	tube   tube.TubedConn
	dialch chan *connDial
	regime int
	rand   *rand.Rand
	err    os.Error
	lk     prof.Mutex
	hlk    sync.Mutex
	h      *handoff
}

type connDial struct {
	Subject string
	Notify  chan io.ReadWriteCloser
}

const (
	backoffLo    = 2e9           // in ns = 2 seconds
	backoffHi    = 60 * 30 * 1e9 // in ns = 30 minutes
	backoffRatio = 1.5           // backoff grows by this ratio, rounded up
	timeForFD    = 5e9           // wait at most 5 seconds to obtain a file descriptor
)

const (
	regimeIdle    = iota
	regimeConnect = iota
	regimeUnAuth  = iota
	regimeAuth    = iota
	regimeReady   = iota
	regimeBusy    = iota
	regimeClosed  = iota
	regimeDialing = iota
)

func regimeToString(regime int) string {
	switch regime {
	case regimeIdle:
		return "idle"
	case regimeConnect:
		return "connecting"
	case regimeUnAuth:
		return "un-authenticated"
	case regimeAuth:
		return "authenticating"
	case regimeReady:
		return "ready"
	case regimeBusy:
		return "busy"
	case regimeClosed:
		return "closed"
	case regimeDialing:
		return "dialing"
	default:
		panic("d·conn —— unknown regime")
	}
	panic("unreach")
}

func MakeConn() *Conn {
	return &Conn{
		tag: connCounter.Pick(),
		dialch: make(chan *connDial),
		rand:   trand.NewThreadUnsafeTimed(),
	}
}

func (y *Conn) GetRegime() int {
	y.lk.Lock()
	defer y.lk.Unlock()
	return y.regime
}

func (y *Conn) setRegime(regime int) os.Error {
	y.lk.Lock()
	defer y.lk.Unlock()
	if y.err != nil || y.regime == regimeClosed {
		return os.ErrorString("d,conn: closed meantime")
	}
	y.regime = regime
	return nil
}

func (y *Conn) RemoteId() *sys.Id {
	y.lk.Lock()
	defer y.lk.Unlock()
	return y.id
}

func (y *Conn) Error() os.Error {
	y.lk.Lock()
	defer y.lk.Unlock()
	return y.err
}

func (y *Conn) getTube() (tube.TubedConn, os.Error) {
	y.lk.Lock()
	defer y.lk.Unlock()
	if y.err != nil {
		return nil, y.err
	}
	return y.tube, nil
}

func (y *Conn) Connect(addr string, fdlim *http.FDLimiter) (err os.Error) {
	var alrm <-chan int
	backoff := backoff.Backoff{
		Lo:      backoffLo,
		Hi:      backoffHi,
		Ratio:   backoffRatio,
		Current: backoffLo,
		Attempt: 0,
	}
	for {
		if err = y.setRegime(regimeConnect); err != nil {
			return err
		}
		if alrm != nil {
			<-alrm
			alrm = nil
		}
		var conn net.Conn
		if fdlim.LockOrTimeout(timeForFD) == nil {
			conn, err = net.Dial("tcp", "", addr)
			if err == nil {
				//fmt.Printf(term.FgGreen+"d·conn[%#p] —— connected\n"+term.Reset, y)
				if err := conn.(*net.TCPConn).SetKeepAlive(true); err != nil {
					//log.Stderrf("%#p d·conn —— cannot set tcp keepalive: %s", y, err)
					fmt.Fprintf(os.Stderr, "tonika: warn: cannot set tcp keepalive\n")
				}
				conn = http.NewConnRunOnClose(conn, func() { fdlim.Unlock() })
				if err = y.Attach(conn); err != nil {
					conn.Close()
				}
				return err
			}
			fdlim.Unlock()
		}
		alrm = alarm.Ignite(backoff.Inc())
	}
	panic("unreachable")
}

func (y *Conn) Attach(rwc io.ReadWriteCloser) os.Error {
	y.lk.Lock()
	if y.err != nil {
		y.lk.Unlock()
		rwc.Close()
		return os.ErrorString("d,conn: closed meantime")
	}
	if y.regime != regimeIdle && y.regime != regimeConnect {
		panic("d,conn: regime logic")
	}
	y.regime = regimeUnAuth
	y.tube = tube.NewTube(rwc, nil)
	y.lk.Unlock()
	return nil
}

type U_Greet struct {
	Build   string // Tonika build
	Version string // Dialer version
}

func (y *Conn) Greet() (string, string, os.Error) {
	y.lk.Lock()
	if y.err != nil {
		y.lk.Unlock()
		return "", "", os.ErrorString("d,conn: closed meantime")
	}
	if y.regime != regimeUnAuth || y.tube == nil {
		panic("d,conn: regime logic")
	}
	y.regime = regimeAuth
	tube := y.tube
	y.lk.Unlock()

	g := &U_Greet{sys.Build, Version}
	err := tube.Encode(g)
	if err == nil {
		err = tube.Decode(g)
	}

	y.lk.Lock()
	if y.err != nil {
		y.lk.Unlock()
		return "", "", os.ErrorString("d,conn: closed meantime")
	}
	if err != nil {
		y.lk.Unlock()
		return "", "", y.kill(os.ErrorString("d,conn: greet failed"))
	}
	y.regime = regimeUnAuth
	y.lk.Unlock()
	return g.Build, g.Version, nil
}

type connAuthFunc func(tube.TubedConn) (sys.Id, tube.TubedConn, os.Error)

func (y *Conn) Auth(af connAuthFunc) (remote sys.Id, err os.Error) {
	y.lk.Lock()
	if y.err != nil {
		y.lk.Unlock()
		return 0, os.ErrorString("d,conn: closed meantime")
	}
	if y.regime != regimeUnAuth || y.tube == nil {
		panic("d,conn: regime logic")
	}
	y.regime = regimeAuth
	tube := y.tube
	y.lk.Unlock()

	remote, tubex, err := af(tube)

	y.lk.Lock()
	if y.err != nil {
		y.lk.Unlock()
		return 0, os.ErrorString("d,conn: closed meantime")
	}
	if err != nil {
		y.lk.Unlock()
		return 0, y.kill(os.ErrorString("d,conn: auth failed"))
	}
	y.id = &remote
	y.tube = tubex
	y.regime = regimeReady
	y.lk.Unlock()
	return remote, nil
}

func (y *Conn) kill(newerr os.Error) (err os.Error) {
	//dbg.PrintStackTrace()
	if newerr == nil {
		panic("d,conn: nil kill err")
	}
	y.lk.Lock()
	if y.err == nil {
		y.err = newerr
		err = newerr
	} else {
		y.lk.Unlock()
		return os.EBADF
	}
	y.id = nil
	y.regime = regimeClosed
	tube := y.tube
	y.tube = nil
	close(y.dialch)
	y.dialch = nil
	y.lk.Unlock()

	if tube != nil {
		tube.Close()
	}
	if err == nil {
		panic("d,conn: err == nil")
	}
	return err
}

func (y *Conn) getRandOrient() uint32 {
	y.lk.Lock()
	defer y.lk.Unlock()
	return uint32(y.rand.Int31n(U_OrientMax))
}

func (y *Conn) getRandSession() uint32 {
	y.lk.Lock()
	defer y.lk.Unlock()
	return uint32(y.rand.Int31())
}

// Wire structs
const U_OrientMax = 2e+9

type U_Orient struct {
	Order   uint32
	Session uint32
}

type U_Subject struct {
	Subject string
}

func (y *Conn) Poll() (subject string, rwc io.ReadWriteCloser, err os.Error) {
	for {
		y.hlk.Lock()
		y.hlk.Unlock()
		//fmt.Printf(term.FgBlack + "d·conn[%#p] —— loop\n" + term.Reset, y)

		y.lk.Lock()
		if y.err != nil || y.tube == nil {
			y.lk.Unlock()
			return "", nil, os.EBADF
		}
		tube := y.tube
		y.lk.Unlock()

		tch := waitForRead(tube.ExposeBufioReader())
		select {
		case m := <-y.dialch: // dial & kill requests come here
			if m == nil {
				//fmt.Printf(term.FgRed + "d·conn[%#p] —— closing accept loop\n"+ term.Reset, y)
				err = os.EBADF
				return "", nil, err
			}

			//fmt.Printf(term.FgYellow + "d·conn[%#p] —— dialing, subject=%s\n" + term.Reset, y, m.Subject)

			order := y.getRandOrient()
			session := y.getRandSession()
			orient := &U_Orient{order,session}
			if err = tube.Encode(orient); err != nil {
				err = y.kill(err)
				m.Notify <- nil
				return "", nil, err
			}
			err = <-tch // Hack. Reuse wait until data from remote
			if err != nil {
				//fmt.Printf(term.FgRed+"d·conn[%#p] —— dial/orient err: %s\n"+ term.Reset, y, err)
				err = y.kill(err)
				m.Notify <- nil
				return "", nil, err
			}
			err = tube.Decode(orient)
			if err != nil {
				//fmt.Printf(term.FgCyan+"d·conn[%#p] —— orient,dec: %s\n"+ term.Reset, y, err)
				err = y.kill(err)
				m.Notify <- nil
				return "", nil, err
			}
			if orient.Order >= order { // our dial prevails
				y.sendCall(session, m, tube)
			} else { // remote dial prevails
				m.Notify <- nil
				return y.receiveCall(orient.Session, tube)
			}

		case err = <-tch: // read is available on the connection
			if err != nil {
				//fmt.Printf(term.FgRed+"d·conn[%#p] —— ring/orient err: %s\n"+ term.Reset, y, err)
				err = y.kill(err)
				return "", nil, err
			}
			orient := &U_Orient{}
			if err = tube.Decode(orient); err != nil {
				//fmt.Printf(term.FgRed+"d·conn[%#p] —— ori,dec,err: %s\n"+ term.Reset, y, err)
				err = y.kill(err)
				return "", nil, err
			}
			orient.Order++
			session := orient.Session
			if err = tube.Encode(orient); err != nil {
				//fmt.Printf(term.FgRed+"d·conn[%#p] —— ori,enc,err: %s\n"+ term.Reset, y, err)
				err = y.kill(err)
				return "", nil, err
			}
			return y.receiveCall(session, tube)
		} // select
	} // for
	panic("unreach")
}

func (y *Conn) sendCall(session uint32, dial *connDial, tube tube.TubedConn) {
	subject := &U_Subject{dial.Subject}
	err := tube.Encode(subject)
	if err != nil || y.setRegime(regimeBusy) != nil {
		y.kill(os.ErrorString("d,conn: send call"))
		dial.Notify <- nil
		return
	}
	//fmt.Printf(term.FgCyan+"d·conn[%#p] —— dial tone, subject=%s\n"+term.Reset, y, dial.Subject)
	y.hlk.Lock()
	h := newHandoff(y, session, func(){ 
		y.setRegime(regimeReady)
		y.lk.Lock()
		y.h = nil
		y.lk.Unlock()
		y.hlk.Unlock() 
	})
	y.lk.Lock()
	y.h = h
	y.lk.Unlock()
	dial.Notify <- h
}

func (y *Conn) receiveCall(session uint32, 
	tube tube.TubedConn) (subject string, rwc io.ReadWriteCloser, err os.Error) {

	u_subject := &U_Subject{}
	err = tube.Decode(u_subject)
	if err != nil || u_subject.Subject == "" || y.setRegime(regimeBusy) != nil {
		err = y.kill(os.ErrorString("d,conn: receive call"))
		return "", nil, err
	}
	//fmt.Printf(term.FgCyan + "d·conn[%#p] —— ring! subject=%s\n"+term.Reset, y, u_subject.Subject)
	y.hlk.Lock()
	h := newHandoff(y, session, func(){ 
		y.setRegime(regimeReady)
		y.lk.Lock()
		y.h = nil
		y.lk.Unlock()
		y.hlk.Unlock() 
	})
	y.lk.Lock()
	y.h = h
	y.lk.Unlock()
	return u_subject.Subject, h, nil
}

// Dial works (and doesn't block) only if there is a concrrently running call to Poll().
// os.EAGAIN means connection is busy right now. Any other error indicates
// the connection is unfit for further use.
func (y *Conn) Dial(subject string) (rwc io.ReadWriteCloser, err os.Error) {
	if subject == "" {
		panic("d,conn: empty subject")
	}

	y.lk.Lock()
	dch := y.dialch
	err = y.err
	if err != nil {
		y.lk.Unlock()
		return nil, os.ErrorString("d,conn: closed meantime")
	}
	if closed(dch) {
		panic("d,conn: logic")
	}
	if y.regime != regimeReady {
		y.lk.Unlock()
		return nil, os.EAGAIN
	}
	y.regime = regimeDialing
	y.lk.Unlock()

	notify := make(chan io.ReadWriteCloser)
	msg := &connDial{subject, notify}
	dch <- msg
	rwc = <-notify
	if rwc == nil {
		return nil, os.EAGAIN
	}
	return rwc, nil
}

func (y *Conn) Close() os.Error {
	err := y.kill(os.EOF)
	if err == os.EOF {
		return nil
	}
	return err
}

// Atomically closes the connection if it is currently unused.
func (y *Conn) CloseIfReady() (err os.Error) {
	y.lk.Lock()
	if y.regime != regimeReady {
		y.lk.Unlock()
		return os.EAGAIN
	}
	if y.err != nil {
		y.lk.Unlock()
		return os.EBADF
	}
	y.err = os.EOF
	y.id = nil
	y.regime = regimeClosed
	tube := y.tube
	y.tube = nil
	close(y.dialch)
	y.dialch = nil
	y.lk.Unlock()

	if tube != nil {
		tube.Close()
	}
	return nil
}

func (y *Conn) String() string {
	y.lk.Lock()
	defer y.lk.Unlock()

	var w bytes.Buffer
	if y.id != nil {
		if y.h != nil {
			fmt.Fprintf(&w, "CTag: %d, HTag: %d, Session: %x, Id: %s, Err: %s -> %s",
				y.tag, y.h.GetTag(), y.h.GetSession(),
				y.id.String(), errorToString(y.err), regimeToString(y.regime))
		} else {
			fmt.Fprintf(&w, "CTag: %d, Id: %s, Err: %s -> %s",
				y.tag, 
				y.id.String(), errorToString(y.err), regimeToString(y.regime))
		}
	} else {
		fmt.Fprintf(&w, "CTag: %d, Id: n/a, Err: %s -> %s",
			y.tag, errorToString(y.err), regimeToString(y.regime))
	}
	return w.String()
}

func (y *Conn) MarshalJSON() ([]byte, os.Error) {
	y.lk.Lock()
	defer y.lk.Unlock()

	var w bytes.Buffer
	if y.id != nil {
		if y.h != nil {
			fmt.Fprintf(&w, "{\"CTag\":%d,\"HTag\":%d,\"Session\":\"%x\","+
				"\"Id\":%s,\"Err\":%s,\"Regime\":%s}",
				y.tag, y.h.GetTag(), y.h.GetSession(),
				y.id.ToJSON(),
				misc.JSONQuote(errorToString(y.err)),
				misc.JSONQuote(regimeToString(y.regime)))
		} else {
			fmt.Fprintf(&w, "{\"CTag\":%d,\"Id\":%s,\"Err\":%s,\"Regime\":%s}",
				y.tag,
				y.id.ToJSON(),
				misc.JSONQuote(errorToString(y.err)),
				misc.JSONQuote(regimeToString(y.regime)))
		}
	} else {
		fmt.Fprintf(&w, "{\"CTag\":%d,\"Err\":%s,\"Regime\":%s}",
			y.tag,
			misc.JSONQuote(errorToString(y.err)),
			misc.JSONQuote(regimeToString(y.regime)))
	}
	return w.Bytes(), nil
}

func errorToString(err os.Error) string {
	if err == nil {
		return "n/a"
	}
	return err.String()
}
