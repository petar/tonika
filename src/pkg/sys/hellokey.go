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
	"os"
	"tonika/crypto"
	"tonika/util/rsa64"
)

type KeyHalves struct {
	bothKeys []byte
}

const (
	KeyHalvesLen = 40
	KeyHalvesBitLen = 8*KeyHalvesLen
)

func GenerateKeyHalves() *KeyHalves {
	urand := crypto.NewTimedRand()
	kh := &KeyHalves{}
	kh.bothKeys = make([]byte, KeyHalvesLen)
	n, _ := urand.Read(kh.bothKeys)
	if n != len(kh.bothKeys) {
		panic("d")
	}
	return kh
}

// TODO: Size checks
func BytesToKeyHalves(p []byte) (*KeyHalves, os.Error) {
	return &KeyHalves{p}, nil
}

func (kh *KeyHalves) Bytes() [] byte { return kh.bothKeys }

type U_KeyHalves struct {
	Halves []byte
}

// HelloKey is the type of keys used to establish symmetric encryption
// when new connections are established.
type HelloKey struct {
	rsa *rsa.PrivateKey
}

var HelloModulusBitLen = crypto.CalcModulusBitLen(KeyHalvesBitLen)

func GenerateHelloKey() *HelloKey {
	randr := crypto.NewTimedRC4Rand()
	pk, err := rsa.GenerateKey(randr, HelloModulusBitLen)
	if err != nil {
		panic("unable to generate Hello key")
	}
	return &HelloKey{pk}
}

func ParseHelloKey(s string) (hk *HelloKey, err os.Error) {
	rsapriv, err := rsa64.Base64ToPriv(s)
	if err != nil {
		return nil, err
	}
	return &HelloKey{rsapriv}, nil
}

func (hk *HelloKey) RSAPrivKey() *rsa.PrivateKey { return hk.rsa }

func (hk *HelloKey) RSAPubKey() *rsa.PublicKey { return &hk.rsa.PublicKey }

func (hk *HelloKey) String() string { return rsa64.PrivToBase64(hk.RSAPrivKey()) }

// U_HelloKey
type U_HelloKey struct {
	RSA U_RSAPubKey
}

func (hk *HelloKey) Proto() *U_HelloKey {
	p := &U_HelloKey{}
	p.RSA = *RSAPubKeyToProto(hk.RSAPubKey())
	return p
}

// HelloPubKey
type HelloPubKey struct {
	rsa *rsa.PublicKey
}

// TODO: Length checks
func UnprotoHelloPubKey(p *U_HelloKey) (*HelloPubKey, os.Error) {
	hk := &HelloPubKey{}
	hk.rsa = ProtoToRSAPubKey(&p.RSA)
	return hk, nil
}

func (hk *HelloPubKey) RSAPubKey() *rsa.PublicKey { return hk.rsa }

func (hk *HelloPubKey) String() string { return rsa64.PubToBase64(hk.RSAPubKey()) }
