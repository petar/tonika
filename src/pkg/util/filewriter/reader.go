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
	"gob"
	"path"
	"os"
	"sort"
	"strings"
	"sync"
)

// FileReader
type FileReader struct {
	prefix  string
	file    *os.File
	dec     *gob.Decoder
	lk      sync.Mutex
}

// MakeFileReader opens a system of files written be FileWriter.
func MakeFileReader(fileprefix string) (*FileReader, os.Error) {
	dir,_ := path.Split(fileprefix)
	if dir == "" {
		dir = "."
	}
	_,err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	return &FileReader{ prefix: fileprefix }, nil
}

func (fr *FileReader) SeekToLatestFile() os.Error {
	fr.lk.Lock()
	defer fr.lk.Unlock()

	// Find the file
	dir, pre := path.Split(fr.prefix)
	if dir == "" {
		dir = "."
	}
	d,err := os.Open(dir, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	files,err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	d.Close()
	files = filterAndSort(files, pre)
	if len(files) == 0 {
		return os.EOF
	}

	// Open it
	file := path.Join(dir, files[len(files)-1])
	fr.file,err = os.Open(file, os.O_RDONLY, 0)
	if err != nil {
		fr.file, fr.dec = nil, nil
		return err
	}
	fr.dec = gob.NewDecoder(fr.file)

	return nil
}

func inFormat(name, prefix string) bool {
	if !strings.HasPrefix(name,prefix) {
		return false
	}
	if len(name) < 6 {
		return false
	}
	if name[len(name)-6] != '-' {
		return false
	}
	for i := 0; i < 5; i++ {
		switch name[len(name)-1-i] {
		case '0','1','2','3','4','5','6','7','8','9':
		default:
			return false
		}
	}
	return true
}

func filterAndSort(names []string, prefix string) []string {
	for i := 0; i < len(names); i++ {
		if !inFormat(names[i],prefix) {
			names[i] = names[len(names)-1]
			names = names[0:len(names)-1]
			i--
		}
	}
	sort.SortStrings(names)
	return names
}

func (fr *FileReader) Read(p []byte) (int, os.Error) {
	fr.lk.Lock()
	defer fr.lk.Unlock()
	if fr.file == nil {
		return 0, os.EINVAL
	}
	return fr.file.Read(p)
}

func (fr *FileReader) Decode(e interface{}) os.Error {
	fr.lk.Lock()
	defer fr.lk.Unlock()
	if fr.dec == nil {
		return os.EINVAL
	}
	return fr.dec.Decode(e)
}

// Read latest bytes or gob

func LatestDecode(fprefix string, e interface{}) os.Error {
	l,err := os.Open(fprefix+"-latest", os.O_RDONLY, 0)
	if err != nil {
		return recoverLatestDecode(fprefix, e)
	}
	dec := gob.NewDecoder(l)
	err = dec.Decode(e)
	l.Close()
	return err
}

func recoverLatestDecode(fprefix string, e interface{}) os.Error {
	r,err := MakeFileReader(fprefix)
	if err != nil {
		return err
	}
	err = r.SeekToLatestFile()
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(r)
	// TODO: Deep copy the latest successful read, so the last 
	// unsuccessful read does not garble the good data
	ok := false
	for {
		err = dec.Decode(e)
		if err != nil {
			break
		}
		ok = true
	}
	if ok {
		return nil
	}
	return os.EOF
}
