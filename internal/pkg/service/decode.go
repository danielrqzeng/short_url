package service

import (
	"context"
	"fmt"
	"iyfiysi.com/short_url/internal/pkg/data"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sync"
	"time"
)

//内部变量的定义
var (
	decodeMgrInstance *decodeMgrType
	decodeMgrOnce     sync.Once
)

//DecodeMgr 实例义单例
func DecodeMgr() *decodeMgrType {
	decodeMgrOnce.Do(func() {
		decodeMgrInstance = &decodeMgrType{}
		decodeMgrInstance.Init()
	})
	return decodeMgrInstance
}

//decodeMgrType 实例定义
type decodeMgrType struct {
	ctx    context.Context
	cancel context.CancelFunc

	maxIncID uint64 //短域名，自增id最大值
}

//Init init
func (obj *decodeMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	obj.maxIncID = 0
	go obj.grLoop(obj.ctx)
}

//grLoop loop
func (obj *decodeMgrType) grLoop(ctx context.Context) {
	secTick := time.NewTicker(time.Second)
	defer secTick.Stop()
	done := false
	for !done {
		select {
		case <-secTick.C:
			{
				//做任务的每秒回调
			}
		case <-ctx.Done():
			{
				done = true
			}
		}
	}
}

//CheckIncID 检查自增id
func (obj *decodeMgrType) CheckIncID() (err error) {
	return
}

//Encode 将长网址编码为短网址
func (obj *decodeMgrType) Decode(shortCode string) (rawUrl string, err error) {
	showID, err := utils.Base62Decode(shortCode)
	if err != nil {
		return
	}

	incID, err := NumShuffleMgr().Decode(showID)
	if err != nil {
		return
	}
	logger.MainLogger.Debug(fmt.Sprintf("decode shortUrl=%s,showID=%d to incID=%d\n", shortCode, showID, incID))

	valid := false
	//是否已经存在在用的短语库中，若存在，则可以解码
	if !valid {
		valid, err = PhraseMgr().IsBeenUsedByPhrase(shortCode)
		if err != nil {
			return
		}
	}

	//不是短语库的短码，则认为其实id生成的短码，此时检测其范围，此步骤可以简单拦截掉大部分乱来的请求
	if !valid {
		valid, err = IDMgr().IsIDValid(incID)
		if err != nil {
			return
		}
	}
	if !valid {
		err = fmt.Errorf("shortUrl=%s,id=%d is not valid", shortCode, incID)
		return
	}

	urlInfo, err := UrlInfoMgr().Get(shortCode)
	if err != nil {
		err = fmt.Errorf("UrlInfoMgr.Get err=%s", err.Error())
		return
	}
	if urlInfo.Status != data.URLStatusEnable {
		err = fmt.Errorf("no available for shortCode=" + urlInfo.ShortCode)
	}
	rawUrl = urlInfo.RawUrl

	return
}
