package helpers

import (
	"strconv"
)

func ParseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
