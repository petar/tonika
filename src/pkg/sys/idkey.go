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


// *** SRC ***

package sys

import (
	"crypto/rsa"
	"crypto/sha256"
	"tonika/util/bytes"
	"tonika/util/rsa64"
)

func IdForKey(pk rsa.PublicKey) Id {

	// Write the Rsa64 representation of public key into Sha256
	pk64 := []byte(rsa64.PubToBase64(&pk))
	sha256 := sha256.New()
	for {
		n,err := sha256.Write(pk64)
		if err != nil {
			panic("sha256 malfunction")
		}
		if n == len(pk64) {
			break
		}
		pk64 = pk64[n:]
	}

	// Compute and fold the Sha256 hash
	h := sha256.Sum()
	if len(h) != 32 {
		panic("expecting 32 bytes")
	}
	
	for i := 1; i < 4; i++ {
		for j := 0; j < 8; j++ {
			h[j] ^= h[8*i+j]
		}
	}
	id64,err := bytes.BytesToInt64(h[0:8])
	if err != nil {
		panic("logic")
	}

	return Id(id64)
}

// Verifies that the Id corresponds to the public key
func VerifyKeyAndId(id Id, pubkey rsa.PublicKey) bool {
	return id == IdForKey(pubkey)
}
