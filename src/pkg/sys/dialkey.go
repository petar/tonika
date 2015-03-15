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

package sys

import (
	"os"
	"tonika/crypto"
	"tonika/util/eye64"
)

// It is essential that all key-types below be immutable, since
// they are passed on from FE to CORE, or CORE to DIALER. They change
// hands a lot and the internal logic would break if any of these systems
// could change these objects.


// DialKey is the type of dial and accept keys.
type DialKey int64

func GenerateDialKey() *DialKey {
	rand := crypto.NewTimedRand()
	dk := DialKey(rand.Int63())
	return &dk
}

func ParseDialKey(s string) (dk *DialKey, err os.Error) {
	i, err := eye64.EyeToU64(s)
	if err != nil {
		return nil, err
	}
	k := DialKey(i)
	return &k, nil
}

func (dk DialKey) Equal(dk2 DialKey) bool { return int64(dk) == int64(dk2) }

func (dk DialKey) Int64() int64 { return int64(dk) }

func (dk DialKey) String() string {
	return eye64.U64ToEye(uint64(dk))
}

type U_DialKey struct {
	Int64 int64
}

func (dk DialKey) Proto() *U_DialKey {
	return &U_DialKey{dk.Int64()}
}

func UnprotoDialKey(p *U_DialKey) (*DialKey, os.Error) {
	dk := DialKey(p.Int64)
	return &dk, nil
}
