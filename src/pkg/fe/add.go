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
	"tonika/http"
	"tonika/sys"
)

type addData struct {
	AdminURL  string
}

func (fe *FrontEnd) replyAdminAdd(req *http.Request) *http.Response {
	// prepare content of page
	data := addData{ 
		AdminURL: fe.adminURL, 
	}
	var w bytes.Buffer
	err := fe.tmplAdd.Execute(data, &w)
	if err != nil {
		return newRespServiceUnavailable()
	}
	// wrap into a page frame
	pdata := pageData {
		Title: sys.Name + " &mdash; Add contact",
		CSSLinks: []string{"add.css"},
		JSLinks: []string{"jquery.scrollTo-1.4.2-min.js", "add.js"},
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
