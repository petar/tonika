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

// This test demonstrates how to write a transparent HTTP/1.1
// proxy, using AsyncClient and AsyncServerConn.

package http

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"
	"tonika/prof"
	"tonika/util/term"
)

var t0 = time.Seconds()
var cc = NewAsyncClient(10e9, 2)

var lk sync.Mutex
var nreq, nresp int

func increq() {
	lk.Lock()
	nreq++
	lk.Unlock()
}

func incresp() {
	lk.Lock()
	nresp++
	lk.Unlock()
}

func prr(pfx string) {
	lk.Lock()
	a := nreq
	b := nresp
	lk.Unlock()
	fmt.Printf(term.Reverse+"          -%s-> %d/%d <--\n"+term.Reset, pfx, a, b)
}

func ds() int64 { return time.Seconds() - t0 }

func sigh() {
	for s := range signal.Incoming {
		fmt.Printf(term.FgRed+"[%v] sig: %s\n"+term.Reset, ds(), s)
		fmt.Println(prof.String())
		os.Exit(0)
	}
}

func pfd() {
	/*
		fmt.Printf(term.FgWhite+"·(FD=??, ALL=%d, STA=%d, INU=%d)\n"+term.Reset,
			runtime.MemStats.Alloc, runtime.MemStats.Stacks,
			runtime.MemStats.InusePages)
	*/
}

func proxyError(asc *AsyncServerConn, err os.Error) {
	asc.Close(justClose)
}

func proxyServer(asc *AsyncServerConn, req *Request, onresp func(resp *Response)) {
	defer prof.SecLeave(prof.SecEnter())
	fmt.Printf(term.FgMagenta + "*request\n" + term.Reset)
	increq()
	prr("REQ")
	pfd()


	/* if req.URL.Host == "maps.gstatic.com" */ {
		d, err2 := DumpRequest(req, true)
		if err2 != nil {
			panic()
		}
		fmt.Print(string(d))
	}
	fmt.Printf("Raw URL: %s\n", req.RawURL)

	switch req.Method {
	case "CONNECT":
		go func() {
			cc.Connect(req, func(req *Request, resp *Response, cconn net.Conn) {
				if cconn != nil {
					// First, respond
					onresp(resp)
					incresp()
					prr("CONN")
					// Then, close AsyncServerConn
					asc.Close(func(sconn net.Conn, sr *bufio.Reader) {
						if sconn == nil {
							cconn.Close()
							return
						}
						MakeBridge(sconn, sr, cconn, nil)
					})
				} else {
					asc.ReadMore()
					onresp(resp)
					incresp()
					prr("CONO")
				}
			})
		}()
		return
	default:
		//case "GET","POST","HEAD":
		asc.ReadMore()
		break
		/*
			default:
				fmt.Printf(term.FgRed+"*unknown: "+term.Reset+"%s\n", req.Method)
				asc.ReadMore()
				if req.Body != nil {
					req.Body.Close()
					req.Body = nil
				}
				onresp(respErrServiceUnavailable)
				incresp()
				prr("RESP")
				return
		*/
	}

	/*
		if req.Method == "GET" || req.Method == "HEAD" {
			req.Body.Close()
			req.Body = nil
		}
	*/
	req.Header["Content-Length"] = "", false
	req.Header["Proxy-Connection"] = "", false

	cc.Fetch(req, func(req *Request, resp *Response) { proxyClient(req, resp, onresp) })
}

var dummy = "<html>" +
	"<head><title>Oxy placeholder</title></head>\n" +
	"<body bgcolor=\"white\">\n" +
	"<center><h1>Oxy placeholder</h1></center>\n" +
	"<hr><center>Go HTTP package</center>\n" +
	"</body></html>"

func proxyClient(req *Request, resp *Response, onresp func(resp *Response)) {
	go func() {
		pfd()
		if resp != nil {
			//if resp.Body != nil {
			//	resp.Body.Close()
			//}
			//resp.Body = StringToBody(dummy)
			//resp.ContentLength = int64(len(dummy))
			resp.Close = false
			if resp.Header != nil {
				resp.Header["Connection"] = "", false
			}

			d, err := DumpResponse(resp, false)
			if err != nil {
				panic()
			}
			fmt.Printf("RESPONDING:\n%s\n", string(d))

		}
		onresp(resp)
		incresp()
		prr("RESP")
	}()
}

func profLoop() {
	for {
		<-alarmOnce(5e9)
		fmt.Println(prof.String())
	}
}

func TestProxy(t *testing.T) {
	go sigh()
	l, err := net.Listen("tcp", ":4949")
	if err != nil {
		t.Fatalf("listen: %s", err)
		return
	}
	k := 0
	for {
		c, err := l.Accept()
		if err != nil {
			panic("accept break")
			return
		}
		k++
		fmt.Printf(term.FgMagenta+"*accept=%d"+term.Reset+"\n", k)
		NewAsyncServerConn(30e9, c, proxyServer, proxyError)
		pfd()
	}
	fmt.Printf(term.FgRed + "graceful exit" + term.Reset)
}
