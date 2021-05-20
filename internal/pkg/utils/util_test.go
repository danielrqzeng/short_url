// gen by iyfiysi at 2021 May 19

// 测试
package utils

import "testing"

func TestMatchReg(t *testing.T) {
	for _, unit := range []struct {
		data     []string
		expected bool
	}{
		{[]string{"127.0.0.1", "127.0.0.1"}, true},
		{[]string{"127.0.0.*", "127.0.0.1"}, true},
		{[]string{"127.0.*.*", "127.0.0.1"}, true},
		{[]string{"126.0.*.*", "127.0.0.1"}, false},
		{[]string{"127.0.*.*", "127.0.0.6"}, true},
		{[]string{"127.*.*.1", "127.0.0.6"}, false},
		{[]string{"127.*.*.1", "127.0.0.1"}, true},
	} {
		actually, err := MatchReg(unit.data[0], unit.data[1])
		if err != nil {
			t.Fatal(err)
		}
		if unit.expected != actually {
			t.Errorf("match regexp %s expected: [%v], actually: [%v]", unit.data, unit.expected, actually)
		}
	}
}
