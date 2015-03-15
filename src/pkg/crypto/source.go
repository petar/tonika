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
	"crypto/rc4"
	"io"
	"os"
	"rand"
	"time"
	"tonika/util/bytes"
)

// OS random source from /dev/urandom
var urand *os.File

func GetUrandom() (r io.Reader, err os.Error) {
	if urand != nil {
		return urand, nil
	}
	urand, err := os.Open("/dev/urandom", os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return urand, nil 
}

func NewRand() *rand.Rand { return rand.New(rand.NewSource(time.Nanoseconds())) }

// Random source using Go's rand, keyed by time
type TimedRand struct {
	*rand.Rand
}

func NewTimedRand() TimedRand {
	return TimedRand{rand.New(rand.NewSource(time.Nanoseconds()))}
}

func (tr TimedRand) Read(p []byte) (n int, err os.Error) {
	for i,_ := range p {
		p[i] = byte(tr.Rand.Int())
	}
	return len(p),nil
}

// Random source using RC4, keyed by time
type TimedRC4Rand struct {
	*rc4.Cipher
}

func NewTimedRC4Rand() TimedRC4Rand {
	ciph, err := rc4.NewCipher(bytes.Int64ToBytes(time.Nanoseconds()))
	if err != nil { 
		panic("rc4 gen error")
	}
	return TimedRC4Rand{ciph}
}

func (tr TimedRC4Rand) Read(p []byte) (n int, err os.Error) {
	for i,_ := range p {
		p[i] = 0
	}
	tr.XORKeyStream(p)
	return len(p),nil
}

// Reader that always returns one
type OnesReader struct{}

func (OnesReader) Read(p []byte) (n int, err os.Error) {
	for i := 0; i < len(p); i++ {
		p[i] = 1
	}
	return len(p), nil
}
