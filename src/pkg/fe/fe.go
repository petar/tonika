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


package fe

import (
	"bytes"
	"fmt"
	"io/ioutil"
	//"json"
	"net"
	"path"
	"os"
	"strings"
	"sync"
	"template"
	"tonika/http"
	"tonika/sys"
	"tonika/vault"
	"tonika/util/misc"
)

type FrontEnd struct {
	server      *http.AsyncServer
	wwwclient   *http.AsyncClient
	tdir, sdir  string
	bank        sys.Bank
	vault       vault.Vault
	adminURL    string
	myURL       string

	tmplPage     *template.Template
	tmplAdmin    *template.Template
	tmplAdd      *template.Template
	tmplEdit     *template.Template
	tmplBug      *template.Template
	tmplAccept   *template.Template
	tmplReinvite *template.Template
	tmplRoot     *template.Template
	tmplMonitor  *template.Template

	useragent string
	lk        sync.Mutex
}

// tdir = fe templates dir
// sdir = fe static files dir
func MakeFrontEnd(
	tdir,sdir string, 
	addr string, 
	allow string,
	bank sys.Bank, 
	vault vault.Vault,
	myid sys.Id) (*FrontEnd, os.Error) {

	fe := &FrontEnd{
		tdir:     tdir,
		sdir:     sdir,
		bank:     bank,
		vault:    vault,
		adminURL: "http://a."+sys.Host,
		myURL:    sys.MakeURL("", myid, ""),
	}
	err := fe.loadTmpls()
	if err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	fe.wwwclient = http.NewAsyncClient(30e9, 2, 3, 60)
	fe.server = http.NewAsyncServer(l, 20e9, 100)
	fe.server.SetAllowHosts(http.MakeAllowHosts(allow))
	go fe.serveLoop()
	return fe, nil
}

func (fe *FrontEnd) String() string {
	fe.lk.Lock()
	defer fe.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "UserAgent: %s\nServerFD: %d/%d, WWWClientFD: %d/%d\n", 
		fe.useragent,
		fe.server.GetFDLimiter().LockCount(), fe.server.GetFDLimiter().Limit(), 
		fe.wwwclient.GetFDLimiter().LockCount(), fe.wwwclient.GetFDLimiter().Limit())
	return w.String()
}

func (fe *FrontEnd) MarshalJSON() ([]byte, os.Error) {
	fe.lk.Lock()
	defer fe.lk.Unlock()
	var w bytes.Buffer
	fmt.Fprintf(&w, "{\"UserAgent\":%s,\"ServerFD\":%d,"+
		"\"ServerFDLim\":%d,\"WWWClientFD\":%d,\"WWWClientFDLim\":%d}", 
		misc.JSONQuote(fe.useragent),
		fe.server.GetFDLimiter().LockCount(), fe.server.GetFDLimiter().Limit(), 
		fe.wwwclient.GetFDLimiter().LockCount(), fe.wwwclient.GetFDLimiter().Limit())
	return w.Bytes(), nil
}

func (fe *FrontEnd) loadTmpls() (err os.Error) {
	fe.tmplPage,err = loadTmpl(fe.tdir, "page.tmpl")
	if err != nil {
		return err
	}
	fe.tmplAdmin,err = loadTmpl(fe.tdir, "admin.tmpl")
	if err != nil {
		return err
	}
	fe.tmplAdd,err = loadTmpl(fe.tdir, "add.tmpl")
	if err != nil {
		return err
	}
	fe.tmplEdit,err = loadTmpl(fe.tdir, "edit.tmpl")
	if err != nil {
		return err
	}
	fe.tmplBug,err = loadTmpl(fe.tdir, "bug.tmpl")
	if err != nil {
		return err
	}
	fe.tmplAccept,err = loadTmpl(fe.tdir, "accept.tmpl")
	if err != nil {
		return err
	}
	fe.tmplReinvite,err = loadTmpl(fe.tdir, "reinvite.tmpl")
	if err != nil {
		return err
	}
	fe.tmplMonitor,err = loadTmpl(fe.tdir, "monitor.tmpl")
	if err != nil {
		return err
	}
	fe.tmplRoot,err = loadTmpl(fe.tdir, "root.tmpl")
	return err
}

func (fe *FrontEnd) serveLoop() {
	for {
		q,err := fe.server.Read()
		if err == nil {
			go fe.serve(q)
		}
	}
}

func (fe *FrontEnd) serve(q *http.Query) {
	req := q.GetRequest()
	//fmt.Printf("Request: %v·%v·%v·%v\n", 
	//	req.URL.Scheme, req.URL.Host, req.URL.Path, req.URL.RawQuery)

	fe.lk.Lock()
	fe.useragent = req.UserAgent
	fe.lk.Unlock()

	rtype, _ := getRequestType(req)
	switch {
	case rtype == feWWWReq:
		fe.serveWWW(q)
	case rtype == feAdminReq:
		fe.serveAdmin(q)
	case rtype == feTonikaReq:
		fe.serveViaVault(q)
	default:
		q.Continue()
		if req.Body != nil {
			req.Body.Close()
		}
		q.Write(newRespBadRequest())
	}
}

func (fe *FrontEnd) serveWWW(q *http.Query) {
	req := q.GetRequest()
	if req.Method == "CONNECT" {
		go fe.wwwConnect(q)
		return
	}
	q.Continue()

	// Rewrite request
	req.Header["Proxy-Connection"] = "", false
	req.Header["Connection"] = "Keep-Alive"
	//req.Header["Keep-Alive"] = "30"
	url := req.URL
	req.URL = nil
	req.RawURL = url.RawPath
	if url.RawQuery != "" {
		req.RawURL += "?" + url.RawQuery
	}
	if url.Fragment != "" {
		req.RawURL += "#" + url.Fragment
	}

	// Dump request, use for debugging
	// dreq, _ := DumpRequest(req, false)
	// fmt.Printf("REQ:\n%s\n", string(dreq))

	resp := fe.wwwclient.Fetch(req)

	// Dump response, use for debugging
	// dresp, _ := DumpResponse(resp, false)
	// fmt.Printf("RESP:\n%s\n", string(dresp))

	resp.Close = false
	if resp.Header != nil {
		resp.Header["Connection"] = "", false
	}
	q.Write(resp)
}

func (fe *FrontEnd) wwwConnect(q *http.Query) {
	req := q.GetRequest()
	resp, conn2 := fe.wwwclient.Connect(req)
	if conn2 == nil {
		q.Continue()
		q.Write(resp)
		return
	}
	err := q.Write(resp)
	if err != nil {
		return
	}
	asc := q.Hijack()
	conn1, r1, _ := asc.Close()
	if conn1 == nil {
		conn2.Close()
		return
	}
	http.MakeBridge(conn1, r1, conn2, nil)
}

func (fe *FrontEnd) serveAdmin(q *http.Query) {
	req := q.GetRequest()
	q.Continue()

	if req.Body != nil {
		req.Body.Close()
	}
	if req.URL == nil {
		q.Write(newRespBadRequest())
		return
	}

	// Static files
	path := path.Clean(req.URL.Path)
	req.URL.Path = path
	const pfxStatic = "/static/"
	if strings.HasPrefix(path, pfxStatic) {
		q.Write(fe.replyStatic(path[len(pfxStatic):]))
		return
	}

	var resp *http.Response
	switch {
	case strings.HasPrefix(path, "/api/"):
		resp = fe.replyAPI(req)
	case strings.HasPrefix(path, "/accept"):
		resp = fe.replyAdminAccept(req)
	case strings.HasPrefix(path, "/add"):
		resp = fe.replyAdminAdd(req)
	case strings.HasPrefix(path, "/bug"):
		resp = fe.replyAdminBug(req)
	case strings.HasPrefix(path, "/edit"):
		resp = fe.replyAdminEdit(req)
	case strings.HasPrefix(path, "/monitor"):
		resp = fe.replyAdminMonitor(req)
	case strings.HasPrefix(path, "/neighbors"):
		resp = fe.replyRoot()
	case strings.HasPrefix(path, "/reinvite"):
		resp = fe.replyAdminReinvite(req)
	case strings.HasPrefix(path, "/"):
		resp = fe.replyAdminMain()
	default:
		q.Write(newRespNotFound())
		return
	}
	//dresp, _ := http.DumpResponse(resp, false)
	//fmt.Printf("RESP:\n%s\n", string(dresp))
	q.Write(resp)
}

func (fe *FrontEnd) replyAPI(req *http.Request) *http.Response { 
	p := req.URL.Path
	p = p[len("/api/"):]
	path := strings.Split(p, "/", -1)
	if len(path) != 1 {
		return newRespNotFound()
	}
	api := path[0]

	query := req.URL.RawQuery
	args,err := http.ParseQuery(query)
	if err != nil {
		return newRespBadRequest()
	}

	switch api {
	case "accept":
		return fe.replyAPIAccept(args)
	case "add":
		return fe.replyAPIAdd(args)
	case "live":
		return fe.replyAPILive(args)
	case "monitor":
		return fe.replyAPIMonitor(args)
	case "revoke":
		return fe.replyAPIRevoke(args)
	case "update":
		return fe.replyAPIUpdate(args)
	case "myinfo":
		return fe.replyAPIMyInfo(args)
	default:
		return newRespBadRequest()
	}
	panic("unreach")
}

func isFile(path string) bool {
	dir, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if dir != nil && dir.IsDirectory() {
		return false
	}
	return true
}

func (fe *FrontEnd) replyStatic(name string) *http.Response {
	full := path.Join(fe.sdir, name)
	if !isFile(full) {
		return newRespNotFound()
	}
	src, err := ioutil.ReadFile(full)
	if err != nil {
		return newRespNotFound()
	}
	return buildResp(string(src))
}
