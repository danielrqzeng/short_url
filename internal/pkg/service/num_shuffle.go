package service

import (
	"fmt"
	"github.com/RQZeng/num-shuffle"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sync"
)

//内部变量的定义
var (
	numShuffleMgrInstance *NumShuffleMgrType
	numShuffleMgrOnce     sync.Once
)

//numShuffleMgr 实例义单例
func NumShuffleMgr() *NumShuffleMgrType {
	numShuffleMgrOnce.Do(func() {
		numShuffleMgrInstance = &NumShuffleMgrType{}
		numShuffleMgrInstance.Init()
	})
	return numShuffleMgrInstance
}

//NumShuffleMgrType 实例定义
type NumShuffleMgrType struct {
	shuffles []*shuffle.ShuffleType
	ranges   [][]uint64
}

//Init init
func (obj *NumShuffleMgrType) Init() {
	obj.shuffles = make([]*shuffle.ShuffleType, 0)
	obj.ranges = make([][]uint64, 0)

	b62Range := [][]string{ //[min,max)
		{"0", "10"},
		{"10", "100"},
		{"100", "1000"},
		{"1000", "10000"},
		{"10000", "100000"},   //5位base62编码,大概包含916132831-14776336=901356495（9亿）个数字
		{"100000", "1000000"}, //6位编码，是我们主要编码，大概包含56800235583-916132832=55884102751(558亿)个数字
		{"1000000", "10000000"},
		{"10000000", "100000000"},
		{"100000000", "1000000000"},
		{"1000000000", "10000000000"},   //10位base62编码
		{"10000000000", "llllllllllll"}, //uint64最大值为18446744073709551615（大概为1800亿亿），对应的base62编码为lYGhA16ahyf
	}

	for _, r := range b62Range {
		b62Min, b62Max := r[0], r[1] //[b62Min,b62Max)
		min, err := utils.Base62Decode(b62Min)
		if err != nil {
			panic("NumShuffleMgrType Base62Decode err=" + err.Error())
		}
		max, err := utils.Base62Decode(b62Max)
		if err != nil {
			panic("NumShuffleMgrType Base62Decode err=" + err.Error())
		}
		s := &shuffle.ShuffleType{}
		err = s.Init(min, max, "test")
		if err != nil {
			panic("NumShuffleMgrType ShuffleType err=" + err.Error())
		}
		obj.shuffles = append(obj.shuffles, s)
		obj.ranges = append(obj.ranges, []uint64{min, max})
	}
}

//Encode ...
func (obj *NumShuffleMgrType) Encode(num uint64) (cipherNum uint64, err error) {
	for idx, r := range obj.ranges {
		if num >= r[0] && num < r[1] {
			s := obj.shuffles[idx]
			cipherNum, err = s.Encode(num)
			return
		}
	}
	err = fmt.Errorf("encode not found range for num=%d", num)
	return
}

//Decode ...
func (obj *NumShuffleMgrType) Decode(num uint64) (plainNum uint64, err error) {
	for idx, r := range obj.ranges {
		if num >= r[0] && num < r[1] {
			s := obj.shuffles[idx]
			plainNum, err = s.Decode(num)
			return
		}
	}
	err = fmt.Errorf("encode not found range for num=%d", num)
	return
}
