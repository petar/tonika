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
	"json"
	"strconv"
	"tonika/http"
	"tonika/sys"
)

type apiAcceptResult struct {
	InviteMsg string "InviteMsg"
	ErrMsg    string "ErrMsg"
}

func (fe *FrontEnd) replyAPIAccept(args map[string][]string) *http.Response {
	// name and email args
	na, ok := args["na"]
	if !ok || na == nil || len(na) != 1 {
		return newRespBadRequest()
	}

	em, ok := args["em"]
	if !ok || em == nil || len(em) != 1 {
		return newRespBadRequest()
	}

	ad, ok := args["ad"]
	if !ok || ad == nil || len(ad) != 1 {
		return newRespBadRequest()
	}

	sl, ok := args["sl"]
	if !ok || sl == nil || len(sl) != 1 {
		return newRespBadRequest()
	}

	dk, ok := args["dk"]
	if !ok || dk == nil || len(dk) != 1 {
		return newRespBadRequest()
	}
	dialkey, err := sys.ParseDialKey(dk[0])
	if err != nil {
		return newRespBadRequest()
	}

	sk, ok := args["sk"]
	if !ok || sk == nil || len(sk) != 1 {
		return newRespBadRequest()
	}
	sigkey, err := sys.ParseSigPubKey(sk[0])
	if err != nil {
		return newRespBadRequest()
	}

	// Logic
	result := &apiAcceptResult{}
	s, err := strconv.Atoi(sl[0])
	var v sys.View
	adding := false
	if err == nil {
		v, err = fe.bank.GetBySlot(s)
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
	if v == nil {
		v = fe.bank.Reserve()
		s = v.GetSlot()
		adding = true
	}

	if na[0] != "" {
		fe.bank.Write(s, "Name", na[0])
	}
	if em[0] != "" {
		fe.bank.Write(s, "Email", em[0])
	}
	if ad[0] != "" {
		fe.bank.Write(s, "Addr", ad[0])
	}
	_, err = fe.bank.Write(s, "SignatureKey", sigkey)
	if err != nil {
		result.ErrMsg = "This friend's invite has already been accepted."
	}
	_, err = fe.bank.Write(s, "DialKey", dialkey)
	if err != nil {
		result.ErrMsg = "This friend's invite has already been accepted."
	}
	if adding {
		result.InviteMsg = fe.makeInvite(v)
	}

	fe.bank.Sync(s)
	fe.bank.Save()

	jb, err := json.Marshal(&result)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(string(jb))
}
