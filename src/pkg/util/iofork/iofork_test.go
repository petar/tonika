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


package iofork

import (
	"bytes"
	//"log"
	"os"
	"testing"
)

func TestIofork(t *testing.T) {
	b1 := bytes.NewBufferString("0123456789")
	r1,r2 := NewReader(b1, 2)

	p3 := make([]byte,3)
	n,err := r1.Read(p3)
	if err != nil || n != 3 || string(p3) != "012" {
		t.Fatalf("")
	}

	p4 := make([]byte,4)
	n,err = r2.Read(p4)
	if err != nil || n != 3 || string(p4[0:3]) != "012" {
		t.Fatalf("")
	}

	n,err = r1.Read(p3)
	if err != nil || n != 3 || string(p3) != "345" {
		t.Fatalf("")
	}

	n,err = r2.Read(p4)
	if err != nil || n != 3 || string(p4[0:3]) != "345" {
		t.Fatalf("")
	}

	n,err = r1.Read(p4)
	if err != nil || n != 4 || string(p4) != "6789" {
		t.Fatalf("")
	}

	p2 := make([]byte,2)
	n,err = r2.Read(p2)
	if err != nil || n != 2 || string(p2) != "67" {
		t.Fatalf("")
	}
	n,err = r2.Read(p2)
	if err != nil || n != 2 || string(p2) != "89" {
		t.Fatalf("")
	}
	n,err = r2.Read(p2)
	if err != os.EOF || n != 0 {
		t.Fatalf("")
	}
}
