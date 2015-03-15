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

// *** SRC ***

import (
	"bytes"
	"crypto/rc4"
	"crypto/rsa"
	"crypto/sha1"
	//"fmt"
	"gob"
	"os"
	"tonika/util/rsa64"
)

type U_CipherMsg struct {
	Seed []byte // RSA-encrypted Seed for the RC4 cipher
	Text []byte // Plaintext, encrypted using RC4 seeded by Seed
}

// CipherMsgKey
type CipherMsgKey struct {
	rsa *rsa.PrivateKey
}

const cipherMsgSeedLen = 20
const cipherMsgSeedBitLen = cipherMsgSeedLen*8

var cipherMsgModulusBitLen = CalcModulusBitLen(cipherMsgSeedBitLen)

func GenerateCipherMsgKey() *CipherMsgKey {
	urand := NewTimedRand()
	key, err := rsa.GenerateKey(urand, cipherMsgModulusBitLen)
	if err != nil {
		panic("unable to generate cipher key")
	}
	return &CipherMsgKey{key}
}

// TODO: Add length checks
func ParseCipherMsgKey(s string) (k *CipherMsgKey, err os.Error) {
	rsapriv, err := rsa64.Base64ToPriv(s)
	if err != nil {
		return nil, err
	}
	return &CipherMsgKey{rsapriv}, nil
}

func (cmk *CipherMsgKey) String() string {
	return rsa64.PrivToBase64(cmk.rsa)
}

func (cmk *CipherMsgKey) PubKey() *CipherMsgPubKey {
	return &CipherMsgPubKey{&cmk.rsa.PublicKey}
}

// CipherMsgPubKey
type CipherMsgPubKey struct {
	rsa *rsa.PublicKey
}

// TODO: Add length checks
func ParseCipherMsgPubKey(s string) (k *CipherMsgPubKey, err os.Error) {
	rsapub, err := rsa64.Base64ToPub(s)
	if err != nil {
		return nil, err
	}
	return &CipherMsgPubKey{rsapub}, nil
}

func (cmpk *CipherMsgPubKey) String() string {
	return rsa64.PubToBase64(cmpk.rsa)
}

// Encipher a message
func EncipherMsg(plaintext []byte, pubkey *CipherMsgPubKey) ([]byte, os.Error) {
	msg := &U_CipherMsg{
		Text: make([]byte, len(plaintext)),
	}
	n := copy(msg.Text, plaintext)
	if n != len(plaintext) {
		panic("crypto, copy text")
	}
	urand := NewTimedRand()
	seed := make([]byte, cipherMsgSeedLen)
	n, _ = urand.Read(seed)
	if n != len(seed) {
		panic("crypto,gen seed")
	}
	cseed, err := EncryptShortMsg(pubkey.rsa, seed, []byte(""))
	if err != nil {
		return nil, err
	}
	msg.Seed = cseed
	rc, err := rc4.NewCipher(seed)
	if err != nil {
		panic("rc4tube")
	}
	rc.XORKeyStream(msg.Text)
	var w bytes.Buffer
	enc := gob.NewEncoder(&w)
	err = enc.Encode(msg)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecipherMsg(ciphertext []byte, privkey *CipherMsgKey) ([]byte, os.Error) {
	r := bytes.NewBuffer(ciphertext)
	dec := gob.NewDecoder(r)
	msg := &U_CipherMsg{}
	err := dec.Decode(msg)
	if err != nil {
		return nil, err
	}
	seed,err := DecryptShortMsg(privkey.rsa, msg.Seed, []byte(""))
	if err != nil {
		return nil, err
	}
	rc, err := rc4.NewCipher(seed)
	if err != nil {
		return nil, err
	}
	rc.XORKeyStream(msg.Text)
	return msg.Text, nil
}

const sha1Len = 20
const sha1BitLen = sha1Len*8

// CalcModulusLen returns the minimum RSA public modulus length, required
// for encrypting a message with size msgbitlen bits, using the 
// EncryptShortMsg routine.
func CalcModulusBitLen(msgbitlen int) int {
	return msgbitlen + 2*sha1BitLen + 2
}

func EncryptShortMsg(pubkey *rsa.PublicKey, plaintext, label []byte) (ciphertext []byte, err os.Error) {
	urand := NewTimedRand()
	return rsa.EncryptOAEP(sha1.New(), urand, pubkey, plaintext, label)
}

func DecryptShortMsg(privkey *rsa.PrivateKey, ciphertext, label []byte) (plaintext []byte, err os.Error) {
	return rsa.DecryptOAEP(sha1.New(), nil, privkey, ciphertext, label)
}
