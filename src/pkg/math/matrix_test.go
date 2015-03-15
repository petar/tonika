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


package math

import (
	"fmt"
	"testing"
	"tonika/util/random"
)

func TestMatrix(t *testing.T) {
	r := random.NewThreadUnsafe(1)
	A := Ones(13, 7)
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			A.Set(i, j, r.Float64())
		}
	}
	// fmt.Printf("%s\n\n", Pretty(A))
	// fmt.Printf("%s\n\n", Pretty(ScalarMul(3,A)))
	X := Mul(A, Transpose(A))
	Y := Mul(A, Transpose(A))
	Z := Copy(Y)
	if eq, _, _ := MatrixApproxEqNanzero(X, Z, 0.0); !eq {
		t.Errorf("Equality expected\n")
	}
	Q := Zeros(13, 7)
	Q = Mul(Q, Transpose(Q)).(*DenseMatrix)
	if eq, _, _ := MatrixApproxEqNanzero(Q, SubInto(Z, X), 0.0); !eq {
		t.Errorf("Equality expected\n")
	}
	fmt.Printf("%s\n", Pretty(X, "g"))
	fmt.Printf("EllInf=%g\n", EllInf(X))
	//t.Logf("%s\n", path.String())
}
