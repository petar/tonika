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


// This package implements some math routines like matrix arithmetic, which are
// needed for testing the behavior of the routing code. The implementations
// here of things like matrix multiplication (and such) are inefficient, but
// that's fine.
//
// TODO:
//   (*) MatrixFactor to replicate the meaning of slices for arrays
//
package math

import (
	"fmt"
	"math"
)

type Matrix interface {
	// # of rows
	M() int

	// # of columns
	N() int

	Get(i,j int) float64
	Set(i,j int, v float64)
}

// (===) Dense implementation

// This is a simple  (i.e. inefficient) implementation of the matrix interface
// using a dense representation
type DenseMatrix struct {
	m,n int
	e []float64
}

// Makers

func MakeDense(m,n int) *DenseMatrix {
	return &DenseMatrix{ m, n, make([]float64, m*n) }
}

func Zeros(m,n int) *DenseMatrix {
	return MakeDense(m,n)
}

func Ones(m,n int) *DenseMatrix {
	R := MakeDense(m,n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			R.Set(i,j, 1.0)
		}
	}
	return R
}

func Eye(n int) *DenseMatrix {
	R := MakeDense(n,n)
	for i := 0; i < n; i++ {
		R.Set(i,i, 1.0)
	}
	return R
}

func Nans(m,n int) *DenseMatrix {
	R := MakeDense(m,n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			R.Set(i,j, math.NaN())
		}
	}
	return R
}

func Kron(m,n int, i,j int) *DenseMatrix {
	R := MakeDense(m,n)
	R.Set(i,j, 1.0)
	return R
}

// Methods

func (A *DenseMatrix) M() int {
	return A.m
}

func (A *DenseMatrix) N() int {
	return A.n
}

func (A *DenseMatrix) toric(i,j int) int {
	return i*A.n + j
}

func (A *DenseMatrix) Get(i,j int) float64 {
	return A.e[A.toric(i,j)]
}

func (A *DenseMatrix) Set(i,j int, v float64) {
	A.e[A.toric(i,j)] = v
}

// (===) Arithmetic operations

// Copies A into a new DenseMatrix
func Copy(A Matrix) Matrix {
	R := MakeDense(A.M(), A.N())
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			R.Set(i,j,A.Get(i,j))
		}
	}
	return R
}

// Transposes A into a new DenseMatrix
func Transpose(A Matrix) Matrix {
	R := MakeDense(A.N(), A.M())
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			R.Set(j,i,A.Get(i,j))
		}
	}
	return R
}

// Computes the result of A+B into A and returns A
func AddInto(A,B Matrix) Matrix {
	if A.M() != B.M() || A.N() != B.N() {
		panic("dimension mismatch")
	}
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			A.Set(i,j, A.Get(i,j)+B.Get(i,j))
		}
	}
	return A
}

// Computes the result of A-B into A and returns A
func SubInto(A,B Matrix) Matrix {
	if A.M() != B.M() || A.N() != B.N() {
		panic("dimension mismatch")
	}
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			A.Set(i,j, A.Get(i,j)-B.Get(i,j))
		}
	}
	return A
}

// Compute f*A into A and return A
func ScalarMul(f float64, A Matrix) Matrix {
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			A.Set(i,j, f*A.Get(i,j))
		}
	}
	return A
}

// Computers the result of A*B into a new DenseMatrix which is returned. This
// uses the plain O(n^3) time algorithm
func Mul(A, B Matrix) Matrix {
	if A.N() != B.M() {
		panic("dimension mismatch")
	}
	R := MakeDense(A.M(), B.N())
	for i := 0; i < A.M(); i++ {
		for j := 0; j < B.N(); j++ {
			t := float64(0.0)
			for k := 0; k < A.N(); k++ {
				t += A.Get(i,k) * B.Get(k,j)
			}
			R.Set(i,j,t)
		}
	}
	return R
}

// (===) Statistics, norms, comparisons

// Returns the larger of the two
func Fmax(f,g float64) float64 {
	if f >= g {
		return f
	}
	return g
}

func ApproxEqNanzero(f,g float64, additiveAccu float64) bool {
	if math.IsNaN(f) { f = 0.0 }
	if math.IsNaN(g) { g = 0.0 }
	return math.Fabs(f-g) <= additiveAccu
}

// Entriwise equality up to given additive error and NaN's treated as zeros
func MatrixApproxEqNanzero(A,B Matrix, additiveAccu float64) (bool, int, int) {
	if A.M() != B.M() || A.N() != B.N() {
		panic("dimension mismatch")
	}
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			if !ApproxEqNanzero(A.Get(i,j), B.Get(i,j),
			additiveAccu) {
				return false, i, j
			}
		}
	}
	return true, 0, 0
}

// Computes L-infinity norm of A, ignoring NaN entries
func EllInf(A Matrix) float64 {
	r := float64(0.0)
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			r = Fmax(r, math.Fabs(A.Get(i,j)))
		}
	}
	return r
}

// (===) Pretty printing

func Pretty(A Matrix, f string) string {
	s := ""
	for i := 0; i < A.M(); i++ {
		for j := 0; j < A.N(); j++ {
			s += fmt.Sprintf("%"+f+", ", A.Get(i,j))
		}
		s += "\n"
	}
	return s
}
