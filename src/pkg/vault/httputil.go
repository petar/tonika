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


// *** SRC ***

package vault

import (
	"io"
	"io/ioutil"
	"os"
	gopath "path"
	"strconv"
	"strings"
	"sync"
	"template"
	"tonika/http"
	"tonika/sys"
	"tonika/util/eye64"
	"tonika/util/misc"
)

func parseURL(req *http.Request) (tid sys.Id, path, query string, err os.Error) {
	
	hostparts := strings.Split(req.Host, ".", -1)
	hostparts = misc.ReverseHost(hostparts)

	if len(hostparts) != 3 || hostparts[0] != sys.Host0 || hostparts[1] != sys.Host1 {
		return 0, "", "", os.EINVAL
	}

	tid, err = sys.ParseId(hostparts[2])
	if err != nil {
		return 0, "", "", os.EINVAL
	}

	query = req.URL.RawQuery
	path = gopath.Clean(req.URL.Path)

	return tid, path, query, nil
}

// Sanitize

func sanitizeResp(resp *http.Response) {
	if resp.Header == nil {
		return
	}
	resp.Header["Vault-Build"] = "", false
	resp.Header["Vault-Proto"] = "", false
	resp.Header["Vault-Hop"] = "", false
	resp.Header["Vault-Origin"] = "", false
}

// Version

func setReqVersion(req *http.Request) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}
	req.Header["Vault-Build"] = sys.Build
	req.Header["Vault-Proto"] = Version
}

func setRespVersion(resp *http.Response) {
	if resp.Header == nil {
		resp.Header = make(map[string]string)
	}
	resp.Header["Vault-Build"] = sys.Build
	resp.Header["Vault-Proto"] = Version
}

// Hops

func setReqHop(req *http.Request, h int) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}
	req.Header["Vault-Hop"] = strconv.Itoa(h)
}

func setRespHop(resp *http.Response, h int) {
	if resp.Header == nil {
		resp.Header = make(map[string]string)
	}
	resp.Header["Vault-Hop"] = strconv.Itoa(h)
}

func parseReqHop(req *http.Request) (h int, err os.Error) { return parseHop(req.Header) }
func parseRespHop(resp *http.Response) (h int, err os.Error) { return parseHop(resp.Header) }

func parseHop(header map[string]string) (h int, err os.Error) {
	if header == nil {
		return -1, os.ErrorString("missing hop")
	}
	hop,ok := header["Vault-Hop"]
	if !ok {
		return -1, os.ErrorString("missing hop")
	}
	h,err = strconv.Atoi(hop)
	if err != nil {
		return -1, err
	}
	if h < 0 {
		return -1, os.ErrorString("invalid hop")
	}
	return h,err
}

// Origin

func setOrigin(req *http.Request, id sys.Id) {
	if req.Header == nil {
		req.Header = make(map[string]string)
	}
	req.Header["Vault-Origin"] = id.Eye()
}

func parseOrigin(req *http.Request) (oid sys.Id, err os.Error) {
	if req.Header == nil {
		err = os.EINVAL
		return
	}
	orig, ok := req.Header["Vault-Origin"]
	if !ok {
		err = os.EINVAL
		return
	}
	u64, err := eye64.EyeToU64(strings.TrimSpace(orig))
	if err != nil {
		err = os.EINVAL
		return
	}
	return sys.Id(u64), nil
}

// Templates

func loadTmpl(tdir, name string) (tmpl *template.Template, err os.Error) {
	path := gopath.Join(tdir, name)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return template.Parse(string(data), nil)
}

var (
	// Service unavailable
	htmlErrServiceUnavailable = "<html>" +
		"<head><title>503 Service Unavailable</title></head>\n" +
		"<body bgcolor=\"white\">\n" +
		"<center><h1>503 Service Unavailable</h1></center>\n" +
		"<hr><center>"+sys.Name+" Front End</center>\n" +
		"</body></html>"
	respErrServiceUnavailable = &http.Response{
		Status: "Service Unavailable",
		StatusCode: 503,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RequestMethod: "GET",
		Body: http.StringToBody(htmlErrServiceUnavailable),
		ContentLength: int64(len(htmlErrServiceUnavailable)),
		Close: false,
	}
	// Not found
	htmlErrNotFound = "<html>" +
		"<head><title>404 Resource Not Found</title></head>\n" +
		"<body bgcolor=\"white\">\n" +
		"<center><h1>404 Resource Not Found</h1></center>\n" +
		"<hr><center>"+sys.Name+" Front End</center>\n" +
		"</body></html>"
	respErrNotFound = &http.Response{
		Status: "Resource Not Found",
		StatusCode: 404,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RequestMethod: "GET",
		Body: http.StringToBody(htmlErrNotFound),
		ContentLength: int64(len(htmlErrNotFound)),
		Close: false,
	}
	// Bad request
	htmlErrBadRequest = "<html>" +
		"<head><title>400 Bad Request</title></head>\n" +
		"<body bgcolor=\"white\">\n" +
		"<center><h1>400 Bad Request</h1></center>\n" +
		"<hr><center>"+sys.Name+" Front End</center>\n" +
		"</body></html>"
	respErrBadRequest = &http.Response{
		Status:        "Bad Request",
		StatusCode:    400,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		RequestMethod: "GET",
		Body:          http.StringToBody(htmlErrBadRequest),
		ContentLength: int64(len(htmlErrBadRequest)),
		Close:         false,
	}
	// OK
	respOK = &http.Response{
		Status: "OK",
		StatusCode: 200,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RequestMethod: "GET",
		Close: false,
	}
	blk sync.Mutex
)

func statusCodeSupported(code int) bool {
	switch code {
	case 200,400,404,503:
		return true
	default:
		return false
	}
	panic("unreach")
}

func newRespServiceUnavailable() *http.Response {
	blk.Lock()
	defer blk.Unlock()
	r, err := http.DupResp(respErrServiceUnavailable)
	if err != nil {
		panic("v")
	}
	return r
}

func newRespBadRequest() *http.Response {
	blk.Lock()
	defer blk.Unlock()
	r, err := http.DupResp(respErrBadRequest)
	if err != nil {
		panic("v")
	}
	return r
}

func newRespNotFound() *http.Response {
	blk.Lock()
	defer blk.Unlock()
	r, err := http.DupResp(respErrNotFound)
	if err != nil {
		panic("v")
	}
	return r
}

func buildResp(html string) *http.Response {
	blk.Lock()
	defer blk.Unlock()
	resp,err := http.DupResp(respOK)
	if err != nil {
		panic("v")
	}
	resp.Body = http.StringToBody(html)
	resp.ContentLength = int64(len(html))
	return resp
}

// If bodylen is negative, chunked encoding is used.
func buildRespFromBody(body io.ReadCloser, bodylen int64) *http.Response {
	blk.Lock()
	defer blk.Unlock()
	resp,err := http.DupResp(respOK)
	if err != nil {
		panic("v")
	}
	resp.Body = body
	if bodylen >= 0 {
		resp.ContentLength = bodylen
	} else {
		resp.TransferEncoding = []string{"chunked"}
	}
	return resp
}

func newRespUnsupported() *http.Response {
	return buildResp("Your friend uses a newer (unsupported) software. Please update.")
}

func newRespNoIndexHTML() *http.Response {
	return buildResp("No index.html file is present.")
}
