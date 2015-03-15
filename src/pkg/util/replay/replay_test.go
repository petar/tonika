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


package replay

import (
	//"log"
	"bytes"
	"testing"
)

func TestReplay(t *testing.T) {
	src := `abcdefgh`
	srcbuf := bytes.NewBufferString(src)
	replay := NewReader(srcbuf, 4)
	p := make([]byte, 3)
	n,err := replay.Read(p)
	if err != nil {
		t.Fatalf("bad read")
	}
	replay.Replay()
	replay.Read(p)
	replay.Read(p)
	if string(p) != `def` {
		t.Fatalf("mismatch")
	}
	n,err = replay.Read(p)
	if err != nil || n != 2 || string(p[0:2]) != `gh` {
		t.Fatalf("mismatch, %s")
	}
}
