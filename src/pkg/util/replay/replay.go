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


// An io.Reader wrapper that remembers content and can be asked to replay
package replay

import (
	"io"
	//"log"
	"os"
)

type ReplayReader interface {
	io.Reader
	Rewind()
}

func NewReader(r io.Reader, l0 int) (rr ReplayReader) {
	rr = &replayReader{
		under: r,
		mem: make([]byte,0,l0),
	}
	return
}

type replayReader struct {
	under io.Reader
	mem   []byte
	err   os.Error
	re    bool
}

func (rr *replayReader) Rewind() {
	if rr.re {
		panic("replay")
	}
	rr.re = true
}

func (rr *replayReader) Read(p []byte) (int, os.Error) {
	switch rr.re {
	case false: // play mode
		if rr.err != nil {
			return -1,rr.err
		}
		n := min(len(p),cap(rr.mem)-len(rr.mem))
		if n == 0 {
			// return an IO error if replay buffer size exceeded
			return -1,os.EIO
		}
		n,err := rr.under.Read(p[0:n])
		if err != nil {
			if n > 0 {
				panic("replay")
			}
			rr.err = err
			return -1,err	
		}
		if copy(rr.mem[len(rr.mem):cap(rr.mem)], p[0:n]) != n {
			panic("replay")
		}
		rr.mem = rr.mem[0:len(rr.mem)+n]
		return n,nil
	case true: // replay mode
		if len(rr.mem) > 0 {
			n := min(len(rr.mem),len(p))
			copy(p,rr.mem[0:n])
			rr.mem = rr.mem[n:]
			return n,nil
		}
		if rr.err != nil {
			return -1,rr.err
		}
		return rr.under.Read(p)
	}
	panic("replay")
}

func min(x,y int) int {
	if x <= y {
		return x
	}
	return y
}
