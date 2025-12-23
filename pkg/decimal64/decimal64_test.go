package decimal64_test

import (
	"math"
	"testing"

	"github.com/patraden/toolkit/pkg/decimal64"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndAccessors(t *testing.T) {
	t.Parallel()

	dec, err := decimal64.New(2, 1234)
	require.NoError(t, err)
	assert.Equal(t, uint8(2), dec.Scale())
	assert.Equal(t, int64(1234), dec.ScaledValue())
	assert.Equal(t, "12.34", dec.String())
}

func TestNewInvalidScaleError(t *testing.T) {
	t.Parallel()

	_, err := decimal64.New(decimal64.MaxScale+1, 0)
	require.Error(t, err)
}

func TestFromInt(t *testing.T) {
	t.Parallel()

	dec, err := decimal64.FromInt(12, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(1200), dec.ScaledValue())
	assert.Equal(t, "12.00", dec.String())

	_, err = decimal64.FromInt(math.MaxInt64, 2)
	require.Error(t, err)

	_, err = decimal64.FromInt(1, decimal64.MaxScale+1)
	require.Error(t, err)
}

func TestAddSubScaleMismatch(t *testing.T) {
	t.Parallel()

	decA, err := decimal64.New(2, 100) // 1.00
	require.NoError(t, err)
	decB, err := decimal64.New(3, 2000) // 2.000
	require.NoError(t, err)

	_, err = decA.Add(decB)
	require.Error(t, err)

	_, err = decA.Sub(decB)
	require.Error(t, err)
}

func TestAddSubOK(t *testing.T) {
	t.Parallel()

	decA, err := decimal64.New(2, 150) // 1.50
	require.NoError(t, err)
	decB, err := decimal64.New(2, 25) // 0.25
	require.NoError(t, err)

	sum, err := decA.Add(decB)
	require.NoError(t, err)
	assert.Equal(t, int64(175), sum.ScaledValue())
	assert.Equal(t, "1.75", sum.String())

	diff, err := decA.Sub(decB)
	require.NoError(t, err)
	assert.Equal(t, int64(125), diff.ScaledValue())
	assert.Equal(t, "1.25", diff.String())
}

func TestAddOverflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    int64
		b    int64
		want bool
	}{
		// Positive overflow cases
		{"max positive + 1", math.MaxInt64, 1, true},
		{"large positive + large positive", math.MaxInt64/2 + 1, math.MaxInt64/2 + 1, true},
		{"max positive + small positive", math.MaxInt64, 100, true},

		// Negative overflow cases
		{"min negative + -1", math.MinInt64, -1, true},
		{"large negative + large negative", math.MinInt64/2 - 1, math.MinInt64/2 - 1, true},
		{"min negative + small negative", math.MinInt64, -100, true},

		// No overflow cases
		{"max positive + 0", math.MaxInt64, 0, false},
		{"0 + max positive", 0, math.MaxInt64, false},
		{"large positive + small positive", math.MaxInt64 / 2, math.MaxInt64 / 2, false},
		{"positive + negative", math.MaxInt64 / 2, -math.MaxInt64 / 2, false},
		{"negative + positive", -math.MaxInt64 / 2, math.MaxInt64 / 2, false},
		{"min negative + 0", math.MinInt64, 0, false},
		{"0 + min negative", 0, math.MinInt64, false},
		{"max positive + min negative", math.MaxInt64, math.MinInt64, false},
		{"min negative + max positive", math.MinInt64, math.MaxInt64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decA, err := decimal64.New(0, tt.a)
			require.NoError(t, err)
			decB, err := decimal64.New(0, tt.b)
			require.NoError(t, err)

			result, err := decA.Add(decB)
			if tt.want {
				require.Error(t, err)
				assert.Equal(t, decimal64.ErrOverflow, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.a+tt.b, result.ScaledValue())
			}
		})
	}
}

func TestMulIntDivInt(t *testing.T) {
	t.Parallel()

	dec, err := decimal64.New(3, 1234) // 1.234
	require.NoError(t, err)

	m := dec.MulInt(2)
	assert.Equal(t, int64(2468), m.ScaledValue())
	assert.Equal(t, "2.468", m.String())

	q, err := dec.DivInt(2)
	require.NoError(t, err)
	assert.Equal(t, int64(617), q.ScaledValue())
	assert.Equal(t, "0.617", q.String())

	_, err = dec.DivInt(0)
	require.Error(t, err)
}

func TestDivIntOverflowMinInt64ByMinusOne(t *testing.T) {
	t.Parallel()

	dec, err := decimal64.New(0, math.MinInt64)
	require.NoError(t, err)

	_, err = dec.DivInt(-1)
	require.Error(t, err)
	assert.Equal(t, decimal64.ErrOverflow, err)
}

func TestPow10ErrorHandling(t *testing.T) {
	t.Parallel()

	// Test FromInt with scale > MaxScale
	_, err := decimal64.FromInt(123, decimal64.MaxScale+1)
	require.Error(t, err)
	assert.Equal(t, decimal64.ErrMaxScale, err)

	// Test FromInt with scale > pow10 table size (13+)
	_, err = decimal64.FromInt(123, 13)
	require.Error(t, err)
	assert.Equal(t, decimal64.ErrMaxScale, err)

	// Test that the error is properly propagated from pow10
	_, err = decimal64.FromInt(123, 20)
	require.Error(t, err)
	assert.Equal(t, decimal64.ErrMaxScale, err)
}
