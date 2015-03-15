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
	//"path"
	//"strconv"
	//"strings"
	"tonika/http"
	"tonika/sys"
)

type rootData struct {
	MyURL     string
	AdminURL  string
	Links     string
}

func (fe *FrontEnd) replyRoot() *http.Response {
	friends := fe.bank.Enumerate()
	var u bytes.Buffer
	u.WriteString("<ul>")
	i := 0
	for _,f := range friends {
		if f.IsOnline() {
			u.WriteString("<li><div class=\"ok\"><a href=\"" + 
				sys.MakeURL("", *f.GetId(), "") + "\">" + f.GetName() +
				"</a></div></li>")
		} else {
			u.WriteString("<li><div class=\"nop\">" + f.GetName() + "</div></li>")
		}
		i++
	}
	u.WriteString("</ul>")
	data := rootData{ 
		MyURL:    fe.myURL,
		AdminURL: fe.adminURL, 
		Links:    u.String(), 
	}
	var w bytes.Buffer
	err := fe.tmplRoot.Execute(&data, &w)
	if err != nil {
		return newRespServiceUnavailable()
	}

	pdata := pageData {
		Title: sys.Name+" &mdash; Start page",
		CSSLinks: []string{"root.css"},
		JSLinks: []string{"root.js"},
		GridLayout: "",
		Content: w.String(),
	}
	var w2 bytes.Buffer
	err = fe.tmplPage.Execute(&pdata, &w2)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(w2.String())
}
