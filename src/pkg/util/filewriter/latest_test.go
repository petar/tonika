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


package filewriter

import (
	"fmt"
	"testing"
)

type Struct struct {
	A,B,C int
}

func TestFileWriter(t *testing.T) {
	// Write
	lenc,err := MakeLatestFileEncoder("test")
	if err != nil {
		t.Fatalf("make: %s", err)
	}
	for i := 0; i < 10; i++ {
		s := &Struct{1,2,3}
		err := lenc.Encode(s)
		if err != nil {
			t.Fatalf("enc: %s", err)
		}
	}
	lenc.Close()

	// Read
	s := &Struct{}
	err := latestDecode("test", s)
	if err != nil {
		t.Fatalf("make: %s", err)
	}

	// Print
	fmt.Printf("Result: %v\n", s)
}
