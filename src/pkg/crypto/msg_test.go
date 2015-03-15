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


package crypto

import (
	"testing"
	"tonika/sys"
)

func TestEncipherMsg(t *testing.T) {
	plain := make([]byte, 100)
	urand := sys.NewTimedRand()
	n, _ := urand.Read(plain)
	if n != len(plain) {
		t.Fatalf("plain text randomization")
	}
	priv := GenerateCipherMsgKey()
	cipher,err := EncipherMsg(plain, priv.PubKey())
	if err != nil {
		t.Fatalf("encipher: %s\n", err)
	}
	plain2,err := DecipherMsg(cipher, priv)
	if err != nil {
		t.Fatalf("decipher: %s\n", err)
	}
	if len(plain2) != len(plain) {
		t.Fatalf("plain text len mismatch\n")
	}
	for i := 0; i < len(plain); i++ {
		if plain[i] != plain2[i] {
			t.Fatalf("mismatch")
		}
	}
}
