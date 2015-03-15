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


// Exponential back-off calculator
package backoff

import (
	"math"
)

type Backoff struct {
	Lo,Hi int64
	Ratio float64
	Current int64
	Attempt int
}

func (b *Backoff) Reset() int64 {
	b.Current = b.Lo
	b.Attempt = 0
	return b.Current
}

func (b *Backoff) Inc() int64 {
	t := int64(math.Ceil(float64(b.Current)*b.Ratio))
	if t <= b.Hi {
		b.Current = t
	}
	b.Attempt++
	return b.Current
}
