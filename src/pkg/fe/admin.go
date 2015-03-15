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
	"strconv"
	"tonika/http"
	"tonika/sys"
)

type pageData struct {
	Title      string
	CSSLinks   []string
	JSLinks    []string
	GridLayout string
	Content    string
}

type friendData struct {
	Id          string
	Slot        string
	Name        string
	Email       string
	StatusMsg   string
	StatusClass string
	AdminURL    string
}

func friendToJSON(a sys.View, adminURL string) *friendData {
	r := &friendData{
		Slot:        strconv.Itoa(a.GetSlot()),
		Name:        a.GetName(),
		Email:       a.GetEmail(),
		StatusMsg:   a.GetStatusMsg(),
		StatusClass: a.GetStatusClass(),
		AdminURL:    adminURL,
	}
	if a.GetId() != nil {
		r.Id = a.GetId().Eye()
	}
	return r
}

func friendsToJSON(friends []sys.View, adminURL string) []*friendData {
	r := make([]*friendData, len(friends))
	for i := 0; i < len(friends); i++ {
		r[i] = friendToJSON(friends[i], adminURL)
	}
	return r
}

type adminData struct {
	MyId      string
	MyName    string
	MyEmail   string
	MyAddr    string
	MyExtAddr string
	AdminURL  string
	Friends   []*friendData
}

func (fe *FrontEnd) replyAdminMain() *http.Response {
	// prepare content of page
	adata := adminData{ 
		MyId: fe.bank.GetMyId().Eye(),
		MyName: fe.bank.GetMyName(),
		MyEmail: fe.bank.GetMyEmail(),
		MyAddr: fe.bank.GetMyAddr(),
		MyExtAddr: fe.bank.GetMyExtAddr(),
		AdminURL: fe.adminURL,
		Friends: friendsToJSON(fe.bank.Enumerate(), fe.adminURL), 
	}
	var w bytes.Buffer
	err := fe.tmplAdmin.Execute(&adata, &w)
	if err != nil {
		return newRespServiceUnavailable()
	}
	// wrap into a page frame
	pdata := pageData {
		Title: sys.Name+" &mdash; Admin",
		CSSLinks: []string{"admin.css"},
		JSLinks: []string{/*"gridlayout.js"*/ "admin.js"},
		GridLayout: "",
/*
`
<div id="GridLayout">
	<div id="GridLayout-params">
	{
		column_width:400,
		column_count:2,
		subcolumn_count:2,
		column_gutter:20,
		align:'center'
	}
	</div>
</div>
`,
*/
		Content: w.String(),
	}
	var w2 bytes.Buffer
	err = fe.tmplPage.Execute(&pdata, &w2)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(w2.String())
}
