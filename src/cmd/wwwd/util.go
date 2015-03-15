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


package main

import (
	"io/ioutil"
	"os"
	"path"
	"template"
	"tonika/http"
)

func loadTmpl(tdir, name string) (tmpl *template.Template, err os.Error) {
	path := path.Join(tdir, name)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	t := template.New(nil)
	t.SetDelims("[[", "]]")
	err = t.Parse(string(data))
	if err != nil {
		return nil, err
	}
	return t, nil
}

func newRespServiceUnavailable() *http.Response {
	htmlErrServiceUnavailable := "<html>" +
		"<head><title>503 Service Unavailable</title></head>\n" +
		"<body bgcolor=\"white\">\n" +
		"<center><h1>503 Service Unavailable</h1></center>\n" +
		"<hr><center>Tonika Web Server</center>\n" +
		"</body></html>"
	respErrServiceUnavailable := &http.Response{
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
	return respErrServiceUnavailable
}

func newRespBadRequest() *http.Response {
	htmlErrBadRequest := "<html>" +
		"<head><title>400 Bad Request</title></head>\n" +
		"<body bgcolor=\"white\">\n" +
		"<center><h1>400 Bad Request</h1></center>\n" +
		"<hr><center>Tonika Web Server</center>\n" +
		"</body></html>"
	respErrBadRequest := &http.Response{
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
	return respErrBadRequest
}

func newRespNotFound() *http.Response {
	htmlErrNotFound := "<html>" +
		"<head><title>404 Resource Not Found</title></head>\n" +
		"<body bgcolor=\"white\">\n" +
		"<center><h1>404 Resource Not Found</h1></center>\n" +
		"<hr><center>Tonika Web Server</center>\n" +
		"</body></html>"
	respErrNotFound := &http.Response{
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
	return respErrNotFound
}

func buildResp(html string) *http.Response {
	resp := &http.Response{
		Status: "OK",
		StatusCode: 200,
		Proto: "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RequestMethod: "GET",
		Close: false,
	}
	resp.Body = http.StringToBody(html)
	resp.ContentLength = int64(len(html))
	return resp
}
