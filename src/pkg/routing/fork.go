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
	"container/list"
	"math"
	"rand"
)

// Represents a list of abstract objects, each with an assigned mass.
type Fork interface {
	// Total mass of the elements in the list
	TotalMass() float64

	// Reset the iterator that guides the Next() method
	Reset()

	// Return the mass, value and presence of the next element
	Next() (prob float64, value interface{}, ok bool)
}

// This is a Fork that supports a "pick" operation, that selects the element
// currently pointed to by the iterator
//
type forkChoice interface {
	// Underlying fork
	Fork

	// Pick the current element in the iteration as the "choice"
	Pick()
}

func choose(fork forkChoice, random *rand.Rand) {
	pivot := random.Float64() * fork.TotalMass()
	fork.Reset()
	for prob, _, ok := fork.Next(); ok; prob, _, ok = fork.Next() {
		pivot -= prob
		if pivot <= 0.0 {
			fork.Pick()
			return
		}
	}
}

// A measure over a list of NeighborLiaison. Zero-value is an initialized instance.
type neighborFork struct {
	list   list.List
	total  float64
	iter   *list.Element
	choice NeighborLiaison
}

type neighborWithMass struct {
	mass     float64
	neighbor NeighborLiaison
}

// Add weighted elements to the measure
func (nf *neighborFork) Add(mass float64, nl NeighborLiaison) {
	if mass < 0 {
		panic("Negative mass!")
	}
	if nf.iter != nil {
		panic("Addition after sampling!")
	}
	nf.total += mass
	nf.list.PushBack(neighborWithMass{mass, nl})
}

// Compute total mass
func (nf *neighborFork) TotalMass() float64 {
	return nf.total
}

// Reset the iterator
func (nf *neighborFork) Reset() {
	nf.iter = nil
}

// Retrieve probability of next element
func (nf *neighborFork) Next() (prob float64, val interface{}, ok bool) {
	if nf.iter == nil {
		nf.iter = nf.list.Front()
	} else {
		nf.iter = nf.iter.Next()
	}
	if nf.iter == nil {
		return math.NaN(), nil, false
	}
	nam := nf.iter.Value.(neighborWithMass)
	return nam.mass, nam.neighbor, true
}

// Pick the current element as the probabilistic choice
func (nf *neighborFork) Pick() {
	if nf.choice != nil {
		panic("Picking more than once!")
	}
	v := nf.iter.Value.(neighborWithMass).neighbor
	nf.choice = v
}

// Returns the chosen element
func (nf *neighborFork) Chosen() NeighborLiaison {
	return nf.choice
}

func (nf *neighborFork) Quantize(random *rand.Rand) {
	choose(nf, random)
}
