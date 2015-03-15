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


package iopipe

import (
	"fmt"
	"io"
	"os"
)

type pipe struct {
	ch chan []byte
	partial []byte
}

func makePipe() *pipe {
	return &pipe{ make(chan []byte, 1), nil }
}

func (pp *pipe) Write(p []byte) (n int, err os.Error) {
	q := make([]byte, len(p))
	copy(q,p)
	pp.ch <- q
	return len(p), nil
}

func (pp *pipe) Read(p []byte) (n int, err os.Error) {
	for {
		if pp.partial != nil {
			n = copy(p, pp.partial)
			if n < 0 {
				panic("x")
			}
			pp.partial = pp.partial[n:]
			if len(pp.partial) == 0 {
				pp.partial = nil
			}
			return n, nil
		}
		pp.partial = <- pp.ch
	}
	panic("unreach")
}

type pipeEnd struct {
	r,w *pipe
}

func (pe *pipeEnd) Write(p []byte) (n int, err os.Error) {
	return pe.w.Write(p)
}

func (pe *pipeEnd) Read(p []byte) (n int, err os.Error) {
	return pe.r.Read(p)
}

func (pe *pipeEnd) Close() (err os.Error) {
	fmt.Printf("tonika/iopipe: Close not implemente\n")
	return nil
}

func MakePipe() (io.ReadWriteCloser, io.ReadWriteCloser) {
	a,b := makePipe(), makePipe()
	return &pipeEnd{a,b}, &pipeEnd{b,a}
}
