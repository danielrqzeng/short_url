package service

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sync"
	"time"
)

//内部变量的定义
var (
	encodeMgrInstance *encodeMgrType
	encodeMgrOnce     sync.Once
)

//encodeMgr ID编码单例
func EncodeMgr() *encodeMgrType {
	encodeMgrOnce.Do(func() {
		encodeMgrInstance = &encodeMgrType{}
		encodeMgrInstance.Init()
	})
	return encodeMgrInstance
}

//encodeMgrType ID编码定义，通过ID自增来生成短码
type encodeMgrType struct {
	ctx    context.Context
	cancel context.CancelFunc
}

//Init init
func (obj *encodeMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	go obj.grLoop(obj.ctx)
}

//grLoop loop
func (obj *encodeMgrType) grLoop(ctx context.Context) {
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

//GetIncID 获取自增id
func (obj *encodeMgrType) GetIncID() (incID uint64, err error) {
	incID, err = IDMgr().MakeID()
	return
}

//Encode 将长网址编码为短网址
func (obj *encodeMgrType) Encode(rawUrl string) (shortUrl string, err error) {
	//是否网址是合法的
	valid, err := LinkMgr().CheckLink(rawUrl)
	if err != nil {
		return
	}
	if !valid {
		err = fmt.Errorf("url=" + rawUrl + " not a good website")
		return
	}

	//logger.MainLogger.Debug("Encode rawUrl=" + rawUrl)
	incID, err := obj.GetIncID()
	if err != nil {
		logger.MainLogger.Debug("Encode rawUrl=" + rawUrl + " fail cuz=" + err.Error())
		return
	}
	//logger.MainLogger.Debug("Encode rawUrl=" + rawUrl)

	showID, err := NumShuffleMgr().Encode(incID)
	if err != nil {
		logger.MainLogger.Debug("Encode rawUrl=" + rawUrl + " fail cuz=" + err.Error())
		return
	}
	//logger.MainLogger.Debug("Encode rawUrl=" + rawUrl)

	b62str := utils.Base62Encode(showID)
	shortCode := b62str
	shortUrl = viper.GetString("url.domain") + shortCode

	//缓存策略
	err = UrlInfoMgr().Add(incID, rawUrl, shortUrl, shortCode)
	if err != nil {
		return
	}
	logger.MainLogger.Debug("Encode rawUrl=" + rawUrl + " to " + shortCode)

	return
}
