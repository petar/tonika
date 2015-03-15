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
	"io"
	"io/ioutil"
	"log"
	"json"
	"os"
	"strconv"
	"rand"
	"tonika/sys"
)

type buttress struct {
	me   *sys.Me         // ourselves
	recs map[int]*friend // friends table
	path string          // filename with db on disk
}

// (===) Reading/writing and json representation

type jsonMe struct {
	Id      string
	SigKey  string // Base64 encoding
	Name    string
	Addr    string
	Email   string
	ExtAddr string
}

type jsonFriend struct {
	Slot      string
	Id        string // Eye64 encoding of Id
	Name      string
	Email     string
	Addr      string
	SigKey    string
	DialKey   string
	AcceptKey string
	HelloKey  string
	Rest      map[string]string
}

type jsonDb struct {
	Me      jsonMe
	Friends []jsonFriend
}

// Creates a blank friend db with no friends. Populates the Me structure with a
// generic name and a newly generated Id and corresponding private key.
func MakeFriendDb(path string) (*buttress, os.Error) {
	me := &sys.Me{}
	log.Stderrf("Generating identity information for you ...")
	me.Init()
	return &buttress{
		me:   me,
		recs: make(map[int]*friend),
		path: path,
	},
		nil
}

func (db *buttress) UnusedSlot() int {
	for {
		s := rand.Int()
		if _, ok := db.recs[s]; !ok {
			return s
		}
	}
	panic("unreach")
}

func ReadFriendDb(path string) (*buttress, os.Error) {
	// Read contents
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Stderrf("Cannot read from friends file \"%s\"\n", path)
		return nil, &Error{ErrLoad, err}
	}
	// Unmarshal json
	book := jsonDb{}
	if err := json.Unmarshal(bytes, &book); err != nil {
		log.Stderrf("Error decoding friends file [%s] json. "+
			"Offending token [%s].\n",
			path, err)
		return nil, &Error{ErrDecode, err}
	}
	db := &buttress{me: &sys.Me{}, recs: make(map[int]*friend), path: path}

	// Deep parse me-data
	db.me = &sys.Me{
		Name:    book.Me.Name,
		Addr:    book.Me.Addr,
		ExtAddr: book.Me.ExtAddr,
		Email:   book.Me.Email,
	}
	id, err := sys.ParseId(book.Me.Id)
	if err != nil {
		log.Stderrf("Error [%v] decoding my ID [%s]\n", err, book.Me.Id)
		return nil, &Error{ErrDecode, book.Me.Id}
	}
	db.me.Id = id
	pk, err := sys.ParseSigKey(book.Me.SigKey)
	if err != nil {
		log.Stderrf("Error [%v] decoding my signature key [%s]\n",
			err, book.Me.SigKey)
		return nil, &Error{ErrDecode, book.Me.SigKey}
	}
	db.me.SignatureKey = pk

	// Deep parse friends
	if book.Friends != nil {
		for i := 0; i < len(book.Friends); i++ {
			// slot
			slot, err := strconv.Atoi(book.Friends[i].Slot)
			if err != nil || slot < 0 {
				log.Stderrf("db, invalid slot number, skipping friends")
				continue
			}
			// id
			id1, err := sys.ParseId(book.Friends[i].Id)
			var id *sys.Id
			if err == nil {
				id = &id1
			}

			// dial key
			dkey, _ := sys.ParseDialKey(book.Friends[i].DialKey)

			// accept key
			akey, _ := sys.ParseDialKey(book.Friends[i].AcceptKey)

			// sig key
			sigk, err := sys.ParseSigPubKey(book.Friends[i].SigKey)

			// hello key
			hellok, err := sys.ParseHelloKey(book.Friends[i].HelloKey)

			// make
			fr := &friend{
				Friend: sys.Friend{
					Slot:         slot,
					Id:           id,
					SignatureKey: sigk,
					HelloKey:     hellok,
					DialKey:      dkey,
					AcceptKey:    akey,
					Name:         book.Friends[i].Name,
					Email:        book.Friends[i].Email,
					Addr:         book.Friends[i].Addr,
					Rest:         book.Friends[i].Rest,
				},
				online: false,
			}
			_, present := db.recs[fr.Slot]
			if present {
				log.Stderrf("Duplicate friend, using latest\n")
			}
			db.recs[fr.Slot] = fr
		} // for
	}
	return db, nil
}


// Saves the database to the friend db file that was used to read it
// TODO: This should be checkpointed to prevent data loss
func (db *buttress) Save() os.Error {
	// Convert to json
	me := db.me
	jm := jsonMe{
		Id:      me.Id.String(),
		SigKey:  me.SignatureKey.String(),
		Name:    me.Name,
		Email:   me.Email,
		Addr:    me.Addr,
		ExtAddr: me.ExtAddr,
	}
	book := &jsonDb{jm, make([]jsonFriend, len(db.recs))}
	k := 0
	for _, v := range db.recs {
		jf := jsonFriend{
			Slot:  strconv.Itoa(v.Slot),
			Name:  v.Name,
			Email: v.Email,
			Addr:  v.Addr,
			Rest:  v.Rest,
		}
		if v.Id != nil {
			jf.Id = v.Id.String()
		}
		if v.SignatureKey != nil {
			jf.SigKey = v.SignatureKey.String()
		}
		if v.HelloKey != nil {
			jf.HelloKey = v.HelloKey.String()
		}
		if v.DialKey != nil {
			jf.DialKey = v.DialKey.String()
		}
		if v.AcceptKey != nil {
			jf.AcceptKey = v.AcceptKey.String()
		}
		book.Friends[k] = jf
		k++
	}
	// Open file
	file, err := os.Open(db.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Stderrf("Cannot open friends file \"%s\"\n", db.path)
		return &Error{ErrSave, err}
	}
	// Marshal to json string
	data, err := json.Marshal(&book)
	if err != nil {
		return &Error{ErrEncode, err}
	}
	bf := bytes.NewBuffer(data)
	n,err := io.Copy(file, bf)
	if err != nil || n != int64(len(data)) {
		return &Error{ErrEncode, err}
	}
	if err := file.Close(); err != nil {
		return &Error{ErrSave, err}
	}
	return nil
}

func (db *buttress) GetMe() *sys.Me { return db.me }

func (db *buttress) Enumerate() []sys.View {
	r := make([]sys.View, len(db.recs))
	i := 0
	for _, f := range db.recs {
		r[i] = f
		i++
	}
	return r
}

func (db *buttress) Remove(slot int) { db.recs[slot] = nil, false }

func (db *buttress) GetById(id sys.Id) *friend {
	for _, f := range db.recs {
		if f.GetId() != nil && *f.GetId() == id {
			return f
		}
	}
	return nil
}

func (db *buttress) GetByAcceptKey(key *sys.DialKey) *friend {
	for _, f := range db.recs {
		if f.GetAcceptKey() != nil && f.GetAcceptKey().Equal(*key) {
			return f
		}
	}
	return nil
}

func (db *buttress) GetByDialKey(key *sys.DialKey) *friend {
	for _, f := range db.recs {
		if f.GetDialKey() != nil && f.GetDialKey().Equal(*key) {
			return f
		}
	}
	return nil
}

func (db *buttress) GetBySlot(slot int) *friend {
	f, ok := db.recs[slot]
	if !ok {
		return nil
	}
	return f
}

func (db *buttress) Attach(slot int, u *friend) os.Error {
	if _, present := db.recs[slot]; present {
		return os.ErrorString("db, slot busy")
	}
	if u.GetId() != nil {
		for _, f := range db.recs {
			if f.GetId() != nil && *f.GetId() == *u.GetId() {
				return os.ErrorString("db, Duplicate Id present")
			}
		}
	}
	u.Slot = slot
	db.recs[slot] = u
	return nil
}
