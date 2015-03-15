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


// Convert Reader/Writers to channels and vice-versa
package ioch

import (
	"bufio"
	"io"
	"os"
)

// This routine converts a reader r into a notification channel ch and a new
// reader rr.  The channel receives nil whenever data is available for reading,
// or an os.Error if r has gone corrupt. Actual reading should be performed
// through rr.
func MakeReaderNotifier(r io.Reader) (ch <-chan os.Error, rr io.Reader) {
	c := make(chan os.Error)
	br := bufio.NewReader(r)
	go func(){
		_, err := br.ReadByte()
		if err != nil {
			c <- err
			close(c)
			return
		}
		err = br.UnreadByte()
		if err != nil {
			c <- err
			close(c)
			return
		}
		c <- nil
	}()
	return c,rr
}

type Result struct {
	N int
	Err os.Error
}

func MakeChanForReader(r io.Reader) chan interface{} {
	c := make(chan interface{})
	go func(){
		for {
			_buf := <-c
			if _buf == nil || closed(c) {
				return
			}
			buf := _buf.([]byte)
			n,err := r.Read(buf)
			c<- Result{n,err}
		}
	}()
	return c
}

func MakeChanForWriter(w io.Writer) chan interface{} {
	c := make(chan interface{})
	go func(){
		for {
			_buf := <-c
			if _buf == nil || closed(c) {
				return
			}
			buf := _buf.([]byte)
			n,err := w.Write(buf)
			c<- Result{n,err}
		}
	}()
	return c
}
