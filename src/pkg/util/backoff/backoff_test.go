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


package backoff

import (
	"fmt"
	"testing"
)

func TestBackoff(t *testing.T) {
	b := Backoff{2, 100, 1.4, 2, 0}
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	fmt.Printf("Backing off: %v\n", b.Inc())
	if b.Inc() != 79 {
		t.Errorf("backoff error\n")
	}
}
