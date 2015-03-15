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


package eye64

import (
	"fmt"
	"rand"
	"testing"
)

func TestEye64(t *testing.T) {
	fmt.Printf("--> Print some would be addresses for looks:\n")
	for i := 0; i < 5; i++ {
		fmt.Printf("http://xyz/%s/\n", U64ToEye(uint64(rand.Int63())))
	}
	fmt.Printf("\n")

	fmt.Printf("--> Print 0, 1, 3, ...\n")
	for i := 0; i < 2*base+1; i++ {
		fmt.Printf("%d = %s\n", i, U64ToEye(uint64(i)))
	}
	fmt.Printf("\n")
	
	x := int64(-1234)
	u1 := uint64(x)
	u2 := uint64(1234567890)
	u3 := uint64(12345678901234)
	u4 := uint64(1<<64 - 1)
	fmt.Printf("%d = %s\n", u1, U64ToEye(u1))
	fmt.Printf("%d = %s\n", u2, U64ToEye(u2))
	fmt.Printf("%d = %s\n", u3, U64ToEye(u3))
	fmt.Printf("%d = %s\n", u4, U64ToEye(u4))
	f1 := U64ToEye(u1)
	f2 := U64ToEye(u2)
	if v, err := EyeToU64(f1); v != u1 {
		t.Errorf("U->E->U mismatch: expect %d vs. seen %d, err: %s", u1, v, err)
	}
	if v, err := EyeToU64(f2); v != u2 {
		t.Errorf("U->E->U mismatch: expect %d vs. seen %d, err: %s", u2, v, err)
	}

	/*
	e1 := "ab678"
	e2 := "ghkp8k"
	w1, _ := EyeToU64(e1)
	w2, _ := EyeToU64(e2)
	if U64ToEye(w1) != e1 {
		t.Errorf("E->U->E mismatch: expect %s vs. seen %s", e1, U64ToEye(w1))
	}
	if U64ToEye(w2) != e2 {
		t.Errorf("E->U->E mismatch: expect %s vs. seen %s", e2, U64ToEye(w2))
	}
	*/
}
