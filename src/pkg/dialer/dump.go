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


package dialer

import (
	"bytes"
	"fmt"
	//"json"
	"os"
	"tonika/util/misc"
)

func (t *telephone) String() string {
	t.lk.Lock()
	defer t.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "    Id: %s\n", t.auth.GetId().Eye())
	fmt.Fprintf(&w, "      Connecting:\n")
	for c, _ := range t.authing {
		fmt.Fprintf(&w, "        %s\n", c.String())
	}
	fmt.Fprintf(&w, "      Established:\n")
	for c, _ := range t.conns {
		fmt.Fprintf(&w, "        %s\n", c.String())
	}
	return w.String()
}

func (t *telephone) MarshalJSON() ([]byte, os.Error) {
	t.lk.Lock()
	defer t.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "{\"Id\":%s,\"Connecting\":[", t.auth.GetId().ToJSON())
	comma := false
	for c, _ := range t.authing {
		cj,err := c.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w, "%s", cj)
		comma = true
	}
	fmt.Fprintf(&w, "],\"Established\":[")
	comma = false
	for c, _ := range t.conns {
		cj,err := c.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w, "%s", cj)
		comma = true
	}
	fmt.Fprintf(&w, "]}")
	return w.Bytes(), nil
}

func (d *Dialer0) String() string {
	d.lk.Lock()
	defer d.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "FD: %d/%d\n", d.fdlim.LockCount(), d.fdlim.Limit())
	fmt.Fprintf(&w, "  Services:\n")
	for subj, _ := range d.listens {
		fmt.Fprintf(&w, "    %s\n", subj)
	}
	fmt.Fprintf(&w, "  Unauthd conns:\n")
	for c, _ := range d.unauthd {
		fmt.Fprintf(&w, "    %s\n", c.String())
	}
	fmt.Fprintf(&w, "  Telephones:\n")
	for _, t := range d.tels {
		fmt.Fprintf(&w, "%s", t.String())
	}
	return w.String()
}

func (d *Dialer0) MarshalJSON() ([]byte, os.Error) {
	d.lk.Lock()
	defer d.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "{\"FD\":%d,\"FDLim\":%d,\"Services\":[", 
		d.fdlim.LockCount(), d.fdlim.Limit())
	comma := false
	for subj, _ := range d.listens {
		if comma {
			fmt.Fprintf(&w,",")
		}
		fmt.Fprintf(&w, "%s", misc.JSONQuote(subj))
		comma = true
	}
	fmt.Fprintf(&w, "],\"Unauthd\":[")
	comma = false
	for c, _ := range d.unauthd {
		cj,err := c.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w, "%s", cj)
		comma = true
	}
	fmt.Fprintf(&w, "],\"Telephones\":[")
	comma = false
	for _, t := range d.tels {
		tj,err := t.MarshalJSON()
		if err != nil {
			return nil, err
		}
		if comma {
			fmt.Fprintf(&w, ",")
		}
		fmt.Fprintf(&w, "%s", tj)
		comma = true
	}
	fmt.Fprintf(&w, "]}")
	return w.Bytes(), nil
}
