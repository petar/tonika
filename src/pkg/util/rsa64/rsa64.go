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

// This package encodes/decodes an RSA private key, using base64 encoding and
// an in-house overall format.
package rsa64

import (
	"big"
	"crypto/rsa"
	"encoding/base64"
	"os"
	"strings"
	"unsafe"
)

// Returns the integer in big-endian byte order
func int64ToBytes(s int64) []byte {
	u := uint64(s)
	l := unsafe.Sizeof(u)
	b := make([]byte, l)
	for i := 0; i < l; i++ {
		b[i] = byte((u >> uint(8*(l-i-1))) & 0xff)
	}
	return b
}

// Invert int64ToBytes
func bytesToInt64(b []byte) (int64, os.Error) {
	u := uint64(0)
	l := unsafe.Sizeof(u)
	if len(b) != l { 
		return 0, os.ErrorString("bad length")
	}
	for i := 0; i < l; i++ {
		u |= uint64(b[i]) << uint(8*(l-i-1))
	}
	return int64(u), nil
}

// Returns the integer in big-endian byte order
func int32ToBytes(s int32) []byte {
	u := uint32(s)
	l := unsafe.Sizeof(u)
	b := make([]byte, l)
	for i := 0; i < l; i++ {
		b[i] = byte((u >> uint(8*(l-i-1))) & 0xff)
	}
	return b
}

// Invert int32ToBytes
func bytesToInt32(b []byte) (int32, os.Error) {
	u := uint32(0)
	l := unsafe.Sizeof(u)
	if len(b) != l { 
		return 0, os.ErrorString("bad length")
	}
	for i := 0; i < l; i++ {
		u |= uint32(b[i]) << uint(8*(l-i-1))
	}
	return int32(u), nil
}

func firstComma(b []byte) int {
	for i := 0; i < len(b); i++ {
		if b[i] == ',' {
			return i
		}
	}
	return -1
}

// This is the alphabet used for the CRC
const alphaCRC = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

// appendCRC attaches a CRC character at position p[dlen],
// based on the contents of p[0:dlen-1]
func appendCRC(p []byte, dlen int) {
	crc := 0
	for i := 0; i < dlen; i++ {
		crc = (crc + int(p[i])) % len(alphaCRC)
	}
	p[dlen] = alphaCRC[crc]
}

// verifyCRC verifies that the last character of s, which is a CRC,
// checks correctly against the prior contents.
func verifyCRC(s string) bool {
	if len(s) < 1 {
		return false
	}
	crc := 0
	for i := 0; i < len(s)-1; i++ {
		crc = (crc + int(s[i])) % len(alphaCRC)
	}
	index := strings.Index(alphaCRC, string(s[len(s)-1]))
	return index == crc
}

// Convert string to an RSA public key
func Base64ToPub (s string) (*rsa.PublicKey, os.Error) {
	if len(s) == 0 {
		return nil, nil
	}
	if !verifyCRC(s) {
		return nil, nil
	}
	s = s[0:len(s)-1]

	enc := base64.StdEncoding
	pk := rsa.PublicKey{}

	buf := make([]byte, 4096) // shoud be big enough
	src := []byte(s)
	k := -1

	// N
	if k = firstComma(src); k < 0 {
		return nil, os.ErrorString("missing delimiter")
	}
	n, err := enc.Decode(buf, src[0:k])
	if err != nil {
		return nil, err
	}
	pk.N = &big.Int{}
	pk.N.SetBytes(buf[0:n])
	src = src[k+1:]
	
	// E
	n, err = enc.Decode(buf, src)
	if err != nil {
		return nil, err
	}
	pke64, err := bytesToInt64(buf[0:n])
	if err != nil {
		return nil, err
	}
	pk.E = int(pke64)

	return &pk, nil
}

// Convert string to an RSA private key
func Base64ToPriv (s string) (*rsa.PrivateKey, os.Error) {
	if len(s) == 0 {
		return nil, nil
	}
	if !verifyCRC(s) {
		return nil, nil
	}
	s = s[0:len(s)-1]

	enc := base64.StdEncoding
	pk := rsa.PrivateKey{}

	buf := make([]byte, 4096) // shoud be big enough
	src := []byte(s)
	k := -1

	// N
	if k = firstComma(src); k < 0 {
		return nil, os.ErrorString("missing delimiter")
	}
	n, err := enc.Decode(buf, src[0:k])
	if err != nil {
		return nil, err
	}
	pk.N = &big.Int{}
	pk.N.SetBytes(buf[0:n])
	src = src[k+1:]
	
	// E
	if k = firstComma(src); k < 0 {
		return nil, os.ErrorString("missing delimiter")
	}
	n, err = enc.Decode(buf, src[0:k])
	if err != nil {
		return nil, err
	}
	pke64, err := bytesToInt64(buf[0:n])
	if err != nil {
		return nil, err
	}
	pk.E = int(pke64)
	src = src[k+1:]
	
	// D
	if k = firstComma(src); k < 0 {
		return nil, os.ErrorString("missing delimiter")
	}
	n, err = enc.Decode(buf, src[0:k])
	if err != nil {
		return nil, err
	}
	pk.D = &big.Int{}
	pk.D.SetBytes(buf[0:n])
	src = src[k+1:]
	
	// P
	if k = firstComma(src); k < 0 {
		return nil, os.ErrorString("missing delimiter")
	}
	n, err = enc.Decode(buf, src[0:k])
	if err != nil {
		return nil, err
	}
	pk.P = &big.Int{}
	pk.P.SetBytes(buf[0:n])
	src = src[k+1:]
	
	// Q
	n, err = enc.Decode(buf, src)
	if err != nil {
		return nil, err
	}
	pk.Q = &big.Int{}
	pk.Q.SetBytes(buf[0:n])

	return &pk, nil
}

// Convert RSA public key to string
func PubToBase64 (pk *rsa.PublicKey) string {
	if pk == nil {
		return ""
	}

	enc := base64.StdEncoding
	
	N := pk.N
	E := pk.E // int

	bN := N.Bytes()
	bE := int64ToBytes(int64(E))

	lN := enc.EncodedLen(len(bN))
	lE := enc.EncodedLen(len(bE))

	result := make([]byte, lN + 1 + lE + 1)

	t := result
	enc.Encode(t, bN)
	t[lN] = ','
	t = t[lN+1:]
	enc.Encode(t, bE)

	appendCRC(result, len(result)-1)
	return string(result)
}

// Convert RSA private key to a string
func PrivToBase64 (pk *rsa.PrivateKey) string {
	if pk == nil {
		return ""
	}

	enc := base64.StdEncoding
	
	N := pk.PublicKey.N
	E := pk.PublicKey.E // int
	D := pk.D
	P := pk.P
	Q := pk.Q

	bN := N.Bytes()
	bE := int64ToBytes(int64(E))
	bD := D.Bytes()
	bP := P.Bytes()
	bQ := Q.Bytes()

	lN := enc.EncodedLen(len(bN))
	lE := enc.EncodedLen(len(bE))
	lD := enc.EncodedLen(len(bD))
	lP := enc.EncodedLen(len(bP))
	lQ := enc.EncodedLen(len(bQ))

	result := make([]byte, lN + lE + lD + lP + lQ + 4 + 1)

	t := result
	enc.Encode(t, bN)
	t[lN] = ','
	t = t[lN+1:]
	enc.Encode(t, bE)
	t[lE] = ','
	t = t[lE+1:]
	enc.Encode(t, bD)
	t[lD] = ','
	t = t[lD+1:]
	enc.Encode(t, bP)
	t[lP] = ','
	t = t[lP+1:]
	enc.Encode(t, bQ)

	appendCRC(result, len(result)-1)
	return string(result)
}

