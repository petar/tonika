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


// Uptime tracker data structure
package uptime

import (
	"fmt"
	"math"
	"time"
)

type Uptime struct {
	Halflife int64 // sec
	Bound float64 // rating does not plunge below 1/Bound or above Bound
	Rating float64
	LastUp int64 // ns
	LastDown int64 // ns
}

// Desired halfline in seconds
func Make(halflife int64, bound float64) Uptime {
	now := time.Nanoseconds()
	return Uptime{halflife,bound,1.0,now,now}
}

// Guess for uptime in ns
func (u *Uptime) Uptime() int64 {
	if u.LastUp > u.LastDown {
		return time.Nanoseconds() - u.LastUp
	}
	return 0
}

// Guess for downtime in ns
func (u *Uptime) Downtime() int64 {
	if u.LastDown > u.LastUp {
		return time.Nanoseconds() - u.LastDown
	}
	return 0
}

func maxf64(x,y float64) float64 {
	if x >= y {
		return x
	}
	return y
}

func minf64(x,y float64) float64 {
	if x <= y {
		return x
	}
	return y
}

// Indicate up time in this moment
func (u *Uptime) Up() float64 {

	now := time.Nanoseconds()
	delta := (now - maxi64(u.LastUp, u.LastDown))/sec
	u.LastUp = now

	base := math.Pow(0.5, -1.0 / float64(u.Halflife))
	u.Rating *= math.Pow(base, float64(delta))
	u.Rating = minf64(u.Bound, u.Rating)	

	return u.Rating

}

// Indicate down time in this moment
func (u *Uptime) Down() float64 {

	now := time.Nanoseconds()
	delta := (now - maxi64(u.LastUp, u.LastDown))/sec
	u.LastDown = now

	base := math.Pow(0.5, -1.0 / float64(u.Halflife))
	u.Rating *= math.Pow(base, float64(-delta))
	u.Rating = maxf64(1.0/u.Bound, u.Rating)	
	
	return u.Rating

}

func maxi64(x,y int64) int64 {
	if x >= y {
		return x
	}
	return y
}

const sec = 1000000000 // 1 sec in nanoseconds

func (u *Uptime) Pretty() string {
	return fmt.Sprintf("Rating: %g, Uptime: %ds, Downtime: %ds\n", u.Rating, u.Uptime()/sec, u.Downtime()/sec)
}
