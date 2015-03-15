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


package vault

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"tonika/http"
	"tonika/math"
	"tonika/sys"
)

type watch struct {
	Fwds             map[sys.Id]*watchFwd
	OKMyReqs         int
	OKOnBehalfReqs   int
	ErrMyReqs        int
	ErrOnBehalfReqs  int
	HopsFwd, HopsBwd math.SmallDist
	TooManyHops      int
	lk               sync.Mutex
	fdlim            *http.FDLimiter
}

type watchFwd struct {
	MyInTraffic  int64
	MyOutTraffic int64
	OnBehalfInTraffic  int64
	OnBehalfOutTraffic int64
	LatencyPerByte math.AvgVar
}

func (v *watch) Init(fdlim *http.FDLimiter) {
	v.lk.Lock()
	defer v.lk.Unlock()
	v.Fwds = make(map[sys.Id]*watchFwd)
	v.HopsFwd.Init(maxHops)
	v.HopsBwd.Init(maxHops)
	v.fdlim = fdlim
}

func (v *watch) getFwd(id sys.Id) *watchFwd {
	f,ok := v.Fwds[id]
	if !ok {
		f = &watchFwd{}
		v.Fwds[id] = f
	}
	return f
}

func (v *watch) AddHopsFwd(k int) {
	v.lk.Lock()
	defer v.lk.Unlock()
	v.HopsFwd.Add(k)
}

func (v *watch) AddHopsBwd(k int) {
	v.lk.Lock()
	defer v.lk.Unlock()
	v.HopsBwd.Add(k)
}

func (v *watch) IncTooManyHops() { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.TooManyHops++ 
}

func (v *watch) IncOKMyReqs() { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.OKMyReqs++ 
}

func (v *watch) IncOKOnBehalfReqs() { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.OKOnBehalfReqs++ 
}

func (v *watch) IncErrMyReqs() { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.ErrMyReqs++ 
}

func (v *watch) IncErrOnBehalfReqs() { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.ErrOnBehalfReqs++ 
}

func (v *watch) IncFwdMyInTraffic(id sys.Id, amt int64) { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.getFwd(id).MyInTraffic += amt 
}

func (v *watch) IncFwdMyOutTraffic(id sys.Id, amt int64) { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.getFwd(id).MyOutTraffic += amt 
}

func (v *watch) IncFwdBehalfInTraffic(id sys.Id, amt int64) { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.getFwd(id).OnBehalfInTraffic += amt 
}

func (v *watch) IncFwdBehalfOutTraffic(id sys.Id, amt int64) { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.getFwd(id).OnBehalfOutTraffic += amt 
}

func (v *watch) RecFwdLatencyPerByte(id sys.Id, lpb float64) { 
	v.lk.Lock()
	defer v.lk.Unlock()
	v.getFwd(id).LatencyPerByte.Add(lpb) 
}

func (v *watch) String() string {
	v.lk.Lock()
	defer v.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "FileFD: %d/%d, OKMyReq: %d, OKBehalfReq: %d, " + 
		"ErrMyReq: %d, ErrBehalfReq: %d, TooManyHops: %d, Fwds:\n",
		v.fdlim.LockCount(), v.fdlim.Limit(),
		v.OKMyReqs, v.OKOnBehalfReqs, v.ErrMyReqs, v.ErrOnBehalfReqs, v.TooManyHops)
	for id,f := range v.Fwds {
		fmt.Fprintf(&w, "%s", f.String(id))
	}
	fmt.Fprintf(&w,"HopsFwd: ")
	for i,k := range v.HopsFwd.Int64Array() {
		fmt.Fprintf(&w,"|%d|=%d, ",i,k)
	}
	fmt.Fprintf(&w,"\nHopsBwd: ")
	for i,k := range v.HopsBwd.Int64Array() {
		fmt.Fprintf(&w,"|%d|=%d, ",i,k)
	}
	fmt.Fprintf(&w,"\n")
	return w.String()
}

func (v *watch) MarshalJSON() ([]byte, os.Error) {
	v.lk.Lock()
	defer v.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "{\"FileFD\":%d,\"FileFDLim\":%d,\"OKMyReq\":%d,\"OKBehalfReq\":%d,"+
		"\"ErrMyReq\":%d,\"ErrBehalfReq\":%d,\"TooManyHops\":%d,\"Fwds\":[",
		v.fdlim.LockCount(), v.fdlim.Limit(),
		v.OKMyReqs, v.OKOnBehalfReqs, v.ErrMyReqs, v.ErrOnBehalfReqs, v.TooManyHops)
	comma := false
	for id,f := range v.Fwds {
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w, "%s", f.toJSON(id))
		comma = true
	}
	fmt.Fprintf(&w,"],\"HopsFwd\":[")
	comma = false
	for _,k := range v.HopsFwd.Int64Array() {
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w,"%d",k)
		comma = true
	}
	fmt.Fprintf(&w,"],\"HopsBwd\":[")
	comma = false
	for _,k := range v.HopsBwd.Int64Array() {
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w,"%d",k)
		comma = true
	}
	fmt.Fprintf(&w,"]}")
	return w.Bytes(), nil
}

func (f *watchFwd) String(id sys.Id) string {
	var w bytes.Buffer
	fmt.Fprintf(&w, 
		"    Id:%s\n" +
		"      MyInTraffic: %d, MyOutTraffic: %d\n" +
		"      BehalfInTraffic: %d, BehalfOutTraffic: %d, LatencyPerByte: %g\n",
		id.Eye(), f.MyInTraffic, f.MyOutTraffic, 
		f.OnBehalfInTraffic, f.OnBehalfOutTraffic, f.LatencyPerByte.GetAvg())
	return w.String()
}

func (f *watchFwd) toJSON(id sys.Id) string {
	var w bytes.Buffer
	fmt.Fprintf(&w, 
		"{\"Id\":%s,\"MyInTraffic\":%d,\"MyOutTraffic\":%d," +
		"\"BehalfInTraffic\":%d,\"BehalfOutTraffic\":%d",
		id.ToJSON(), f.MyInTraffic, f.MyOutTraffic, 
		f.OnBehalfInTraffic, f.OnBehalfOutTraffic)
	lat := f.LatencyPerByte.GetAvg()
	if math.IsNum(lat) {
		fmt.Fprintf(&w, ",\"LatencyPerByte\":%g}", lat)
	} else {
		fmt.Fprintf(&w, "}")
	}
	return w.String()
}
