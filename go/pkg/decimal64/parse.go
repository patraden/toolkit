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
	"math/big"
)

// Int64FromStr parses a decimal-like string into a scaled int64 with
// rounding half away from zero. It supports plain decimals and scientific
// notation. Returns (0, false) if input is empty/invalid or result overflows int64.
func Int64FromStr(inputStr string, scale int) (int64, bool) {
	// First: exact rational parse (no exponent), keeps exactness for decimals
	rat := new(big.Rat)
	if _, ok := rat.SetString(inputStr); ok {
		tenPow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil) //nolint:mnd
		rat.Mul(rat, new(big.Rat).SetInt(tenPow))

		// Round half away from zero
		half := big.NewRat(1, 2) //nolint:mnd
		if rat.Sign() >= 0 {
			rat.Add(rat, half)
		} else {
			rat.Sub(rat, half)
		}

		resultInt := new(big.Int)
		resultInt.Quo(rat.Num(), rat.Denom())

		if !resultInt.IsInt64() {
			return 0, false
		}

		return resultInt.Int64(), true
	}

	// Fallback: big.Float for exponent forms (e.g., 2e-07)
	bf := new(big.Float).SetPrec(128).SetMode(big.ToNearestAway) //nolint:mnd,varnamelen
	if _, ok := bf.SetString(inputStr); !ok {
		return 0, false
	}

	tenPow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil) //nolint:mnd
	bf.Mul(bf, new(big.Float).SetInt(tenPow))

	// Round half away from zero: add/subtract 0.5 then truncate toward zero
	half := new(big.Float).SetFloat64(0.5) //nolint:mnd
	if bf.Sign() >= 0 {
		bf.Add(bf, half)
	} else {
		bf.Sub(bf, half)
	}

	resultInt := new(big.Int)
	bf.Int(resultInt) // trunc toward zero

	if !resultInt.IsInt64() {
		return 0, false
	}

	return resultInt.Int64(), true
}

// NewFromString parses inputStr at the specified scale and returns Decimal64.
// It uses half-away-from-zero rounding. Returns ErrInvalidDecimal on parse
// failure or int64 overflow.
func NewFromString(inputStr string, scale uint8) (Decimal64, error) {
	scaled, ok := Int64FromStr(inputStr, int(scale))
	if !ok {
		return Decimal64{}, ErrInvalidDecimal
	}

	return New(scale, scaled)
}
