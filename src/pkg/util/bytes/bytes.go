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


package bytes

import (
	"os"
	"unsafe"
)

func BytesToInt64(b []byte) (int64, os.Error) {
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

func Int64ToBytes(s int64) []byte {
	u := uint64(s)
	l := unsafe.Sizeof(u)
	b := make([]byte, l)
	for i := 0; i < l; i++ {
		b[i] = byte((u >> uint(8*(l-i-1))) & 0xff)
	}
	return b
}
