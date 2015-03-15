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
	"bytes"
	"fmt"
	"os"
	"time"
	"tonika/prof"
	"tonika/util/misc"
)

func (c *Core) dumpProf() string { return prof.String() }

func (c *Core) marshalProfJSON() string { return misc.JSONQuote(prof.String()) }

func (c *Core) String() string {
	var w bytes.Buffer
	fmt.Fprintf(&w, "Build: %s, Id: %s, Lifetime: %d ns\n\n",
		c.GetBuild(), c.GetMyId().String(), time.Nanoseconds()-c.since)
	fmt.Fprintf(&w, "Dialer:\n%s\nCompass:\n%s\nVault:\n%s\nFrontEnd:\n%s\n",
		c.dialer.String(), c.compass.String(), c.vault.String(), c.fe.String())
	return w.String()
}

func (c *Core) MarshalJSON() ([]byte, os.Error) {
	var w bytes.Buffer
	fmt.Fprintf(&w, "{\"Build\":%s,\"Id\":%s,\"Lifetime\":%d,",
		misc.JSONQuote(c.GetBuild()), c.GetMyId().ToJSON(), time.Nanoseconds()-c.since)
	nj,err := c.FriendsMarshalJSON()
	if err != nil {
		return nil, err
	}
	dj,err := c.dialer.MarshalJSON()
	if err != nil {
		return nil, err
	}
	cj,err := c.compass.MarshalJSON()
	if err != nil {
		return nil, err
	}
	vj,err := c.vault.MarshalJSON()
	if err != nil {
		return nil, err
	}
	fj,err := c.fe.MarshalJSON()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&w, "\"Friends\":%s,\"Dialer\":%s,\"Compass\":%s,\"Vault\":%s,\"FrontEnd\":%s}", 
		nj, dj, cj, vj, fj)
	return w.Bytes(), nil
}

func (c *Core) FriendsMarshalJSON() ([]byte, os.Error) {
	var w bytes.Buffer
	views := c.Enumerate()
	w.WriteString("[")
	comma := false
	for _,v := range views {
		if v.GetId() != nil {
			if comma {
				w.WriteString(",")
			}
			comma = true
			w.WriteString("\""+v.GetId().String()+"\"")
		}
	}
	w.WriteString("]")
	return w.Bytes(), nil
}
