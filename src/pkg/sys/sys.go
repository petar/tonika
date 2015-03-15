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

import (
	"fmt"
	"crypto/sha1"
	"os"
	"tonika/util/bytes"
	"tonika/util/eye64"
	"tonika/util/misc"
)

// Interfaces

type AuthLocal interface {
	GetId() *Id
	GetSignatureKey() *SigKey
}

type AuthRemote interface {
	GetId() *Id
	GetSignatureKey() *SigPubKey
	GetDialKey() *DialKey
	GetAcceptKey() *DialKey
	GetHelloKey() *HelloKey
}

type Identity interface {
	GetName() string
	GetEmail() string
	GetAddr() string
}

type Health interface {
	GetSlot() int
	GetStatusMsg() string
	GetStatusClass() string
	IsOnline() bool
}

type Info interface {
	Identity
	AuthRemote
}

type View interface {
	Info
	Health
}

// Structs

// Me
type Me struct {
	Id Id
	SignatureKey *SigKey
	Name    string
	Email   string
	Addr    string
	ExtAddr string
}

func (m *Me) Init() {
	key := GenerateSigKey()
	m.Id = key.Id()
	m.SignatureKey = key
}

func (m *Me) GetId() *Id { return &m.Id }
func (m *Me) GetAddr() string { return m.Addr }
func (m *Me) GetExtAddr() string { return m.ExtAddr }
func (m *Me) GetName() string { return m.Name }
func (m *Me) GetEmail() string { return m.Email }
func (m *Me) GetSignatureKey() *SigKey { return m.SignatureKey }

// Friend
type Friend struct {
	Slot int                     // we generate
	Name string                  // we choose
	Email string                 // we choose
	Id *Id                       // they provide
	SignatureKey *SigPubKey      // they provide
	DialKey *DialKey             // they provide
	Addr string	             // they provide
	AcceptKey *DialKey           // we generate
	HelloKey *HelloKey           // we generate
	Rest map[string]string
}

func (f *Friend) GetSlot() int { return f.Slot }

func (f *Friend) GetName() string { return f.Name }
func (f *Friend) GetEmail() string { return f.Email }
func (f *Friend) GetAddr() string { return f.Addr }

func (f *Friend) GetId() *Id { return f.Id }
func (f *Friend) GetSignatureKey() *SigPubKey { return f.SignatureKey }
func (f *Friend) GetDialKey() *DialKey { return f.DialKey }
func (f *Friend) GetAcceptKey() *DialKey { return f.AcceptKey }
func (f *Friend) GetHelloKey() *HelloKey { return f.HelloKey }

func (f *Friend) Init() {
	f.AcceptKey = GenerateDialKey()
	f.HelloKey = GenerateHelloKey()
}

func (f *Friend) IsComplete() bool {
	return f.Id != nil && 
		f.SignatureKey != nil && 
		f.HelloKey != nil && 
		f.DialKey != nil && 
		f.AcceptKey != nil
}

func (f *Friend) PrettyBrief() string {
	return fmt.Sprintf("[%s/%s/%s]", f.Name, f.Id.String(), f.Addr)
}

// Id
type Id uint64

func ParseId(s string) (Id, os.Error) {
	u, err := eye64.EyeToU64(s)
	id := Id(u)
	return id, err
}

func (id Id) Eye() string {
	return eye64.U64ToEye(uint64(id))
}

func (id Id) Equal(x Id) bool { return uint64(id) == uint64(x) }

func (id Id) Mask() Id {
	sha := sha1.New()
	b := bytes.Int64ToBytes(int64(uint64(id)))
	sha.Write(b)
	sum := sha.Sum()
	i,err := bytes.BytesToInt64(sum[0:64/8])
	if err != nil {
		panic("id, mask")
	}
	return Id(uint64(i))
}

func (id Id) String() string { return id.Eye() }

func (id Id) ToJSON() string { return misc.JSONQuote(id.Eye()) }

func (id Id) MarshalJSON() ([]byte, os.Error) { return []byte(id.ToJSON()), nil }

// Presence
type Presence struct {
	Id Id
	MaybeOnline bool
	Reachable bool
	Rating float64 // uptime rating using util/uptime
}

const (
	InitialRating = 1.0
	Halflife = 60*60 // in sec = 1 hour 
	RatingBound = 1e20
)
