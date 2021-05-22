package data

import (
	"sync"
	"time"
)

//URLInfo url info,原始url&短网址url&一些创建&统计信息
type URLInfo struct {
	IncrID    uint64 //自增id
	RawUrl    string //原始url,比如：https://www.baidu.com/
	ShortUrl  string //短网址url,比如:https://surl4.me/20QQ20
	ShortCode string //短码20QQ20

	//创建信息
	CreateAt time.Time
	URLType  uint64 //类型信息,data.URLType*
	Status   uint64 //状态信息,data.URLStatus*

	//禁止信息
	BanAt   time.Time //禁止时间
	UnBanAt time.Time //解除禁止时间
	BanCuz  string    //禁止原因
	BanBy   string    //被谁禁止，sys代表是系统自动禁止，其他的自定义名称，能标识即可

	RedirectTime     uint64    //重定向次数（一般用302暂时定向，301为永久定向，不利于做统计）
	LastRedirectTs   time.Time //最近一次重定向时间
	IncrRedirectTime uint64    //新增的重定向次数
}

//URLInfoPool URLInfo的对象池
var URLInfoPool = sync.Pool{
	New: func() interface{} {
		return new(URLInfo)
	},
}

//BaseResponse 通用回应
type BaseResponse struct {
	RetCode int32  `json:"retCode"` //0:成功，其他失败
	RetMsg  string `json:"retMsg"`  //对retCode的描述
	MsgShow string `json:"msgShow"` //显示什么信息
}
