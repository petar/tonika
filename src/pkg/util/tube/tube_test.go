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


package tube

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

// Test gob works when chunks sent are larger than bufio.Buffer size

type block struct {
	x, y, z float64
	p       string
}

var btest = block{87.34, 424234.3423, -253.43e93, "vkjfkeiruhtdsgfsd"}
var kac = []byte("abcdefg")
var kca = []byte("klfjlsfdoe")

func accept(ch chan int, t *testing.T) {
	l, err := net.Listen("tcp", ":58777")
	if err != nil {
		t.Fatalf("lis")
	}
	c, err := l.Accept()
	if err != nil {
		t.Fatalf("acc")
	}
	err = l.Close()
	if err != nil {
		t.Fatalf("acc, lcl")
	}
	br, err := bufio.NewReaderSize(c, 10)
	if err != nil {
		t.Fatalf("acc, br")
	}
	tu := NewTube(c, br)

	fmt.Printf(" Acc: Intro auth ...\n")
	tx, err := authIntro(tu)
	if err != nil {
		t.Fatalf("acc, auth: %s", err)
	}
	fmt.Printf(" Acc: Intro auth done.\n")

	err = tx.Encode(&btest)
	if err != nil {
		t.Fatalf("acc, enc")
	}

	err = tx.Close()
	if err != nil {
		t.Fatalf("acc, clo")
	}
	ch <- 1
}

func connect(ch chan int, t *testing.T) {
	c, err := net.Dial("tcp", "", ":58777")
	if err != nil {
		t.Fatalf("con, dia")
	}
	br, err := bufio.NewReaderSize(c, 8)
	if err != nil {
		t.Fatalf("con, br")
	}
	tu := NewTube(c, br)

	fmt.Printf(" Conn: Intro auth ...\n")
	tx, err := authIntro(tu)
	if err != nil {
		t.Fatalf("acc, auth: %s", err)
	}
	fmt.Printf(" Conn: Intro auth done.\n")

	b := block{}
	err = tx.Decode(&b)
	if err != nil {
		t.Fatalf("con, dec: %s", err)
	}

	diff(t, "block", &btest, &b)

	err = tx.Close()
	if err != nil {
		t.Fatalf("acc, clo")
	}
	ch <- 1
}

func diff(t *testing.T, prefix string, have, want interface{}) {
	hv := reflect.NewValue(have).(*reflect.PtrValue).Elem().(*reflect.StructValue)
	wv := reflect.NewValue(want).(*reflect.PtrValue).Elem().(*reflect.StructValue)
	if hv.Type() != wv.Type() {
		t.Errorf("%s: type mismatch %v vs %v", prefix, hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s: %s = %v want %v", prefix, hv.Type().(*reflect.StructType).Field(i).Name, hf, wf)
		}
	}
}

func TestTube(t *testing.T) {
	ch := make(chan int)
	fmt.Printf("Test without RC4-wrapper ...\n")
	go accept(ch, t)
	time.Sleep(2e8)
	go connect(ch, t)
	<-ch
	<-ch
}
