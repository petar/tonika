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


package compass

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sync"
	"time"
	"tonika/sys"
	tmath "tonika/math"
)

type watch struct {
	Neighbors map[sys.Id]*watchNeighbor
	MinProb   float64
	MaxProb   float64
	Sources   int
	lk        sync.Mutex
}

type watchNeighbor struct {
	TimeAppeared int64
	InTraffic    int64
	OutTraffic   int64
}

func (c *watch) Init() {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.Neighbors = make(map[sys.Id]*watchNeighbor)
	c.MinProb = math.NaN()
	c.MaxProb = math.NaN()
}

func (c *watch) AddNeighbor(id sys.Id) {
	c.lk.Lock()
	defer c.lk.Unlock()
	n := &watchNeighbor{TimeAppeared: time.Nanoseconds()}
	c.Neighbors[id] = n
}

func (c *watch) RemoveNeighbor(id sys.Id) {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.Neighbors[id] = nil,false
}

func (c *watch) RecordProb(value float64) {
	c.lk.Lock()
	defer c.lk.Unlock()
	if math.IsNaN(c.MinProb) || value < c.MinProb {
		c.MinProb = value
	}
	if math.IsNaN(c.MaxProb) || value > c.MaxProb {
		c.MaxProb = value
	}
}

func (c *watch) SetSources(v int) {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.Sources = v
}

func (c *watch) SetNeighborInTraffic(id sys.Id, amount int64) {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.Neighbors[id].InTraffic = amount
}

func (c *watch) SetNeighborOutTraffic(id sys.Id, amount int64) {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.Neighbors[id].OutTraffic = amount
}

func (c *watch) String() string {
	c.lk.Lock()
	defer c.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, 
		"MinDRes: %g, MaxDRes: %g, Sources: %d, Neighbors:\n", 
		c.MinProb, c.MaxProb, c.Sources)
	for id,n := range c.Neighbors {
		fmt.Fprintf(&w, "    Id: %s, Duration: %d ns, InTraffic: %d bytes, OutTraffic: %d bytes\n",
			id.Eye(), time.Nanoseconds() - n.TimeAppeared, n.InTraffic, n.OutTraffic)
	}
	return w.String()
}

func (c *watch) MarshalJSON() ([]byte, os.Error) {
	c.lk.Lock()
	defer c.lk.Unlock()
	var w bytes.Buffer
	if tmath.IsNum(c.MinProb) && tmath.IsNum(c.MaxProb) {
		fmt.Fprintf(&w, 
			"{\"MinDRes\":%g,\"MaxDRes\":%g,\"Sources\":%d,\"Neighbors\":[", 
			c.MinProb, c.MaxProb, c.Sources)
	} else {
		fmt.Fprintf(&w, 
			"{\"Sources\":%d,\"Neighbors\":[", c.Sources)
	}
	comma := false
	for id,n := range c.Neighbors {
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w, "{\"Id\":%s,\"Duration\":%d,\"InTraffic\":%d,\"OutTraffic\":%d}",
			id.ToJSON(), time.Nanoseconds() - n.TimeAppeared, 
			n.InTraffic, n.OutTraffic)
		comma = true
	}
	fmt.Fprintf(&w, "]}")
	return w.Bytes(), nil
}
