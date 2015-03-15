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
	"gob"
	"io"
	"log"
	"net"
	//"runtime"
	"strconv"
	"testing"
	"time"
	"tonika/proto"
	"tonika/util/term"
)

const (
	localId  = sys.Id(1177)
	remoteId = sys.Id(9955)
)

func remote(id sys.Id, rwc io.ReadWriteCloser, t *testing.T) {
	enc := gob.NewEncoder(rwc)
	dec := gob.NewDecoder(rwc)
	// send auth
	pauth := ProtoAuth{id}
	if err := enc.Encode(&pauth); err != nil {
		t.Fatalf("remote send auth")
	}
	// read auth
	if err := dec.Decode(&pauth); err != nil {
		t.Fatalf("remote read/dec auth")
	}
	if pauth.Id != localId {
		t.Fatalf("bad local id")
	}
	log.Stderrf(term.FgCyan + "** authed!" + term.Reset)
	// read orient
	porient := ProtoOrient{}
	if err := dec.Decode(&porient); err != nil {
		t.Fatalf("remote read/dec orient")
	}
	// send orient
	porient.Order++
	if err := enc.Encode(&porient); err != nil {
		t.Fatalf("remote write/enc orient")
	}
	log.Stderrf(term.FgCyan + "** oriented!" + term.Reset)
	// read subject
	psubj := ProtoSubject{}
	if err := dec.Decode(&psubj); err != nil {
		t.Fatalf("remote read/dec subject")
	}
	log.Stderrf(term.FgCyan+"** ok! subject: %s"+term.Reset, psubj.Subject)
	time.Sleep(2e9)
	rwc.Close()
}

func listen(t *testing.T) {
	time.Sleep(1e9)
	//log.Stderrf("R· starting remote listener\n")
	l, err := net.Listen("tcp", ":22333")
	if err != nil {
		panic("Can't listen")
	}
	for {
		c, err := l.Accept()
		if err != nil {
			//log.Stderrf("R· accept error: %v", err)
		} else {
			//log.Stderrf("R· accept success!")
			go remote(remoteId, c, t)
		}
		time.Sleep(1e9)
	}
}

func TestDialer0(t *testing.T) {
	localFriend := proto.Friend{
		Id:   localId,
		Addr: "localhost:55555",
	}
	remoteFriend := proto.Friend{
		Id:   remoteId,
		Addr: "localhost:22333",
	}
	cmdch, err := MakeDialer0(localFriend)
	if err != nil {
		t.Fatalf("make dialer")
	}

	go listen(t)

	cmdch <- cmdAdd{remoteFriend}

	// dials
	time.Sleep(5e9)
	nDials := 5
	nfs := make([]chan io.ReadWriteCloser, nDials)
	for i := 0; i < nDials; i++ {
		nfs[i] = make(chan io.ReadWriteCloser, 1)
		cmdch <- cmdDial{remoteId, "S" + strconv.Itoa(i), nfs[i]}
	}

	fvr := make(chan int)
	<-fvr
}
