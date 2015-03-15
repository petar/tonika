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

type acceptData struct {
	Name     string
	Email    string
	Addr     string
	Slot     string
	DialKey  string
	SigKey   string
	AdminURL string
}

func (fe *FrontEnd) replyAdminAccept(req *http.Request) *http.Response {
	args, err := http.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return buildResp("Invalid accept link")
	}
	// Read SignatureKey
	s, ok := args["sk"]
	if !ok || s == nil || len(s) != 1 {
		return buildResp("Invalid accept link")
	}
	sigkeyText := s[0]
	_, err = sys.ParseSigPubKey(sigkeyText)
	if err != nil {
		return buildResp("Invalid accept link")
	}
	// Read DialKey
	d, ok := args["dk"]
	if !ok || d == nil || len(d) != 1 {
		return buildResp("Invalid accept link")
	}
	dialkeyText := d[0]
	dialkey, err := sys.ParseDialKey(dialkeyText)
	if err != nil {
		return buildResp("Invalid accept link")
	}
	// Read AcceptKey (optional)
	a, ok := args["ak"]
	var acceptkey *sys.DialKey
	if ok && a != nil && len(a) == 1 {
		acceptkey, err = sys.ParseDialKey(a[0])
		if err != nil {
			return buildResp("Invalid accept link")
		}
	}
	// Read Name
	var name string
	n, ok := args["na"]
	if ok && n != nil && len(n) == 1 {
		name = n[0]
	}
	// Read Email
	var email string
	e, ok := args["em"]
	if ok && e != nil && len(e) == 1 {
		email = e[0]
	}
	// Read Addr
	var addr string
	ad, ok := args["ad"]
	if ok && ad != nil && len(ad) == 1 {
		addr = ad[0]
	}

	// Reconcile with local database
	var v sys.View
	if acceptkey != nil {
		v, err = fe.bank.GetByAcceptKey(acceptkey)
		if err != nil {
			v = nil
		}
	}
	if v == nil {
		v, err = fe.bank.GetByDialKey(dialkey)
		if err != nil {
			v = nil
		}
	}
	if v != nil {
		if v.GetName() != "" {
			name = v.GetName()
		}
		if v.GetEmail() != "" {
			email = v.GetEmail()
		}
		if v.GetAddr() != "" {
			addr = v.GetAddr()
		}
	}
	var slotText string
	if v != nil {
		slotText = strconv.Itoa(v.GetSlot())
	}

	// prepare content
	data := acceptData{
		Name:     name,
		Email:    email,
		Addr:     addr,
		Slot:     slotText,
		DialKey:  dialkeyText,
		SigKey:   sigkeyText,
		AdminURL: fe.adminURL,
	}
	var w bytes.Buffer
	err = fe.tmplAccept.Execute(&data, &w)
	if err != nil {
		return newRespServiceUnavailable()
	}

	// prepare page
	pdata := pageData{
		Title:      sys.Name+" &mdash; Accept invitation?",
		CSSLinks:   []string{"accept.css"},
		JSLinks:    []string{"accept.js"},
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
