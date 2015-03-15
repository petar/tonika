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
	"big"
	"crypto/rsa"
)

type U_RSAPubKey struct {
	N []byte
	E int
}

func RSAPubKeyToProto(pubkey *rsa.PublicKey) *U_RSAPubKey {
	return &U_RSAPubKey{
		N: pubkey.N.Bytes(),
		E: pubkey.E,
	}
}

func ProtoToRSAPubKey(proto *U_RSAPubKey) *rsa.PublicKey {
	return &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(proto.N),
		E: proto.E,
	}
}
