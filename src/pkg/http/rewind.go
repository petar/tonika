// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io"
	"os"
)

// rewindBody implements io.ReadCloser, intended for use
// in conjunction with Response.Body or Request.Body.
// Reading retrieves the contents of an underlying []byte.
// Rewinding resets the reader position in the buffer to 0,
// so it can be used again.
type rewindBody struct {
	body []byte
	pos  int
}

func (rwb *rewindBody) Read(p []byte) (n int, err os.Error) {
	if rwb.pos < 0 {
		return -1, os.EOF
	}
	n = copy(p, rwb.body)
	if n > 0 {
		return
	}
	return 0, os.EOF
}

func (rwb *rewindBody) Close() os.Error {
	rwb.pos = -1
	return nil
}

func (rwb *rewindBody) Rewind() os.Error {
	rwb.pos = 0
	return nil
}
