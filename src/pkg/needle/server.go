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


package needle

import (
	"net"
	"os"
	"sync"
	"time"
	"tonika/needle/proto"
	pb "goprotobuf.googlecode.com/hg/proto"
)

// TODO:
//   -- Add HTTP API server
//   -- Use LLRB for expiration algorithm

type Server struct {
	udp *net.UDPConn	// Socket that receives UDP pings from the clients
	ids map[int64]*client   // Id-to-client map
	lk  sync.Mutex          // Lock for ids field
}

const (
	ExpirePeriod    = 30e9     // Run expiration loop every 30 secs
	ClientFreshness = 5e9      // Expire clients who haven't pinged in the past 5 secs
)

// client describes real-time information for a given client
type client struct {
	id       int64
	lastSeen int64
	addr     *net.UDPAddr
}

func MakeServer(uaddr *net.UDPAddr) (*Server, os.Error) {
	conn, err := net.ListenUDP("udp", uaddr)
	if err != nil {
		return nil, err
	}
	err = conn.SetReadTimeout(ExpirePeriod / 2)
	if err != nil {
		return nil, err
	}
	s := &Server{
		udp: conn,
		ids: make(map[int64]*client),
	}
	go s.loop()
	return s, nil
}

// expire removes all client structures that have not been refreshed recently
func (s *Server) expire(now int64) {
	s.lk.Lock()
	defer s.lk.Unlock()

	for id, cl := range s.ids {
		if now - cl.lastSeen > ClientFreshness {
			s.ids[id] = nil, false
		}
	}
}

func (s *Server) updateClient(id int64, now int64, addr *net.UDPAddr) {
	s.lk.Lock()
	defer s.lk.Unlock()

	cl, ok := s.ids[id]
	if ok {
		cl.lastSeen = now
		cl.addr = addr
	} else {
		s.ids[id] = &client{
			id:       id,
			lastSeen: now,
			addr:     addr,
		}
	}
}

func (s *Server) poll() os.Error {

	// Read next UDP packet
	b := make([]byte, 32)
	n, addr, err := s.udp.ReadFromUDP(b)
	if err != nil {
		return err
	}
	
	// Decode packet contents
	payload := &proto.Ping{}
	err = pb.Unmarshal(b[0:n], payload)
	if err != nil {
		return err
	}

	// Make necessary updates
	s.updateClient(*payload.Id, time.Nanoseconds(), addr)

	return nil
}

func (s *Server) loop() {
	lastExpire := time.Nanoseconds()
	for {
		s.poll()
		now := time.Nanoseconds()
		if now - lastExpire > ExpirePeriod {
			s.expire(now)
			lastExpire = time.Nanoseconds()
		}
	}
}
