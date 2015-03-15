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


// *** SRC (Standards-related code) ***

// This package converts 64-bit ints to and from strings, using our in-house
// encoding Eyeball-64. The encoding is case-incensitive and uses only
// easily-visually recognized characters, i.e. we don't use lower-case
// 'L' or '1', since both are hard to eyeball and distinguish in fonts like
// Helvetica and Arial. The last character is a simple checksum to prevent 
// from typing and copying mistakes.
package eye64

import (
	"os"
	"strconv"
	"strings"
)

const (
	alpha = "abcdefghjkmnpqrstuwxyz23456789"
	base = len(alpha)
	zero = "aa"
)

// Return the first number n such that n*base >= 1<<64.
func cutoff64(base int) uint64 {
	if base < 2 {
		panic("eye64")
	}
	return (1<<64-1)/uint64(base) + 1
}

func EyeToU64(s string) (n uint64, err os.Error) {
	if len(s) < 2 {
		return 0, os.EINVAL
	}
	s0 := strings.ToLower(s)

	n = 0
	cutoff := cutoff64(base)
	crc := 0

	for i := 0; i < len(s0)-1; i++ {
		var v byte
		w := strings.Index(alpha, string(s0[i]))
		if w < 0 {
			return 0, &strconv.NumError{s, os.EINVAL}
		} else {
			crc = (crc + (w % base)) % base
			v = byte(w)
		}
		if n >= cutoff {
			// n*b overflows
			return 1<<64-1, &strconv.NumError{s, os.ERANGE}
		}
		n *= uint64(base)

		n1 := n + uint64(v)
		if n1 < n {
			// n+v overflows
			return 1<<64-1, &strconv.NumError{s, os.ERANGE}
		}
		n = n1
	} // for

	w := strings.Index(alpha, string(s0[len(s0)-1]))
	if w < 0 {
		return 0, &strconv.NumError{s, os.EINVAL}
	} 
	if w != crc {
		return 0, &strconv.NumError{s, os.EINVAL}
	}

	return n, nil
}

func U64ToEye(u uint64) string {
	if u == 0 {
		return zero
	}

	// Assemble decimal in reverse order.
	var crc int
	var buf [32]byte
	j := len(buf)-1
	b := uint64(base)
	for u > 0 {
		j--
		buf[j] = alpha[u % b]
		crc = (crc + int(u % b)) % base
		u /= b
	}
	buf[len(buf)-1] = alpha[crc]

	return string(buf[j:])
}
