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


package routing

import (
	"testing"
	// "rand"
	"fmt"
	"math"
	"tonika/util/rand"
)

func abs(n int32) int32 {
	if n >= 0 {
		return n
	}
	return -n
}

func TestFork(t *testing.T) {
	r := rand.NewThreadUnsafe(1)
	stats := [5]int32{}
	n := 150000
	for i := 0; i < n; i++ {
		nm := new(neighborFork)
		nm.Add(1, 1)
		nm.Add(2, 2)
		nm.Add(3, 3)
		nm.Add(4, 4)
		nm.Add(5, 5)
		nm.Quantize(r)
		stats[(nm.Chosen()).(int)-1]++
	}
	fmt.Printf("Frequencies %v\n", stats)
	var a int32 = int32(n / 15)
	var d int32 = int32(math.Sqrt(float64(n)))
	fmt.Printf("Expected [")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", (int32(i)+1)*a)
	}
	fmt.Printf("]\n")
	fmt.Printf("Acceptable deviation %d\n", d)
	for i := 0; i < 5; i++ {
		if abs(stats[i]-(int32(i)+1)*a) > d {
			t.Errorf("Frequency out of bounds!")
		}
	}
}
