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
	//"json"
	"log"
	"os"
	"path"
	"rand"
	"strconv"
	//"sync"
	"time"
	"tonika/sys"
	"tonika/monitor"
	"tonika/dialer"
	"tonika/compass"
	"tonika/prof"
	"tonika/vault"
	"tonika/fe"
)

// TODO:
//   (*) Dialer must get a copy of friend (not friend itself) so that potential
//   changes don't mess with its operation

type Core struct {
	since   int64
	db      *buttress
	monitor *monitor.Monitor
	dialer  *dialer.Dialer0
	compass *compass.Compass0
	vault   *vault.Vault0
	fe      *fe.FrontEnd
	lk      prof.Mutex
}

type Args struct {
	Addr     string
	DbFile   string
	HomeDir  string
	CacheDir string
	FEDir    string
	FEAddr   string
	FEAllow  string
}

func MakeCore(args *Args) (core *Core, err os.Error) {
	// Db
	db, err := ReadFriendDb(args.DbFile)
	if err != nil {
		log.Stderrf("Friends file is either missing or corrupt, making new one")
		db, err = MakeFriendDb(args.DbFile)
		if err != nil {
			log.Stderrf("Couldn't create the friends file, sorry mate")
			return nil, err
		}
	}
	me := db.GetMe()
	if args.Addr != "" {
		me.Addr = args.Addr
	}
	if me.Addr == "" {
		port := 20000 + rand.Intn(20000)
		log.Stderrf("Tonika: Bounding Tonika to localhost port %d", port)
		me.Addr = ":" + strconv.Itoa(port)
		if err = db.Save(); err != nil {
			log.Stderrf("We failed to save your friend file")
			return nil, err
		}
	}

	// Dialer
	dialer, err := dialer.MakeDialer0(me, me.Addr, 100) // limit file descriptors to 100
	if err != nil {
		log.Stderrf("Problem starting the Dialer System: %s\n", err)
		return nil, err
	}

	// Compass
	compass := compass.MakeCompass0(*me.GetId(), dialer)

	// Vault
	vault, err := vault.MakeVault0(*me.GetId(), args.HomeDir,
		args.CacheDir, 30, dialer, compass)
	if err != nil {
		log.Stderrf("Problem starting Vault System: %s\n", err)
		return nil, err
	}

	// OK
	c := &Core{
		since:   time.Nanoseconds(),
		db:      db,
		dialer:  dialer,
		compass: compass,
		vault:   vault,
	}

	// Monitor
	c.monitor = monitor.MakeMonitor(c, sys.MonitorServerURL, sys.MonitorFrequency)

	myid := c.GetMyId()
	c.lk.Lock() // Block usage of methods until we exit from init

	// Front End
	fe, err := fe.MakeFrontEnd(args.FEDir+"/tmpl", args.FEDir+"/static", 
		args.FEAddr, args.FEAllow, c, vault, myid)
	if err != nil {
		log.Stderrf("Problem starting Front End System: %s\n", err)
		return nil, err
	}
	c.fe = fe
	c.lk.Unlock()
	c.syncAll()
	c.Save()
	
	go c.logLoop(path.Join(args.CacheDir, "tonika.log"))
	go c.loop()
	return c, nil
}

func (c *Core) loop() {
	c.lk.Lock()
	d := c.dialer
	c.lk.Unlock()
	for {
		su := d.WaitForStatus()
		c.lk.Lock()
		r := c.db.GetById(su.Id)
		if r != nil {
			r.SetOnline(su.Online)
		}
		c.lk.Unlock()
	}
}

func (c *Core) syncAll() {
	all := c.Enumerate()
	for _,v := range all {
		c.Sync(v.GetSlot())
	}
}

func (c *Core) GetBuild() string {
	return sys.Build
}

func (c *Core) GetMyId() sys.Id {
	c.lk.Lock()
	defer c.lk.Unlock()
	return (*c.db.GetMe().GetId())
}

func (c *Core) GetMySignatureKey() *sys.SigPubKey {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.db.GetMe().GetSignatureKey().PubKey()
}

func (c *Core) GetMyName() string {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.db.GetMe().GetName()
}

func (c *Core) GetMyEmail() string {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.db.GetMe().GetEmail()
}

func (c *Core) GetMyAddr() string {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.db.GetMe().GetAddr()
}

func (c *Core) GetMyExtAddr() string {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.db.GetMe().GetExtAddr()
}

func (c *Core) SetMy(key string, v interface{}) {
	c.lk.Lock()
	defer c.lk.Unlock()
	switch key {
	case "Name":
		c.db.GetMe().Name = v.(string)
	case "Email":
		c.db.GetMe().Email = v.(string)
	case "ExtAddr":
		c.db.GetMe().ExtAddr = v.(string)
	default:
		panic("logic")
	}
}

func (c *Core) GetBySlot(slot int) (sys.View, os.Error) {
	c.lk.Lock()
	defer c.lk.Unlock()
	r := c.db.GetBySlot(slot)
	if r != nil {
		return r, nil
	}
	return nil, os.EINVAL
}

func (c *Core) GetById(id sys.Id) (sys.View, os.Error) {
	c.lk.Lock()
	defer c.lk.Unlock()
	r := c.db.GetById(id)
	if r != nil {
		return r, nil
	}
	return nil, os.EINVAL
}

func (c *Core) GetByAcceptKey(key *sys.DialKey) (sys.View, os.Error) {
	c.lk.Lock()
	defer c.lk.Unlock()
	r := c.db.GetByAcceptKey(key)
	if r != nil {
		return r, nil
	}
	return nil, os.EINVAL
}

func (c *Core) GetByDialKey(key *sys.DialKey) (sys.View, os.Error) {
	c.lk.Lock()
	defer c.lk.Unlock()
	r := c.db.GetByDialKey(key)
	if r != nil {
		return r, nil
	}
	return nil, os.EINVAL
}

func (c *Core) Revoke(slot int) {
	c.lk.Lock()
	defer c.lk.Unlock()
	f := c.db.GetBySlot(slot)
	if f == nil {
		return
	}
	c.db.Remove(slot)
	if f.GetId() != nil {
		c.dialer.Revoke(*f.GetId())
	}
}

func (c *Core) Enumerate() []sys.View {
	c.lk.Lock()
	defer c.lk.Unlock()
	return c.db.Enumerate()
}

func (c *Core) Reserve() sys.View {
	c.lk.Lock()
	defer c.lk.Unlock()
	slot := c.db.UnusedSlot()
	f := &friend{sys.Friend{Slot: slot},false}
	f.Init()
	if c.db.Attach(slot, f) != nil {
		panic("c")
	}
	return f
}

// TODO: Note that dialkey, etc. are passed on as *DialKey, *SigKey, etc.
func (c *Core) Write(slot int, key string, v interface{}) (sys.View, os.Error) {
	c.lk.Lock()
	defer c.lk.Unlock()
	f := c.db.GetBySlot(slot)
	if f == nil {
		return nil, os.EINVAL
	}
	switch key {
	case "Name":
		f.Name = v.(string)
	case "Email":
		f.Email = v.(string)
	case "Addr":
		f.Addr = v.(string)
	case "Id":
		panic("disabled")
		if f.Id != nil {
			return nil, os.EINVAL
		}
		id := v.(sys.Id)
		f.Id = &id
	case "SignatureKey":
		// For now we'll allow overriding an old SignatureKey
		/*
		if f.SignatureKey != nil {
			return nil, os.EINVAL
		}
		*/
		f.SignatureKey = v.(*sys.SigPubKey)
		id := f.SignatureKey.Id()
		f.Id = &id
	case "DialKey":
		// For now we'll allow overriding an old SignatureKey
		/*
		if f.DialKey != nil {
			return nil, os.EINVAL
		}
		*/
		f.DialKey = v.(*sys.DialKey)
	default:
		return nil, os.EINVAL
	}
	return f, nil
}

func (c *Core) Save() {
	c.lk.Lock()
	defer c.lk.Unlock()
	c.db.Save()
}

func (c *Core) SyncAddr(slot int) {
	c.lk.Lock()
	defer c.lk.Unlock()
	f := c.db.GetBySlot(slot)
	if f == nil || !f.IsComplete() {
		return
	}
	g := *f // copy the friend structure
	id := *g.GetId()
	c.dialer.Update(id, g.Addr)
}

func (c *Core) Sync(slot int) {
	c.lk.Lock()
	defer c.lk.Unlock()
	f := c.db.GetBySlot(slot)
	if f == nil || !f.IsComplete() {
		return
	}
	g := *f // copy the friend structure
	id := *g.GetId()
	c.dialer.Revoke(id)
	c.dialer.Add(&g, g.Addr)
}
