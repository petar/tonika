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

import "math"

type AvgVar struct {
	count int64
	sum, sumsq float64
}

func (av *AvgVar) Init() {
	av.count = 0
	av.sum = 0.0
	av.sumsq = 0.0
}

func (av *AvgVar) Add(sample float64) {
	av.count++
	av.sum += sample
	av.sumsq += sample*sample
}

func (av *AvgVar) GetCount() int64 {
	return av.count
}

func (av *AvgVar) GetAvg() float64 {
	return av.sum / float64(av.count)
}

func (av *AvgVar) GetVar() float64 {
	a := av.GetAvg()
	return av.sumsq/float64(av.count) - a*a
}

func (av *AvgVar) GetStdDev() float64 {
	return math.Sqrt(av.GetVar())
}
