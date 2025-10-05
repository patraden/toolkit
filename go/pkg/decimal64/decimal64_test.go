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

	dec := decimal64.New(2, 1234)
	assert.Equal(t, uint8(2), dec.Scale())
	assert.Equal(t, int64(1234), dec.ScaledValue())
	assert.Equal(t, "12.34", dec.String())
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

	decA := decimal64.New(2, 100)  // 1.00
	decB := decimal64.New(3, 2000) // 2.000

	_, err := decA.Add(decB)
	require.Error(t, err)

	_, err = decA.Sub(decB)
	require.Error(t, err)
}

func TestAddSubOK(t *testing.T) {
	t.Parallel()

	decA := decimal64.New(2, 150) // 1.50
	decB := decimal64.New(2, 25)  // 0.25

	sum, err := decA.Add(decB)
	require.NoError(t, err)
	assert.Equal(t, int64(175), sum.ScaledValue())
	assert.Equal(t, "1.75", sum.String())

	diff, err := decA.Sub(decB)
	require.NoError(t, err)
	assert.Equal(t, int64(125), diff.ScaledValue())
	assert.Equal(t, "1.25", diff.String())
}

func TestMulIntDivInt(t *testing.T) {
	t.Parallel()

	dec := decimal64.New(3, 1234) // 1.234

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
