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


package rsa64

import (
	"crypto/rsa"
	"fmt"
	"testing"
	"os"
)

func TestI64Enc(t *testing.T) {
	i1 := int64(-4355894345534)
	i2 := int64(9853975344)
	if di1, _ := bytesToInt64(int64ToBytes(i1)); di1 != i1 {
		t.Errorf("i->B->i error, %v\n", i1)
	}
	if di2, _ := bytesToInt64(int64ToBytes(i2)); di2 != i2 {
		t.Errorf("i->B->i error, %v\n", i2)
	}
}

func TestI32Enc(t *testing.T) {
	i1 := int32(30000)
	i2 := int32(-32034)
	if di1, _ := bytesToInt32(int32ToBytes(i1)); di1 != i1 {
		t.Errorf("i->B->i error, %v\n", i1)
	}
	if di2, _ := bytesToInt32(int32ToBytes(i2)); di2 != i2 {
		t.Errorf("i->B->i error, %v\n", i2)
	}
}

func TestRsaEnc(t *testing.T) {
	urandom, err := os.Open("/dev/urandom", os.O_RDONLY, 0)
	if err != nil {
		t.Errorf("failed to open /dev/urandom")
	}

	pk, err := rsa.GenerateKey(urandom, 256)
	if err != nil {
		t.Errorf("failed to generate key")
	}

	// Private key test

	priv64 := PrivToBase64(pk)
	fmt.Printf("PrivKey:\n%s\n\n", priv64)

	privdec, err := Base64ToPriv(priv64)
	if err != nil {
		t.Errorf("encoding doesn't read (%v)\n", err)
	}

	if !privEq(privdec, pk) {
		t.Errorf("private keys don't match\n")
	}

	// Public key test

	pub64 := PubToBase64(&pk.PublicKey)
	fmt.Printf("PubKey:\n%s\n\n", pub64)

	pubdec, err := Base64ToPub(pub64)
	if err != nil {
		t.Errorf("encoding doesn't read (%v)\n", err)
	}

	if !pubEq(pubdec, &pk.PublicKey) {
		t.Errorf("public keys don't match\n")
	}

}

func privEq(pk1 *rsa.PrivateKey, pk2 *rsa.PrivateKey) bool {
	if pk1.N.Cmp(pk2.N) != 0 {
		return false
	}
	if pk1.E != pk2.E {
		return false
	}
	if pk1.D.Cmp(pk2.D) != 0 {
		return false
	}
	if pk1.P.Cmp(pk2.P) != 0 {
		return false
	}
	if pk1.Q.Cmp(pk2.Q) != 0 {
		return false
	}
	return true
}

func pubEq(pk1 *rsa.PublicKey, pk2 *rsa.PublicKey) bool {
	if pk1.N.Cmp(pk2.N) != 0 {
		return false
	}
	if pk1.E != pk2.E {
		return false
	}
	return true
}
