// Copyright 2025 The Toolkit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package decimal64

import (
	"fmt"
	"math/big"
	"strconv"
)

const MaxScale = 12

type Decimal64 struct {
	scale uint8
	value int64
}

// New constructs from raw scaled value.
func New(scale uint8, scaledValue int64) (Decimal64, error) {
	if scale > MaxScale {
		return Decimal64{}, ErrMaxScale
	}

	return Decimal64{
		scale: scale,
		value: scaledValue,
	}, nil
}

// FromInt constructs Decimal64 from integer.
func FromInt(valueInt int64, scale uint8) (Decimal64, error) {
	if scale > MaxScale {
		return Decimal64{}, ErrMaxScale
	}

	factor := pow10(scale)
	if willMulOverflow(valueInt, factor) {
		return Decimal64{}, ErrOverflow
	}

	return New(scale, valueInt*factor)
}

// String prints the decimal value.
func (d Decimal64) String() string {
	if d.scale == 0 {
		return strconv.FormatInt(d.value, 10)
	}

	base := pow10(d.scale)

	quotient := d.value / base
	remainder := d.value % base

	sign := ""
	if quotient < 0 || (quotient == 0 && d.value < 0) {
		sign = "-"
	}

	if quotient < 0 {
		quotient = -quotient
	}

	if remainder < 0 {
		remainder = -remainder
	}

	frac := fmt.Sprintf("%0*d", d.scale, remainder)

	return sign + strconv.FormatInt(quotient, 10) + "." + frac
}

func (d Decimal64) Scale() uint8       { return d.scale }
func (d Decimal64) ScaledValue() int64 { return d.value }

func (d Decimal64) Add(other Decimal64) (Decimal64, error) {
	if d.scale != other.scale {
		return Decimal64{}, fmt.Errorf("%w: %d vs %d", ErrScaleMismatch, d.scale, other.scale)
	}

	if willAddOverflow(d.value, other.value) {
		return Decimal64{}, ErrOverflow
	}

	return Decimal64{
		scale: d.scale,
		value: d.value + other.value,
	}, nil
}

func (d Decimal64) Sub(other Decimal64) (Decimal64, error) {
	if d.scale != other.scale {
		return Decimal64{}, fmt.Errorf("%w: %d vs %d", ErrScaleMismatch, d.scale, other.scale)
	}

	return Decimal64{
		scale: d.scale,
		value: d.value - other.value,
	}, nil
}

func (d Decimal64) MulInt(k int64) Decimal64 {
	return Decimal64{
		scale: d.scale,
		value: d.value * k,
	}
}

func (d Decimal64) DivInt(divisor int64) (Decimal64, error) {
	if divisor == 0 {
		return Decimal64{}, ErrDivisionByZero
	}

	return Decimal64{
		scale: d.scale,
		value: d.value / divisor,
	}, nil
}

func willMulOverflow(multiplicand, multiplier int64) bool {
	// Use big.Int to safely detect overflow without complex branching.
	a := new(big.Int).SetInt64(multiplicand)
	b := new(big.Int).SetInt64(multiplier)
	prod := new(big.Int).Mul(a, b)

	return !prod.IsInt64()
}

func willAddOverflow(augend, addend int64) bool {
	// Use big.Int to detect overflow on addition consistently with multiplication.
	a := new(big.Int).SetInt64(augend)
	b := new(big.Int).SetInt64(addend)
	sum := new(big.Int).Add(a, b)

	return !sum.IsInt64()
}
