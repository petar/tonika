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


// Fixed-size FIFO
package fixfifo

import (
	"fmt"
)

type FixFifo struct {
	slice []interface{}
	first int64
	last int64
}

func Make(capacity int) *FixFifo {
	return &FixFifo{ slice: make([]interface{}, capacity) }
}

func (ff *FixFifo) Slice() []interface{} {
	return ff.slice
}

func (ff *FixFifo) Len() int {
	return int(ff.last-ff.first)
}

func (ff *FixFifo) Full() bool {
	return ff.Len() >= len(ff.slice)
}

// Push an element in the FIFO and return true if there was space to push it
func (ff *FixFifo) Push(value interface{}) bool {
	if int(ff.last-ff.first) >= len(ff.slice) {
		return false
	}
	ff.slice[int(ff.last % int64(len(ff.slice)))] = value
	ff.last++
	return true
}

// Pop an element or return nil
func (ff *FixFifo) Pop() interface{} {
	if ff.last == ff.first {
		return nil
	}
	result := ff.slice[int(ff.first % int64(len(ff.slice)))]
	ff.slice[int(ff.first % int64(len(ff.slice)))] = nil
	ff.first++
	return result
}

func (ff *FixFifo) String() string {
	return fmt.Sprintf("FF(%v,%v)", ff.first, ff.last)
}
