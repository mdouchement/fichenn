package ftime

import (
	"errors"
	"strconv"
	"time"
)

// MustParseDuration calls ParseDuration and panics if an error occurs.
func MustParseDuration(s string) time.Duration {
	d, err := ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

// ParseDuration parses a duration string.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300s", "5d" or "2h45m".
// Valid time units are "s", "m", "h", "d".
func ParseDuration(s string) (time.Duration, error) {
	units := map[rune]time.Duration{
		'd': 24 * time.Hour,
		'h': time.Hour,
		'm': time.Minute,
		's': time.Second,
	}
	var d time.Duration
	var digits []byte

	for _, r := range s {
		switch r {
		case 'd', 'h', 'm', 's':
			v, _ := strconv.Atoi(string(digits))
			d += time.Duration(v) * units[r]
			digits = digits[:0]
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			digits = append(digits, byte(r))
		default:
			return d, errors.New("ftime: invalid duration " + strconv.Quote(s))
		}
	}

	return d, nil
}
