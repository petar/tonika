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
	//"tonika/sys"
)

type editData struct {
	Slot  int
	Name  string
	Email string
	Addr  string

	Id    string
	DialKey   string
	AcceptKey string
	SigKey    string
	HelloKey  string
}

func (fe *FrontEnd) replyAdminEdit(req *http.Request) *http.Response {

	args, err := http.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return newRespBadRequest()
	}
	s, ok := args["s"]
	if !ok || s == nil || len(s) != 1 {
		return newRespBadRequest()
	}
	sn, err := strconv.Atoi(s[0])
	if err != nil {
		return newRespBadRequest()
	}
	v, err := fe.bank.GetBySlot(sn)
	if err != nil {
		return newRespBadRequest()
	}

	data := editData{
		Slot:  sn,
		Name:  v.GetName(),
		Email: v.GetEmail(),
		Addr: v.GetAddr(),
	}
	if v.GetId() != nil {
		data.Id = v.GetId().Eye()
	}
	if v.GetDialKey() != nil {
		data.DialKey = v.GetDialKey().String()
	}
	if v.GetAcceptKey() != nil {
		data.AcceptKey = v.GetAcceptKey().String()
	}
	if v.GetHelloKey() != nil {
		data.HelloKey = v.GetHelloKey().String()
	}
	if v.GetSignatureKey() != nil {
		data.SigKey = v.GetSignatureKey().String()
	}

	// prepare content of page
	var w bytes.Buffer
	err = fe.tmplEdit.Execute(&data, &w)
	if err != nil {
		return newRespServiceUnavailable()
	}

	// wrap into a page frame
	pdata := pageData{
		Title:      "Tonika &mdash; Edit contact",
		CSSLinks:   []string{"edit.css"},
		JSLinks:    []string{"edit.js"},
		GridLayout: "",
		Content:    w.String(),
	}
	var w2 bytes.Buffer
	err = fe.tmplPage.Execute(&pdata, &w2)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(w2.String())
}
