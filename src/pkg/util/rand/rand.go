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


// Single source of randomness for the entire system, which depends on a
// initilizing seed. This should allow for easy reproduction of bugs, etc.
package rand

import (
    // "fmt"
    // "os"
    "rand"
    "time"
)

type _ThreadUnsafeSource struct {
    rand.Source
}

func (r *_ThreadUnsafeSource) Seed(seed int64) {
    r.Source.Seed(seed)
    // fmt.Fprintf(os.Stderr, "Random seed set to %x\n", seed)
}

func newSource() *_ThreadUnsafeSource {
    un := _ThreadUnsafeSource{}
    un.Source = rand.NewSource(1)
    return &un
}

func NewThreadUnsafe(seed int64) *rand.Rand {
    r := rand.New(newSource())
    r.Seed(seed)
    return r
}

func NewThreadUnsafeTimed() *rand.Rand {
	return NewThreadUnsafe(time.Nanoseconds())
}
