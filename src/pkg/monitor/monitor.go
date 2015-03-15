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


package monitor

import (
	"bytes"
	"fmt"
	"json"
	"net"
	"time"
	"tonika/crypto"
	"tonika/http"
	"tonika/sys"
)

type Monitor struct {
	dumper json.Marshaler
	key    *crypto.CipherMsgPubKey
}

func MakeMonitor(dumper json.Marshaler, reportURL string, every int64) *Monitor {
	key,err := crypto.ParseCipherMsgPubKey(sys.MonitorPubKey)
	if err != nil {
		panic("invalid monitor key")
	}
	mon := &Monitor{dumper, key}
	url,err := http.ParseURL(reportURL)
	if err != nil {
		panic("mon, bad report URL")
	}
	go mon.report(url, every)
	return mon
}

func printJSON(j []byte) {
	var w bytes.Buffer
	err := json.Indent(&w, j, "| ", "        ")
	if err != nil {
		fmt.Printf("mon,json,err: %s\nORIG:\n%s\n", err, string(j))
		return
	}
	fmt.Printf("MON:\n%s\n", w.String())
}

func (mon *Monitor) report(url *http.URL, every int64) {
	i := 0
	for {
		// Sleep between updates
		if i > 0 {
			time.Sleep(every)
		}
		i++

		// Prepare HTTP request
		jj,err := mon.dumper.MarshalJSON()
		if err != nil {
			continue
		}
		//printJSON(jj)

		repbuf, err := EncodeReport(jj, mon.key)
		if err != nil {
			continue
		}
		bodylen, body := http.BytesToBody(repbuf)

		req := &http.Request{
			Method: "POST",
			URL: url,
			Body: body,
			ContentLength: int64(bodylen),
			Close: true,
			UserAgent: sys.Name + "-Client-Monitor",
		}

		// Connect
		conn,err := net.Dial("tcp", "", url.Host)
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			continue
		}
		cc := http.NewClientConn(conn,nil)

		// Send request
		wch := make(chan int, 1)
		req.Body = http.NewRunOnClose(req.Body, func(){ wch <- 1 })
		err = cc.Write(req)
		if err != nil {
			cc.Close()
			conn.Close()
		} else {
			<-wch // wait until the request has been sent
			cc.Close()
			conn.Close()
		}
	}
}
