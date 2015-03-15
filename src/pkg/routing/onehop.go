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


package routing

import (
	"bytes"
	"fmt"
	"os"
	"time"
	. "tonika/sys"
)

// oneHopRouting is a trivial routing algorithm that can route messages
// to one's immediate neighbors.
type oneHopRouting struct {
	me        Id
	neighbors map[Id]int64
}

func MakeOneHopRouting(id Id) *oneHopRouting {
	return &oneHopRouting{ me:id, neighbors: make(map[Id]int64) }
}

func (oh *oneHopRouting) NeedBand() bool { return false }

func (oh *oneHopRouting) OnAddNeighbor(nid Id) {
	oh.neighbors[nid] = time.Nanoseconds()
}

func (oh *oneHopRouting) OnRemoveNeighbor(nid Id) {
	oh.neighbors[nid] = 0, false
}

func (oh *oneHopRouting) NewEmptyMsg() interface{} { panic("no band") }

func (oh *oneHopRouting) OnReceive(from Id, msg interface{}) os.Error { panic("no band") }

func (oh *oneHopRouting) OnClock() {}

func (oh *oneHopRouting) Sweep() {}

func (oh *oneHopRouting) PollSend(to Id) interface{} { panic("no band") }

func (oh *oneHopRouting) Id() Id { return oh.me }

func (oh *oneHopRouting) Neighbors() []Id {
	r := make([]Id, len(oh.neighbors))
	i := 0
	for nid,_ := range oh.neighbors {
		r[i] = nid
		i++
	}
	return r
}

func (oh *oneHopRouting) SourceCount() int { return len(oh.neighbors)+1 }

func (oh *oneHopRouting) FlowFork(s,t Id) (fork Fork, err os.Error) {
	_,ok := oh.neighbors[t]
	if !ok {
		return nil, os.ErrorString("no route")
	}
	nf := &neighborFork{}
	nf.Add(1.0, t)
	return nf, nil
}

func (oh *oneHopRouting) FlowStep(s,t Id) (hop Id, err os.Error) {
	_,ok := oh.neighbors[t]
	if !ok {
		return 0, os.ErrorString("no route")
	}
	return t, nil
}

func (oh *oneHopRouting) String() string {
	var w bytes.Buffer
	fmt.Fprintf(&w, "Liaisons:\n")
	for nid,_ := range oh.neighbors {
		fmt.Fprintf(&w, "  %v\n", nid.String())
	}
	return w.String()
}

func (oh *oneHopRouting) MarshalJSON() ([]byte, os.Error) {
	var w bytes.Buffer
	fmt.Fprintf(&w, "{Algorithm: \"OneHop\", \"Neighbors\":[")
	comma := false
	for nl,_ := range oh.neighbors {
		if comma {
			w.WriteString(",")
		}
		comma = true
		fmt.Fprintf(&w, "{\"Id\":\"%s\"}", nl.String())
	}
	fmt.Fprintf(&w, "]}")
	return w.Bytes(), nil
}
