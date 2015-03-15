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


package vault

import (
	"fmt"
	"testing"
	"tonika/http"
)

var (
	testReq = &http.Request{
		Host: "tonika",
		URL: &http.URL{
			Path:      "/42ak87623/aha",
			RawQuery:  "?d=3&g=5",
		},
	}
)

func TestSanitize(t *testing.T) {
	tid, path, query, err := sanitizeRequest(testReq) 
	if err != nil {
		t.Errorf("err %s", err)
	}
	fmt.Printf("%v — %v — %v\n", tid, path, query)
}
