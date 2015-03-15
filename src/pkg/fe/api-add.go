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


// Copyright 2009 The Tonika Authors. All rights reserved.

package fe

import (
	"json"
	"tonika/http"
	"tonika/sys"
)

type apiAddResult struct {
	Msg string "Msg"
}

func (fe *FrontEnd) makeInvite(v sys.View) string {
	return "Hello " + v.GetName() + ",\n\n" +
		"I want to invite you in my "+sys.Name+" circle of friends.\n\n" +
		"Here's what you need to do:\n\n" +
		"  1. Download and install "+sys.Name+" from:\n\n" +
		"    "+sys.WWWURL+"\n\n" +
		"  2. Run "+sys.Name+".\n\n" +
		"  3. Click on the link below to accept my invitation:\n\n" +
		"" + fe.makeInviteLink(v) + "\n\n" +
		"Cheers,\n--" + fe.bank.GetMyName() + "\n\n" +
		"We hope you enjoy "+sys.Name+".\n--Team "+sys.Name+"\n"
}

func (fe *FrontEnd) makeInviteLink(v sys.View) string {
	// The variable names in the link are from the point of view of the receiver
	akopt := ""
	if v.GetDialKey() != nil {
		akopt = "&ak="+http.URLEscape(v.GetDialKey().String())
	}
	link := fe.adminURL+"/accept?" + 
		"na="+http.URLEscape(fe.bank.GetMyName())+
		"&em="+http.URLEscape(fe.bank.GetMyEmail())+
		"&sk="+http.URLEscape(fe.bank.GetMySignatureKey().String())+
		"&dk="+http.URLEscape(v.GetAcceptKey().String())+
		"&ad="+http.URLEscape(fe.bank.GetMyExtAddr())+
		akopt
	return link
}

func (fe *FrontEnd) replyAPIAdd(args map[string][]string) *http.Response {
	name, ok := args["n"]
	if !ok {
		return newRespBadRequest()
	}
	email, ok := args["e"]
	if !ok {
		return newRespBadRequest()
	}
	if name == nil || email == nil {
		return newRespBadRequest()
	}
	var n,e string
	if len(name) > 0 {
		n = name[0]
	}
	if len(email) > 0 {
		e = email[0]
	}

	v := fe.bank.Reserve()
	fe.bank.Write(v.GetSlot(), "Name", n)
	fe.bank.Write(v.GetSlot(), "Email", e)
	fe.bank.Save()

	j := &apiAddResult{ fe.makeInvite(v) }
	jb,err := json.Marshal(&j)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(string(jb))
}
