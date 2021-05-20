package service

import (
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/viper"
	"sync"
	"time"
)

//内部变量的定义
var (
	LinkMgrInstance *LinkMgrType
	LinkMgrOnce     sync.Once
)

//LinkMgr 用户连接管理实例
func LinkMgr() *LinkMgrType {
	LinkMgrOnce.Do(func() {
		LinkMgrInstance = &LinkMgrType{}
		LinkMgrInstance.Init()
	})
	return LinkMgrInstance
}

//LinkMgrType 实例定义
type LinkMgrType struct {
	ctx    context.Context
	cancel context.CancelFunc
}

//Init init
func (obj *LinkMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	go obj.grLoop(obj.ctx)
}

//grLoop loop
func (obj *LinkMgrType) grLoop(ctx context.Context) {
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

func (obj *LinkMgrType) IsUrl(link string) (isUrl bool) {
	isUrl = govalidator.IsURL(link)
	return
}

//CheckLink 检测链接是否合法,valid=true代表合法
func (obj *LinkMgrType) CheckLink(link string) (valid bool, err error) {
	valid = true
	if !obj.IsUrl(link) {
		valid = false
		err = fmt.Errorf("not valid link")
		return
	}

	if viper.GetBool("link.doCheck") {
		valid, err = IsUrlSecurity(link)
		return
	}

	return
}
