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


// Defines ANSI terminal control sequences
package term

import (
)

const (
	Reset = "\x1b[0m"
	Bright = "\x1b[1m"
	Dim = "\x1b[2m"
	Underscore = "\x1b[4m"
	Blink = "\x1b[5m"
	Reverse = "\x1b[7m"
	Hidden = "\x1b[8m"

	FgBlack = "\x1b[30m"
	FgRed = "\x1b[31m"
	FgGreen = "\x1b[32m"
	FgYellow = "\x1b[33m"
	FgBlue = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan = "\x1b[36m"
	FgWhite = "\x1b[37m"

	BgBlack = "\x1b[40m"
	BgRed = "\x1b[41m"
	BgGreen = "\x1b[42m"
	BgYellow = "\x1b[43m"
	BgBlue = "\x1b[44m"
	BgMagenta = "\x1b[45m"
	BgCyan = "\x1b[46m"
	BgWhite = "\x1b[47m"
)
