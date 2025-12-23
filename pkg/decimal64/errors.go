package decimal64

import "errors"

var (
	ErrScaleMismatch  = errors.New("scale mismatch")
	ErrOverflow       = errors.New("decimal64 overflow")
	ErrMaxScale       = errors.New("scale exceeded max")
	ErrDivisionByZero = errors.New("division by zero")
	ErrInvalidDecimal = errors.New("invalid decimal string")
)
