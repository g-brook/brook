package utils

type TunnelType string

const (
	Http  TunnelType = "http"
	Https TunnelType = "https"
	Tcp   TunnelType = "tcp"
	Udp   TunnelType = "upd"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}

// NilDefault NilDefault[T comparable]
//
//	@Description:
//	@param value
//	@param defaultValue
//	@return T
func NilDefault[T comparable](value T, defaultValue T) T {
	return DefaultValue(!IsNotNil(value), value, defaultValue)
}

// NumberDefault NumberDefault[T Number]
//
//	@Description: If value not eq 0 return value, if not defaultValue.
//	@param value
//	@param defaultValue
//	@return T
func NumberDefault[T Number](value T, defaultValue T) T {
	return DefaultValue(value != 0, value, defaultValue)
}

// IsNotNil IsNotNil[T comparable]
//
//	@Description:
//	@param value
//	@return bool
func IsNotNil[T comparable](value T) bool {
	var zero T
	if zero == value {
		return false
	}
	return true
}

// DefaultValue DefaultValue[T any]
//
//	@Description:
//	@param b
//	@param val
//	@param def
//	@return T
func DefaultValue[T any](b bool, val T, def T) T {
	if b {
		return val
	}
	return def
}
