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


// *** SRC ***

package compass

import (
	"bytes"
	"fmt"
	//"sync"
	"os"
	"time"
	//"tonika/dbg"
	"tonika/dialer"
	"tonika/prof"
	"tonika/routing"
	"tonika/sys"
	//"tonika/util/term"
)

type Compass interface {
	//QueryMeasure(s,t sys.Id) *sys.Id
	QueryQuantize(s,t sys.Id) *sys.Id
}

type Compass0 struct {
	w        watch
	d        dialer.Dialer
	algo     routing.Algorithm
	liaisons map[sys.Id]*liaison
	lk       prof.Mutex
}

type liaison struct {
	id  sys.Id
	edc *EncodeDecodeCloser
	wlk prof.Mutex
}

const (
	clockTick      = 60e9 // clock ticks once every minute
	sweepFrequency = 2 // 2 clock ticks = 2 min
)

func MakeCompass0(id sys.Id, d dialer.Dialer) *Compass0 {
	c := &Compass0{
		d:        d,
		algo:     routing.MakeOneHopRouting(id),
		liaisons: make(map[sys.Id]*liaison),
	}
	c.w.Init()
	go c.connectLoop()
	if c.algo.NeedBand() {
		go c.acceptLoop()
	}
	go c.clockLoop()
	return c
}

func (c *Compass0) String() string { 
	c.lk.Lock()
	defer c.lk.Unlock()

	return c.w.String()
}

func (c *Compass0) MarshalJSON() ([]byte, os.Error) { 
	c.lk.Lock()
	defer c.lk.Unlock()

	var w bytes.Buffer
	pr, err := c.algo.MarshalJSON()
	if err != nil {
		return nil, err
	}
	drv, err := c.w.MarshalJSON() 
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&w, "{\"Algorithm\":%s,\"Driver\":%s}", string(pr), string(drv))
	return w.Bytes(), nil
}

func (c *Compass0) clockLoop() {
	i := int64(0)
	for {
		if c.isHealthy() == nil {
			return
		}
		c.lk.Lock()
		c.algo.OnClock()
		c.lk.Unlock()
		if i % sweepFrequency == 0 {
			c.sweep()
		}
		time.Sleep(clockTick)
		i++
	}
}

func (c *Compass0) sweep() {
	c.lk.Lock()
	c.algo.Sweep()
	c.w.SetSources(c.algo.SourceCount())
	if c.algo.NeedBand() {
		for _,l := range c.liaisons {
			go c.write(l)
		}
	}
	c.lk.Unlock()
}

func (c *Compass0) write(l *liaison) {
	for {
		// Pre-send
		c.lk.Lock()
		l1,ok := c.liaisons[l.id]
		if c.d == nil || !ok || l1 != l {
			c.lk.Unlock()
			c.rem(l)
			return
		}

		msg := c.algo.PollSend(l.id)
		if msg == nil {
			c.lk.Unlock()
			return
		}
		c.lk.Unlock()

		// Send
		l.wlk.Lock()
		err := l.edc.Encode(msg)
		l.wlk.Unlock()

		// Post-send
		c.lk.Lock()
		if err != nil {
			c.lk.Unlock()
			c.rem(l)
			return
		}
		l2,ok := c.liaisons[l.id]
		if c.d == nil || !ok || l2 != l {
			c.lk.Unlock()
			c.rem(l)
			return
		}
		c.w.SetNeighborOutTraffic(l.id, l.edc.OutTraffic())
		c.lk.Unlock()
	}
}

func (c *Compass0) acceptLoop() {
	for {
		d := c.isHealthy()
		if d == nil {
			return
		}
		aid, rwc := d.Accept("compass0")
		if rwc == nil {
			continue
		}
		edc := newEncodeDecodeCloser(rwc)
		l := c.addLiaison(aid, edc)
		if l == nil {
			edc.Close()
			continue
		}
		if c.algo.NeedBand() {
			go c.read(l)
		}
	}
}

func (c *Compass0) addLiaison(id sys.Id, edc *EncodeDecodeCloser) *liaison {
	c.lk.Lock()
	defer c.lk.Unlock()
	_,ok := c.liaisons[id]
	if ok {
		return nil
	}
	l := &liaison{id: id, edc: edc}
	c.liaisons[id] = l
	c.algo.OnAddNeighbor(id)
	c.w.AddNeighbor(id)
	return l
}

func (c *Compass0) read(l *liaison) {
	for {
		c.lk.Lock()
		zmsg := c.algo.NewEmptyMsg()
		c.lk.Unlock()
		err := l.edc.Decode(zmsg)
		if err != nil {
			c.rem(l)
			return
		}

		c.lk.Lock()
		l1,ok := c.liaisons[l.id]
		if c.d == nil || !ok || l1 != l {
			c.lk.Unlock()
			c.rem(l)
			return
		}
		c.algo.OnReceive(l.id, zmsg)
		c.w.SetNeighborInTraffic(l.id, l.edc.InTraffic())
		c.w.SetSources(c.algo.SourceCount())
		c.lk.Unlock()
	}
	panic("unreach")
}

func (c *Compass0) rem(l *liaison) {
	//dbg.PrintStackTrace()
	c.lk.Lock()
	l1,ok := c.liaisons[l.id]
	if !ok || l1 != l {
		c.lk.Unlock()
		if l.edc != nil {
			l.edc.Close()
		}
		return
	}
	c.w.RemoveNeighbor(l.id)
	c.algo.OnRemoveNeighbor(l.id)
	c.liaisons[l.id] = nil, false
	c.lk.Unlock()
	if l.edc != nil {
		l.edc.Close()
	}
}

func (c *Compass0) haveId(id sys.Id) bool {
	c.lk.Lock()
	defer c.lk.Unlock()
	_,ok := c.liaisons[id]
	return ok
}

func (c *Compass0) connectLoop() {
	for {
		d := c.isHealthy()
		if d == nil {
			return
		}

		aid := d.WaitForArrival()
		if c.haveId(aid) {
			continue
		}

		rwc := d.Dial(aid, "compass0")
		if rwc == nil {
			continue
		}
		edc := newEncodeDecodeCloser(rwc)
		
		l := c.addLiaison(aid, edc)
		if l == nil {
			edc.Close()
			continue
		}

		if c.algo.NeedBand() {
			go c.read(l)
		}
	}
}

func (c *Compass0) isHealthy() dialer.Dialer {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.d
}

func (c *Compass0) ShutDown() {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.d = nil
}

func (c *Compass0) QueryQuantize(s,t sys.Id) *sys.Id {
	c.lk.Lock()
	defer c.lk.Unlock()
	r,err := c.algo.FlowStep(s,t)
	if err != nil {
		return nil
	}
	return &r
}
