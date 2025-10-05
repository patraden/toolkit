package decimal64_test

import (
	"testing"

	"github.com/patraden/toolkit/pkg/decimal64"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		scale uint8
		want  int64
		ok    bool
	}{
		{"positive decimal", "0.00035", 10, 3500000, true},
		{"scientific notation 1", "2e-07", 10, 2000, true},
		{"scientific notation 2", "3e-07", 10, 3000, true},
		{"negative decimal string", "-0.00035", 10, -3500000, true},
		{"empty string", "", 10, 0, false},
		{"bad string", "abc", 10, 0, false},
		{"positive decimal > 1", "1.2345678901", 10, 12345678901, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dec, err := decimal64.NewFromString(tt.in, tt.scale)
			if !tt.ok {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.scale, dec.Scale())
			assert.Equal(t, tt.want, dec.ScaledValue())
		})
	}
}
