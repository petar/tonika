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


package sys


type U_AuthConn_M1 struct {
	DialKey   U_DialKey
	Challange []byte
}

type U_AuthAcc_M1 struct {
	Challange []byte
}

type U_AuthConn_M2 struct {
	Resp      []byte
}

type U_AuthAcc_M2 struct {
	DialKey U_DialKey
	Resp    []byte
}
