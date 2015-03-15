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


package routing

import (
	"fmt"
	"json"
	"os"
	. "tonika/sys"
)

// A Algorithm is an abstract algorithm that can exchange 
// messages with neighbors (friends) and answer next-hop queries 
// for given source-destination pairs.
type Algorithm interface {

	// CONFIG

	NeedBand() bool

	// NOTIFICATIONS

	OnAddNeighbor(nid Id)

	OnRemoveNeighbor(nid Id)

	NewEmptyMsg() interface{}
	OnReceive(from Id, msg interface{}) os.Error

	OnClock()

	// COMMANDS

	Sweep()

	PollSend(to Id) (msg interface{})

	// QUERIES

	fmt.Stringer
	json.Marshaler

	// Id returns the ID of the local node
	Id() Id

	// Returns an array of all neighbors' IDs
	Neighbors() []Id

	// SourceCount returns the number of routable destinations
	SourceCount() int

	FlowFork(s,t Id) (fork Fork, err os.Error)

	FlowStep(s,t Id) (hop Id, err os.Error)
}
