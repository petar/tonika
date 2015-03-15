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


package filewriter

import (
	"bytes"
	"fmt"
	"gob"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"tonika/util/varint"
)

// FileWriter
type FileWriter struct {
	prefix  string
	file    *os.File
	written int64
	lk,glk  sync.Mutex
	enc     *gob.Encoder
}

func MakeFileWriter(fileprefix string) (*FileWriter, os.Error) {
	w := &FileWriter{ prefix: fileprefix }
	err := w.recycle()
	if err != nil {
		return nil, err
	}
	w.enc = gob.NewEncoder(w)
	return w, nil
}

func isFileAvailable(path string) bool {
	fi, _ := os.Lstat(path)
	return fi == nil
}

func generateName(prefix string) (string, os.Error) {
	dir, pre := path.Split(prefix)
	if dir == "" {
		dir = "."
	}
	d,err := os.Open(dir, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	files,err := d.Readdirnames(-1)
	if err != nil {
		return "", err
	}
	d.Close()
	files = filterAndSort(files, pre)
	if len(files) == 0 {
		return path.Join(dir, pre+"-00001"), nil
	}
	last := files[len(files)-1]
	k,err := strconv.Atoi(last[len(last)-5:])
	if err != nil {
		panic("filereader, atoi")
	}
	var w bytes.Buffer
	fmt.Fprintf(&w, "%s-%05d", pre, k+1)
	return path.Join(dir, w.String()), nil
}

func (w *FileWriter) recycle() os.Error {
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}
	w.written = 0
	name,err := generateName(w.prefix)
	if err != nil {
		return err
	}
	file, err := os.Open(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	w.file = file
	return nil
}

func (w *FileWriter) Write(p []byte) (int, os.Error) {
	w.lk.Lock()
	defer w.lk.Unlock()

	n,err := w.file.Write(p)
	if err != nil {
		return 0, err
	}
	if n != len(p) {
		panic("filewriter, write")
	}
	w.written += int64(n)
	if w.written > 1024*1024*100 {  // If > 100MB, start a new file
		return n, w.recycle()
	}
	return n, nil
}

func (w *FileWriter) Encode(e interface{}) os.Error {
	w.glk.Lock()
	defer w.glk.Unlock()
	return w.enc.Encode(e)
}

func (w *FileWriter) Close() os.Error {
	w.lk.Lock()
	defer w.lk.Unlock()
	err := w.file.Close()
	w.file = nil
	w.enc = nil
	return err
}

// LatestFileEncoder
type LatestFileEncoder struct {
	writer *FileWriter
	latest *os.File
	lk     sync.Mutex
}

func MakeLatestFileEncoder(fileprefix string) (*LatestFileEncoder, os.Error) {
	fw,err := MakeFileWriter(fileprefix)
	if err != nil {
		return nil, err
	}
	l, err := os.Open(fileprefix+"-latest", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	return &LatestFileEncoder{writer:fw, latest:l}, nil
}

func (w *LatestFileEncoder) Encode(e interface{}) os.Error {
	w.lk.Lock()
	defer w.lk.Unlock()

	// Write to main file
	err := w.writer.Encode(e)
	if err != nil {
		return err
	}

	// Update latest
	err = w.latest.Truncate(0)
	if err != nil {
		return err
	}
	_,err = w.latest.Seek(0,0)
	if err != nil {
		return err
	}

	return gob.NewEncoder(w.latest).Encode(e)
}

func (w *LatestFileEncoder) Close() os.Error {
	w.lk.Lock()
	defer w.lk.Unlock()

	err := w.writer.Close()
	w.latest.Close()
	w.writer, w.latest = nil, nil

	return err
}

// selfDelimitedWriter
type selfDelimitedWriter struct {
	io.Writer
	lk sync.Mutex
}

func newSelfDelimitedWriter(w io.Writer) *selfDelimitedWriter {
	return &selfDelimitedWriter{Writer:w}
}

func (sdw *selfDelimitedWriter) Write(p []byte) (int, os.Error) {
	sdw.lk.Lock()
	defer sdw.lk.Unlock()

	h := varint.EncodeVarint(uint64(len(p)))
	_,err := sdw.Writer.Write(h)
	if err != nil {
		return 0, err
	}
	return sdw.Writer.Write(p)
}
