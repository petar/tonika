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
	"tonika/http"
	"tonika/sys"
)

func (fe *FrontEnd) serveViaVault(q *http.Query) {
	req := q.GetRequest()
	resp, err := fe.vault.Serve(req)
	if err != nil {
		serveUserMsg("Could not fetch data from "+sys.Name+".", q)
		return
	}

	q.Continue()
	if req.Body != nil {
		req.Body.Close()
	}
	q.Write(resp)
}

func serveUserMsg(msg string, q *http.Query) {
	q.Continue()
	req := q.GetRequest()
	if req.Body != nil {
		req.Body.Close()
	}
	q.Write(buildResp(msg))
}
