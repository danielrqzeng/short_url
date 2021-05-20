package utils

import (
	"testing"
)

func TestBase62(t *testing.T) {
	testData := []struct {
		n        uint64
		expected string
	}{
		{0, "0"},
		{10, "a"},
		{630, "aa"},
		{2222821365901088, "abc123EFG"},
		{3781504209452600, "hjNv8tS3K"},
	}
	for _, testCase := range testData {
		r := Base62Encode(testCase.n)
		if r != testCase.expected {
			t.Fatalf("encode expected '%v', but got '%v'", testCase.expected, r)
		}
		d, err := Base62Decode(r)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if d != testCase.n {
			t.Fatalf("decode expected '%v', but got '%v'", testCase.n, d)
		}
	}
}
