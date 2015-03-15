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


// This package implements some math routines like matrix arithmetic, which are
// needed for testing the behavior of the routing code. The implementations
// here of things like matrix multiplication (and such) are inefficient, but
// that's fine.

package math

type SmallDist struct {
	samples []int64	
}

func (s *SmallDist) Init(n int) {
	s.samples = make([]int64, n)	
}

func (s *SmallDist) Len() int { return len(s.samples) }

func (s *SmallDist) Add(i int) {
	if i >= len(s.samples) {
		s.samples[len(s.samples)-1]++
	} else {
		s.samples[i]++
	}
}

func (s *SmallDist) Int64Array() []int64 {
	return s.samples
}
