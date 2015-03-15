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


// Profiling primitives
package prof

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sort"
	"sync"
	"time"
	"tonika/math"
)

type prof struct {
	// Measures time spent in a section of code
	secs map[string]*section
	lk sync.Mutex
}

type section struct {
	dur     math.AvgVar
	nactive int64
	lk      sync.Mutex
}

type profHandle struct {
	sec  *section
	t0   int64
}

func newProf() *prof {
	return &prof{ secs: make(map[string]*section) }
}

func (p *prof) SecEnter(sd int) *profHandle {
	_,file,line,ok := runtime.Caller(1+sd)
	if !ok {
		panic("prof, no source info")
	}
	return p.SecEnterName(file + ":" + strconv.Itoa(line))
}

func (p *prof) SecEnterName(name string) *profHandle {
	s := p.fetch(name)
	t0 := time.Nanoseconds()
	s.lk.Lock()
	s.nactive++
	s.lk.Unlock()
	return &profHandle{ sec: s, t0: t0 }	
}

func (p *prof) SecLeave(f *profHandle) {
	f.sec.lk.Lock()
	f.sec.dur.Add(float64(time.Nanoseconds() - f.t0))
	f.sec.nactive--
	if f.sec.nactive < 0 {
		panic("profiler section negative entry count")
	}
	f.sec.lk.Unlock()
}

func (p *prof) fetch(name string) *section {
	p.lk.Lock()
	defer p.lk.Unlock()
	s,ok := p.secs[name]
	if !ok {
		s = &section{}
		p.secs[name] = s
	}
	return s
}

type sectionPrinter struct {
	name string
	avg,stddev float64
	count int64
	nactive int64
}

type printer []*sectionPrinter

func (p printer) Len() int { return len(([]*sectionPrinter)(p)) }

func (p printer) Less(i, j int) bool { return p[i].avg > p[j].avg }

func (p printer) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type timeDiv struct {
	ubound int64 // upper bound for membership
	name string
}

func (td *timeDiv) Ni(ns float64) bool { return ns <= float64(td.ubound) }

var timeDivs = []*timeDiv{
	&timeDiv{ 1<<62, "===== SLOW =========" },
	&timeDiv{ 1e9, "===== 1 sec ========" },
	&timeDiv{ 1e8, "===== 1/10 sec =====" },
	&timeDiv{ 1e7, "===== 1/100 sec ====" },
	&timeDiv{ 1e6, "===== 1/1000 sec ===" },
	&timeDiv{ 1e5, "===== FAST =========" },
	&timeDiv{ -1, "This should NOT print!" },
}

func (p *prof) String() string {

	p.lk.Lock()
	pr := make(printer, len(p.secs))
	i := 0
	for k,sec := range p.secs {
		sec.lk.Lock()
		pr[i] = &sectionPrinter {
			name: k,
			avg: sec.dur.GetAvg(),
			stddev: sec.dur.GetStdDev(),
			count: sec.dur.GetCount(),
			nactive: sec.nactive,
		}
		sec.lk.Unlock()
		i++
	}
	p.lk.Unlock()

	if len(pr) == 0 {
		return ""
	}
	sort.Sort(pr)
	var b bytes.Buffer
	td := -1
	for _,t := range pr {
		var j int
		for j = 1; timeDivs[td+j].Ni(t.avg); j++ {}
		if j > 1 {
			td += j-1
			b.WriteString("\n"+timeDivs[td].name+"\n")
		}
		b.WriteString(fmt.Sprintf("%20s:\n\t%e ±(%e)ns,\t#(%d),\t?(%d)\n", 
			t.name, t.avg, t.stddev, t.count, t.nactive))
	}
	return b.String()
}

var global = newProf()

func SecEnter() *profHandle { return global.SecEnter(1) }

func SecEnterName(name string) *profHandle { return global.SecEnterName(name) }

func SecLeave(f *profHandle) { global.SecLeave(f) }

func String() string { return global.String() }
