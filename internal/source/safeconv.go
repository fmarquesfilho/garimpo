package source

import "math"

// SafeInt32 converts an int to int32, clamping to math.MaxInt32 if it overflows.
func SafeInt32(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}
