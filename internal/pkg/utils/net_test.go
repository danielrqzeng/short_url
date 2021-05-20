// gen by iyfiysi at 2021 May 19

// 获取机器网络基本信息

package utils

import "testing"

func TestIsIPAType(t *testing.T) {
	for _, unit := range []struct {
		data     string
		expected bool
	}{
		{"127.0.0.1", false},
		{"192.168.0.100", false},
		{"11.0.0.100", false},
		{"10.0.0.100", true},
	} {
		actually, err := IsIPAType(unit.data)
		if err != nil {
			t.Fatal(err)
		}
		if unit.expected != actually {
			t.Errorf("ip %s expected: [%v], actually: [%v]", unit.data, unit.expected, actually)
		}
	}
}

func TestIsIPBType(t *testing.T) {
	for _, unit := range []struct {
		data     string
		expected bool
	}{
		{"172.30.0.14", true},
	} {
		actually, err := IsIPBType(unit.data)
		if err != nil {
			t.Fatal(err)
		}
		if unit.expected != actually {
			t.Errorf("ip %s expected: [%v], actually: [%v]", unit.data, unit.expected, actually)
		}
	}
}
