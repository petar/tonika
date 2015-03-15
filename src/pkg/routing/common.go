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


// Some definitions common to all routing schemes
package routing

type Carrier         float64			// Routing arithmetic base 
type Epoch	     int64			// Type for representing epochs
type NeighborLiaison interface{}		// Abstract reference to a neighbor

func (c Carrier) F64() float64 {
	return float64(c)
}

func Max(x,y Carrier) Carrier {
    if x >= y {
        return x
    }
    return y
}
