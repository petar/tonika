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
	"crypto/sha1"
	"os"
	"tonika/crypto"
	"tonika/util/rsa64"
)

// It is essential that all key-types below be immutable, since
// they are passed on from FE to CORE, or CORE to DIALER. They change
// hands a lot and the internal logic would break if any of these systems
// could change these objects.


// SigKey is the public+private signature key.
type SigKey struct {
	rsa *rsa.PrivateKey
}

const SignatureModulusLen = 64 // in bytes, = 512 bits

// Generates an RSA private key and computes the corresponding 64-bit
// Id for it, using Id=Fold(Sha256(Rsa64(PublicKey)))
//
// TODO:
//   (*) Decide on default signature key size
//   (*) Allow custom-size private keys, and/or
func GenerateSigKey() *SigKey {
	randr := crypto.NewTimedRand()
	rsapriv, err := rsa.GenerateKey(randr, SignatureModulusLen*8)
	if err != nil {
		panic("unable to generate RSA key")
	}
	return &SigKey{rsapriv}
}

// TODO: Add length checks
func ParseSigKey(s string) (sk *SigKey, err os.Error) {
	rsapriv, err := rsa64.Base64ToPriv(s)
	if err != nil {
		return nil, err
	}
	return &SigKey{rsapriv}, nil
}

func (sk *SigKey) RsaPrivKey() *rsa.PrivateKey { return sk.rsa }

func (sk *SigKey) RsaPubKey() *rsa.PublicKey { return &sk.rsa.PublicKey }

func (sk *SigKey) String() string {
	return rsa64.PrivToBase64(sk.RsaPrivKey())
}

func (sk *SigKey) PubKey() *SigPubKey {
	return &SigPubKey{&sk.rsa.PublicKey}
}

func (sk *SigKey) Id() Id {
	return IdForKey(sk.rsa.PublicKey)
}

func (sk *SigKey) Sign(msg []byte) ([]byte, os.Error) {
	// Hash the message
	hash := sha1.New()
	n,err := hash.Write(msg)
	if err != nil || n != len(msg) {
		return nil, err
	}
	hashed := hash.Sum()

	// Sign the message
	urand := crypto.NewTimedRand()
	s, err := rsa.SignPKCS1v15(urand, sk.RsaPrivKey(), rsa.HashSHA1, hashed)
	if err != nil {
		return nil, err
	}
	
	return s, nil
}

// SigPubKey is the public part of the signature key.
type SigPubKey struct {
	rsa *rsa.PublicKey
}

// TODO: Add length checks
func ParseSigPubKey(s string) (sk *SigPubKey, err os.Error) {
	rsapub, err := rsa64.Base64ToPub(s)
	if err != nil {
		return nil, err
	}
	return &SigPubKey{rsapub}, nil
}

func (sk *SigPubKey) RsaPubKey() *rsa.PublicKey { return sk.rsa }

func (sk *SigPubKey) String() string {
	return rsa64.PubToBase64(sk.RsaPubKey())
}

func (sk *SigPubKey) Id() Id {
	return IdForKey(*sk.rsa)
}

func (sk *SigPubKey) Verify(msg, sign []byte) os.Error {
	// Hash message
	hash := sha1.New()
	n,err := hash.Write(msg)
	if err != nil || n != len(msg) {
		return err
	}
	hashed := hash.Sum()

	// Verify hashed against sign
	return rsa.VerifyPKCS1v15(sk.RsaPubKey(), rsa.HashSHA1, hashed, sign)
}

func GenerateSigChallange() []byte {
	urand := crypto.NewTimedRand()
	ch := make([]byte, 20)  // we are using SHA1 hash for RSA signature
	n,err := urand.Read(ch)
	if err != nil || n != len(ch) {
		panic("sys, chall")
	}
	return ch
}
