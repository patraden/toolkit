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

// Package decimal64 provides a tiny fixed‑point decimal type backed by int64.
//
// The number is represented as value / 10^scale, where value is an int64 and
// scale is a small non-negative integer (up to MaxScale). This is useful for
// precise monetary/math operations without floating-point rounding error.
//
// Example:
//
//	// Parse string with rounding half away from zero to 4 fractional digits
//	d, _ := decimal64.NewFromString("12.34567", 4) // -> 12.3457
//	fmt.Println(d.String())                         // "12.3457"
//	fmt.Println(d.ScaledValue())                    // 123457
//
// You can also construct from integers:
//
//	d, _ := decimal64.FromInt(12, 2) // 12.00
//
// Arithmetic requires equal scales:
//
//	a := decimal64.New(2, 150) // 1.50
//	b := decimal64.New(2, 25)  // 0.25
//	sum, _ := a.Add(b)         // 1.75
package decimal64
