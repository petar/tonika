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


package fe

import (
	"bytes"
	"fmt"
	"json"
	"tonika/http"
	"tonika/sys"
)

type apiLiveResult struct {
	Links string "Links"
}

func (fe *FrontEnd) replyAPILive(args map[string][]string) *http.Response {
	var w bytes.Buffer
	fmt.Fprintf(&w, "<ul>")
	friends := fe.bank.Enumerate()
	for _,f := range friends {
		if f.IsOnline() {
			fmt.Fprintf(&w, "<li><div class=\"ok\"><a href=\"" + 
				sys.MakeURL("", *f.GetId(), "") + 
				"\">" + f.GetName() + "</a></div></li>")
		} else {
			fmt.Fprintf(&w, "<li><div class=\"nop\">" + f.GetName() + "</div></li>")
		}
	}
	fmt.Fprintf(&w, "</ul>")
	r := &apiLiveResult{w.String()}
	jb, err := json.Marshal(r)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(string(jb))
}
