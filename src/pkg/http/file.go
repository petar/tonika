// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io"
	"os"
)

type fileBody struct {
	file *os.File
}

// FileToBody returns an io.ReadCloser which reads the contents
// of a file in a memory efficient manner, i.e. without reading
// the whole file initially. This function is intended for serving 
// large files.
func FileToBody(name string) (body io.ReadCloser, bodylen int64, err os.Error) {
	file,err := os.Open(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, 0, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	return &fileBody{file}, fi.Size, nil
}

func (fb *fileBody) Read(p []byte) (n int, err os.Error) {
	n,err = fb.file.Read(p)
	if err != nil {
		fb.file.Close()
		fb.file = nil
	}
	return n, err
}

func (fb *fileBody) Close() os.Error {
	if fb.file == nil {
		return nil
	}
	fb.file.Close()
	return nil
}
