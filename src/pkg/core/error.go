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


package core

import (
	"fmt"
)

// A returned Error indicates task was not completed
type Error struct {
	no  int
	arg interface{}
}

func (e *Error) String() string { return fmt.Sprintf("TonErr: %d, %v", e.no, e.arg) }

const (
	ErrCreate = iota
	ErrLoad   = iota
	ErrSave   = iota
	ErrEncode = iota
	ErrDecode = iota
	ErrDup    = iota
)
