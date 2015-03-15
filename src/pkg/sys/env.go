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


package sys

// Build A.B.C.D convention:
//   A(STAGE): Denotes a major stage, like e.g. A=1 means Alpha Test Stage
//   B(PROTOCOL): Increments when a protocol is modified or a new protocol is included
//   C(FEATURE): Increments when a new feature is included
//   D(FIX): Increments on bug fixes and cosmetic changes

const (
	Name             = "Tonika"
	Host0            = "org"
	Host1            = "5ttt"
	Host             = Host1+"."+Host0
	WWWURL           = "http://"+Host
	Build            = "1.0.0.1"
	Released         = "6/6/2010"
	MonitorServerURL = "http://mon."+Host+":49494"
	MonitorFrequency = 10 * 60 * 1e9 // 10 mins, in nanoseconds
	MonitorPubKey    = "d/vZhRk6COhwfTCwzG63JesRFpAWb6+yjyVb/Rt8tTaUnPgPj/B8UVRiDB+afjrT/hf4qY5+ZUoYIXIHcLM=,AAAAAAAAAAM=K"
	TangraServerURL  = "http://tangra."+Host+":37373/green"
	//TangraServerURL  = "http://localhost:37373/green" // for testing
)

// hostPrefix --> hostPrefix.5ttt.org
func MakeHost(hostPrefix string, id Id) string {
	if hostPrefix != "" {
		return hostPrefix+"."+id.String()+"."+Host
	} else {
		return id.String()+"."+Host
	}
	panic("unreach")
}

// hostPrefix --> http://hostPrefix.5ttt.org/path
func MakeURL(hostPrefix string, id Id, path string) string {
	if len(path) > 0 && path[0] != '/' {
		return "http://"+MakeHost(hostPrefix, id)+"/"+path
	} else {
		return "http://"+MakeHost(hostPrefix, id)+path
	}
	panic("unreach")
}
