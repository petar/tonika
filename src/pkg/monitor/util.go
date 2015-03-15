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
	"bytes"
	"compress/zlib"
	"os"
	"tonika/crypto"
)

func EncodeReport(p []byte, key *crypto.CipherMsgPubKey) ([]byte, os.Error) {
	zipbuf := zip(p)
	cipher, err := crypto.EncipherMsg(zipbuf, key)
	if err != nil {
		return nil, err
	}
	return cipher, nil
}

func DecodeReport(p []byte, maxlen int, key *crypto.CipherMsgKey) ([]byte, os.Error) {
	zipbuf, err := crypto.DecipherMsg(p, key)
	if err != nil {
		return nil, err
	}
	r, err := unzip(zipbuf, maxlen)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func zip(p []byte) []byte {
	var w bytes.Buffer
	gz,err := zlib.NewWriter(&w)
	if err != nil {
		panic("mon,gz")
	}
	n,err := gz.Write(p)
	if err != nil || n != len(p) {
		panic("mon,gz")
	}
	err = gz.Close()
	if err != nil {
		panic("mon,gz close")
	}
	return w.Bytes()
}

func unzip(p []byte, maxlen int) ([]byte, os.Error) {
	w := bytes.NewBuffer(p)
	gz,err := zlib.NewReader(w)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, maxlen)
	n, err := gz.Read(buf)
	if err != nil {
		return nil, err
	}
	neof, err := gz.Read(buf[n:])
	if neof != 0 || err != os.EOF {
		return nil, os.EINVAL
	}
	return buf[0:n], nil
}
