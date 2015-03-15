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


package monitor

import (
	"fmt"
	"testing"
	"tonika/crypto"
)

func TestEncodeReport(t *testing.T) {
	const rep = "Hello World, ···"
	key := crypto.GenerateCipherMsgKey()
	buf,err := EncodeReport([]byte(rep), key.PubKey())
	if err != nil {
		t.Fatalf("encode: %s", err)
	}
	fmt.Printf("compressed size = %d\n", len(buf))
	rep2,err := DecodeReport(buf, 400, key)
	if err != nil {
		t.Fatalf("decode: %s", err)
	}
	rep1 := []byte(rep)
	if len(rep2) != len(rep1) {
		t.Fatalf("rep len")
	}
	for i := 0; i < len(rep1); i++ {
		if rep1[i] != rep2[i] {
			t.Errorf("byte mismatch at pos %d", i)
		}
	}
}
