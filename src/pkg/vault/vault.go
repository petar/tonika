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
	//"fmt"
	"net"
	"os"
	"path"
	//"sync"
	"tonika/dialer"
	"tonika/compass"
	"tonika/http"
	"tonika/prof"
	"tonika/sys"
)

type Vault interface {
	Serve(req *http.Request) (*http.Response, os.Error)
}

type Vault0 struct {
	id        sys.Id // our id
	w         watch 
	hdir,cdir string // home dir, cache dir
	d         dialer.Dialer
	c         compass.Compass
	lk        prof.Mutex
	fdlim     http.FDLimiter
}

const maxHops = 10

func MakeVault0(id sys.Id, hdir,cdir string, fdlim int, 
	d dialer.Dialer, c compass.Compass) (*Vault0, os.Error) {

	if !isDirectory(hdir) {
		return nil, os.ErrorString("Bad home directory")
	}
	if !isDirectory(cdir) {
		return nil, os.ErrorString("Bad cache directory")
	}
	v := &Vault0{
		id:      id,
		hdir:    hdir,
		cdir:    cdir,
		d:       d,
		c:       c,
	}
	v.fdlim.Init(fdlim)
	v.w.Init(&v.fdlim)
	go v.accept()
	return v,nil
}

func (v *Vault0) String() string { return v.w.String() }

func (v *Vault0) MarshalJSON() ([]byte, os.Error) { return v.w.MarshalJSON() }

func (v *Vault0) accept() {
	for {
		d := v.isHealthy()
		if d == nil {
			return
		}
		_, conn := d.Accept("vault0")
		if conn == nil {
			continue
		}
		go v.serveOnBehalf(conn)
	}
}

// Serve connections coming from Dialer (i.e. from neighbors on the network)
func (v *Vault0) serveOnBehalf(c net.Conn) {
	asc := http.NewAsyncServerConn(c)
	req,err := asc.Read()
	if err != nil {
		asc.Close()
		c.Close()
		return
	}

	// Increase the hop count, due to successful arrival
	hop,err := parseReqHop(req)
	if err == nil {
		setReqHop(req, hop+1)
	} else {
		setReqHop(req, 0)
	}

	resp,err := v.serve(req, false)
	if resp != nil {
		setRespVersion(resp)
	}

	asc.Write(req, resp)
	asc.Close()
	c.Close()

	if err != nil {
		v.w.IncErrOnBehalfReqs()
	} else {
		v.w.IncOKOnBehalfReqs()
	}
}

func (v *Vault0) Serve(req *http.Request) (*http.Response, os.Error) {
	setReqHop(req, 0)
	setOrigin(req, v.id)

	resp,err := v.serve(req, true)
	if err != nil {
		v.w.IncErrMyReqs()
	} else {
		h,err := parseRespHop(resp)
		if err == nil {
			v.w.AddHopsBwd(h)  // Stat hop count back at origin
		}
		v.w.IncOKMyReqs()
	}
	// TODO: If response is not 200 OK, respond with an error. This way
	// we can have a consistent behavior in case we are talking to a more
	// advanced vault.
	sanitizeResp(resp)
	return resp,err
}

// The hop count (Vault-Hop) in a request or response is, by convention, equal
// to the number of hops the request/response had already travelled by the time
// it was received at its origin.

func (v *Vault0) serve(req *http.Request, my bool) (*http.Response, os.Error) {
	// We don't expect a body in requests (at this stage, there are no POSTs)
	if req.Body != nil {
		req.Body.Close()
		req.Body = nil
	}

	// Parse the URL
	tid,fpath,query,err := parseURL(req)
	if err != nil {
		return newRespBadRequest(), os.ErrorString("bad request")
	}

	// If request is at its final destination
	if tid == v.id {
		h,err := parseReqHop(req)
		if err == nil {
			v.w.AddHopsFwd(h)  // Stat hop count at destination
		}
		resp,err := v.serveLocal(fpath,query)
		if err == nil {
			setRespHop(resp,0)
		}
		return resp,err
	}

	// Stop if too many hops or hops not included
	h,err := parseReqHop(req)
	if err != nil || h > maxHops {
		v.w.IncTooManyHops()
		return newRespServiceUnavailable(), os.ErrorString("too many hops")
	}

	// Parse the HTTP header
	sid,err := parseOrigin(req)
	if err != nil {
		sid = v.id		
	}
	hid := v.c.QueryQuantize(sid,tid)
	if hid == nil {
		return newRespServiceUnavailable(), os.ErrorString("no route to destination")
	}
	d := v.isHealthy()
	if d == nil {
		return newRespServiceUnavailable(), os.ErrorString("service unavailable")
	}
	cc := d.Dial(*hid, "vault0")
	if cc == nil {
		return newRespServiceUnavailable(), os.ErrorString("service unavailable")
	}
	pcc := prof.NewConn(cc)
	acc := http.NewAsyncClientConn(pcc)

	// Pre-fetch
	_,err = parseReqHop(req)
	if err != nil {
		setReqHop(req, 0)
	}
	setReqVersion(req)

	// Fetch and post-fetch
	resp,err := acc.Fetch(req)
	if err != nil {
		acc.Close()
		pcc.Close()
		return newRespServiceUnavailable(), os.ErrorString("service unavailable")
	}

	// Update hop
	setRespVersion(resp)
	h,err = parseRespHop(resp)
	if err == nil {
		setRespHop(resp, h+1)
	} else {
		setRespHop(resp, 0)
	}

	// Do we understand the response?
	if !statusCodeSupported(resp.StatusCode) {
		acc.Close()
		pcc.Close()
		return newRespUnsupported(), nil
	}

	// Set hooks for when body is fully read
	if resp.Body == nil {
		// Update traffic stats
		if my {
			v.w.IncFwdMyInTraffic(*hid, pcc.InTraffic())
			v.w.IncFwdMyOutTraffic(*hid, pcc.OutTraffic())
		} else {
			v.w.IncFwdBehalfInTraffic(*hid, pcc.InTraffic())
			v.w.IncFwdBehalfOutTraffic(*hid, pcc.OutTraffic())
		}
		v.w.RecFwdLatencyPerByte(*hid, float64(pcc.Duration())/float64(pcc.InTraffic()))

		acc.Close()
		pcc.Close()
	} else {
		resp.Body = http.NewRunOnClose(resp.Body, func() {
			// Update traffic stats
			if my {
				v.w.IncFwdMyInTraffic(*hid, pcc.InTraffic())
				v.w.IncFwdMyOutTraffic(*hid, pcc.OutTraffic())
			} else {
				v.w.IncFwdBehalfInTraffic(*hid, pcc.InTraffic())
				v.w.IncFwdBehalfOutTraffic(*hid, pcc.OutTraffic())
			}
			v.w.RecFwdLatencyPerByte(*hid, float64(pcc.Duration())/float64(pcc.InTraffic()))

			acc.Close()
			pcc.Close()
		})
	}

	return resp, nil
}

func (v *Vault0) serveLocal(fpath, query string) (*http.Response, os.Error) {
	fpath = path.Clean(fpath)
	if len(fpath) > 0 && fpath[0] == '/' {
		fpath = fpath[1:]
	}
	if fpath == "" {
		fpath = "index.html"
	}
	full := path.Join(v.hdir, fpath)
	if !isFile(full) {
		if fpath == "index.html" {
			return newRespNoIndexHTML(), nil
		} else {
			return newRespNotFound(), os.ErrorString("not found")
		}
	}

	// Serve file if we can allocate a file descriptor in time
	if v.fdlim.LockOrTimeout(10e9) == nil {
		body, bodylen, err := http.FileToBody(full)
		if err != nil {
			if body != nil {
				body.Close()
			}
			v.fdlim.Unlock()
			return newRespServiceUnavailable(), os.ErrorString("service unavailable")
		}
		body = http.NewRunOnClose(body, func() { v.fdlim.Unlock() })
		return buildRespFromBody(body, bodylen), nil
	} else {
		return newRespServiceUnavailable(), os.ErrorString("service unavailable")
	}
	panic("unreach")
}

func (v *Vault0) isHealthy() dialer.Dialer {
	v.lk.Lock()
	defer v.lk.Unlock()
	return v.d
}

func (v *Vault0) ShutDown() {
	v.lk.Lock()
	v.d = nil
	v.lk.Unlock()
}
