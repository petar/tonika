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


package core

import (
	//"fmt"
	"testing"
)

func TestFriend(t *testing.T) {
	fdb0 := MakeFriendDb("test.db")
	if fdb0 == nil {
		t.Fatalf("Error making\n")
	}
	fdb0.SetMyName("Petar")
	fdb0.SetMyAddr("serdika.isp.nah")
	fdb0.SetFriend(proto.Friend{
		Id:   87598375,
		Name: "Chris",
		Addr: "145cpw",
	},
		true)
	fdb0.SetFriend(proto.Friend{
		Id:   122938129,
		Name: "Jennie",
		Addr: "nyny",
	},
		true)
	err := fdb0.Save()
	if err != nil {
		t.Fatalf("Error saving: %v\n", err)
	}

	fdb := ReadFriendDb("test.db")
	if fdb == nil {
		t.Fatalf("Open friend db error")
	}
	if len(fdb.GetAllFriends()) != 2 {
		t.Errorf("Bad # of parsed friends")
	}
	fs := fdb.GetAllFriends()
	jennie, present := fs[122938129]
	if !present || jennie.Name != "Jennie" {
		t.Errorf("Incorrect friend entries\n")
	}
}
