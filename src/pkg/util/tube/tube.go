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


package tube

import (
	"bufio"
	"crypto/rc4"
	"gob"
	"io"
	"os"
)

type EncodeDecoder interface {
	Encode(e interface{}) os.Error
	Decode(e interface{}) os.Error
}

type ExposeBufioReader interface {
	ExposeBufioReader() *bufio.Reader
}

type TubedConn interface {
	EncodeDecoder
	ExposeBufioReader
	io.Reader
	io.Writer
	io.Closer
	Hijack() (io.ReadWriteCloser, *bufio.Reader)
}

// Tube is a wrapper around a net.Conn with a buffered reader
// and associated gob encoder and decoder
type Tube struct {
	io.ReadWriteCloser
	*bufio.Reader
	*gob.Encoder
	*gob.Decoder
}

func NewTube(rwc io.ReadWriteCloser, r *bufio.Reader) *Tube {
	if r == nil {
		r = bufio.NewReader(rwc)
	}
	return &Tube{
		ReadWriteCloser: rwc,
		Reader:          r,
		Encoder:         gob.NewEncoder(rwc),
		Decoder:         gob.NewDecoder(r),
	}
}

func (t *Tube) Read(p []byte) (n int, err os.Error) {
	return t.Reader.Read(p)
}

func (t *Tube) ExposeBufioReader() *bufio.Reader {
	return t.Reader
}

func (t *Tube) Hijack() (io.ReadWriteCloser, *bufio.Reader) {
	defer func() { 
		t.ReadWriteCloser, t.Reader = nil, nil
	}()
	return t.ReadWriteCloser, t.Reader
}

// Due to the complex concurrent ways in which tubes are used, a Tube must
// behave properly when Close is called after Hijack, or repeatedly.
func (t *Tube) Close() (err os.Error) {
	if t.ReadWriteCloser != nil {
		err = t.ReadWriteCloser.Close()
		t.ReadWriteCloser = nil
	}
	t.Reader = nil
	return err
}

// RC4Tube is a tube whose traffic is symmetrically encrypted with RC4
type RC4Tube struct {
	io.ReadWriteCloser
	*bufio.Reader
	*gob.Encoder
	*gob.Decoder

	rc, wc *rc4.Cipher
	wb     *bufio.Writer
}

func NewRC4Tube(t TubedConn, rk, wk []byte) *RC4Tube {
	rc, err := rc4.NewCipher(rk)
	if err != nil {
		panic("rc4tube")
	}
	wc, err := rc4.NewCipher(wk)
	if err != nil {
		panic("rc4tube")
	}
	rwc, rb := t.Hijack()
	t2 := &RC4Tube{
		ReadWriteCloser: rwc,
		Reader:          rb,
		rc:              rc,
		wc:              wc,
		wb:              bufio.NewWriter(rwc),
	}
	t2.Encoder = gob.NewEncoder(t2)
	t2.Decoder = gob.NewDecoder(t2)
	return t2
}

func (t *RC4Tube) Read(p []byte) (n int, err os.Error) {
	n, err = t.Reader.Read(p)
	if n > 0 {
		t.rc.XORKeyStream(p[0:n])
	}
	return n, err
}

func (t *RC4Tube) Write(p []byte) (int, os.Error) {
	for {
		n := min(t.wb.Available(), len(p))
		if n == 0 {
			panic("rc4tube, write")
		}
		t.wc.XORKeyStream(p[0:n])
		m, err := t.wb.Write(p[0:n])
		if m != n {
			panic("rc4tube, write2")
		}
		if err != nil {
			return m, err
		}
		err = t.wb.Flush()
		if err != nil {
			return m, err
		}
		p = p[m:]
		if len(p) == 0 {
			return m, nil
		}
	}
	panic("unreach")
}

func (t *RC4Tube) ExposeBufioReader() *bufio.Reader {
	return t.Reader
}

func (t *RC4Tube) Hijack() (io.ReadWriteCloser, *bufio.Reader) {
	panic("not supported")
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
