package utils

import (
	"fmt"
	"math"
	"strings"
)


const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base     = uint64(len(alphabet))
)

// Base62Encode decoded integer to base62 string.
func Base62Encode(n uint64) string {
	if n == 0 {
		return "0"
	}

	b := make([]byte, 0, 512)
	for n > 0 {
		r := math.Mod(float64(n), float64(base))
		n /= base
		b = append([]byte{alphabet[int(r)]}, b...)
	}
	return string(b)
}

// Base62Decode a base62 encoded string to int.
// Returns an error if input s is not valid base62 literal [0-9a-zA-Z].
func Base62Decode(s string) (uint64, error) {
	var r uint64
	for _, c := range []byte(s) {
		i := strings.IndexByte(alphabet, c)
		if i < 0 {
			return 0, fmt.Errorf("unexpected character %c in base62 literal", c)
		}
		r = base*r + uint64(i)
	}
	return r, nil
}