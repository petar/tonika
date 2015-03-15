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


package vault

import (
	"os"
)

func isFile(fpath string) bool {
	dir, err := os.Lstat(fpath)
	if err != nil {
		return false
	}
	if dir != nil && dir.IsDirectory() {
		return false
	}
	return true
}

func isDirectory(dpath string) bool {
	dir, err := os.Lstat(dpath)
	if err != nil {
		return false
	}
	if dir != nil && dir.IsDirectory() {
		return true
	}
	return false
}
