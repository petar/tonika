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


// Tonika Web Server

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"strings"
	"sync"
	"os"
	"template"
	"time"
	"tonika/http"
	"tonika/sys"
	"tonika/util/filewriter"
	"tonika/util/misc"
)

type Server struct {
	as          *http.AsyncServer
	dir         string
	logwriter   *filewriter.FileWriter
	stats       *WWWStats
	statwriter  *filewriter.LatestFileEncoder
	indextmpl   *template.Template
	lk          sync.Mutex
}

type WWWStats struct {
	Views     int64
	Downloads int64
}

func MakeServer(bind,dir string, fdlim int) (*Server, os.Error) {
	statPrefix := dir + "/log/towstats"
	logPrefix := dir + "/log/towlog"
	templateDir := dir + "/template"

	// Stats
	stats := &WWWStats{}
	err := filewriter.LatestDecode(statPrefix, stats)
	if err != nil {
		stats = &WWWStats{}
	}
	statwriter, err := filewriter.MakeLatestFileEncoder(statPrefix)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "tonika-wwwd: Starting with Views=%d, Downloads=%d\n", 
		stats.Views, stats.Downloads)

	// Logger
	logwriter,err := filewriter.MakeFileWriter(logPrefix)
	if err != nil {
		return nil, err
	}

	// Templates
	tmpl,err := loadTmpl(templateDir, "index.html")
	if err != nil {
		return nil, err
	}

	// HTTP Bind
	l,err := net.Listen("tcp",bind)
	if err != nil {
		return nil, err
	}

	// OK
	s := &Server{
		as:         http.NewAsyncServer(l, 20e9, fdlim),
		dir:        dir,
		logwriter:  logwriter,
		stats:      stats,
		statwriter: statwriter,
		indextmpl:  tmpl,
	}
	go s.statLoop()
	return s, nil
}

func (s *Server) statLoop() {
	for {
		s.lk.Lock()
		fmt.Fprintf(os.Stderr, "tonika-wwwd: Views: %d, Downloads: %d                   \r", 
			s.stats.Views, s.stats.Downloads)
		if err := s.statwriter.Encode(s.stats); err != nil {
			fmt.Fprintf(os.Stderr, "tonika-wwwd: Error writing stats: %s\n", err)
		}
		s.lk.Unlock()
		time.Sleep(60e9) // 1 min
	}
}

func (s *Server) getStaticDir() string { return s.dir+"/static" }
func (s *Server) getTemplateDir() string { return s.dir+"/template" }

func (s *Server) Accept() os.Error {
	q,err := s.as.Read()
	if err == nil {
		go s.serve(q)
	}
	return nil
}

func (s *Server) serve(q *http.Query) {
	q.Continue()
	req := q.GetRequest()
	if req.Body != nil {
		req.Body.Close()
	}
	s.log(req, q.RemoteAddr())

	// Path cleanup
	path := path.Clean(req.URL.Path)
	if path == "" {
		path = "/"
	}

	// Host analysis
	hostport := strings.Split(req.Host, ":", 2)
	if len(hostport) == 0 {
		q.Write(newRespBadRequest())
		return
	}
	hostparts := misc.ReverseHost(strings.Split(hostport[0], ".", -1))

	// http://5ttt.org, http://www.5ttt.org
	if len(hostparts) < 3 || hostparts[2] == "www" {
		if path == "/" {
			path = "/index.html"
		}

		if isIndex(path) {
			s.lk.Lock()
			s.stats.Views++
			s.lk.Unlock()
		}
		
		var resp *http.Response
		if isIndex(path) {
			resp = s.replyIndex()
		} else {
			resp = s.replyStatic(path)
		}

		if isDownload(path) && resp.Body != nil {
			resp.Body = http.NewRunOnClose(resp.Body, func() {
					// TODO: This also counts incomplete downloads, but for now
					// it's fine.
					s.lk.Lock()
					s.stats.Downloads++
					s.lk.Unlock()
				})
		}
		q.Write(resp)
		return
	}

	// Remove 5ttt.org from host
	hostparts = hostparts[2:]

	// http://*.a.5ttt.org/*
	if hostparts[0] == "a" {
		q.Write(s.replyStatic("/tadmin.html")) // Trying to access a Tonika Admin
		return
	}

	// http://*.[id].5ttt.org
	if _, err := sys.ParseId(hostparts[0]); err == nil {
		q.Write(s.replyStatic("/turl.html")) // Trying to access a Tonika URL
		return
	}

	// Otherwise
	q.Write(newRespNotFound())
}

func isIndex(path string) bool {
	return strings.HasPrefix(path, "/index.html")
}

func isDownload(path string) bool {
	parts := strings.Split(path, "/", 3)
	return len(parts) >= 2 && parts[1] == "download"
}

func (s *Server) log(req *http.Request, raddr net.Addr) {
	var w bytes.Buffer
	fmt.Fprintf(&w, "Time: %s, From: %s, URL: %s, Agent: %s\n",
		time.LocalTime().Format(time.UnixDate),
		raddr.String(),
		req.URL.String(), req.UserAgent)
	_,err := s.logwriter.Write(w.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "tonika-wwwd: Error writing log: %s\n", err)
	}
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

func (s *Server) replyStatic(name string) *http.Response {
	full := path.Join(s.getStaticDir(), name)
	if !isFile(full) {
		return newRespNotFound()
	}
	src, err := ioutil.ReadFile(full)
	if err != nil {
		return newRespServiceUnavailable()
	}
	return buildResp(string(src))
}

type IndexPageData struct {
	Downloads string
}

func (s *Server) replyIndex() *http.Response {
	data := &IndexPageData{}
	s.lk.Lock()
	data.Downloads = downloadsToString(s.stats.Downloads)
	s.lk.Unlock()

	var w bytes.Buffer
	s.lk.Lock()
	err := s.indextmpl.Execute(data, &w)
	s.lk.Unlock()
	if err != nil {
		return newRespServiceUnavailable()
	}

	return buildResp(w.String())
}

func downloadsToString(d int64) string {
	p := make([]int64, 8)
	k := 0
	for i := 0; i < len(p); i++ {
		p[i] = d % 1000
		if p[i] != 0 {
			k = i+1
		}
		d /= 1000
	}
	if k == 0 {
		return "0"
	}
	var w bytes.Buffer
	fmt.Fprintf(&w, "%d", p[k-1])
	for i := 1; i < k; i++ {
		w.WriteString(",")
		fmt.Fprintf(&w, "%03d", p[k-1-i])
	}
	return w.String()
}
