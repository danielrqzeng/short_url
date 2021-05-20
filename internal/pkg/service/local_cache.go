package service

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/pkg/data"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/model"
	"iyfiysi.com/short_url/internal/pkg/utils"

	//"github.com/allegro/bigcache/v3" //这个适用二进制数据
	"github.com/patrickmn/go-cache" //这个cache，可以用于结构数据
	"sync"
	"time"
)

//内部变量的定义
var (
	localCacheMgrInstance *localCacheMgrType
	localCacheMgrOnce     sync.Once
)

//LocalCacheMgr 本地缓存单例
func LocalCacheMgr() *localCacheMgrType {
	localCacheMgrOnce.Do(func() {
		localCacheMgrInstance = &localCacheMgrType{}
		localCacheMgrInstance.Init()
	})
	return localCacheMgrInstance
}

//localCacheMgrType 实例定义
type localCacheMgrType struct {
	ctx    context.Context
	cancel context.CancelFunc

	//缓存统计
	hint  uint64
	miss  uint64
	total uint64

	cache *cache.Cache
}

//Init init
func (obj *localCacheMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())

	expireSec := time.Second * time.Duration(viper.GetInt("url.localCacheExpire"))
	obj.cache = cache.New(expireSec, 2*time.Minute) //每5min会做一次清理，key到过期时间了，还不会清理，要等到5min
	obj.cache.OnEvicted(func(key string, val interface{}) {
		urlInfo := val.(*data.URLInfo)
		logger.MainLogger.Debug("key=" + urlInfo.ShortCode + " evicted")
		data.URLInfoPool.Put(val)
		if urlInfo.Status != data.URLStatusNotExist && urlInfo.IncrRedirectTime > 0 {
			//TODO ,此处可能还需要想想是否定时更新，否则数据落地会有些慢
			model.URLMgr().IncrRedirectTime(urlInfo.IncrID, urlInfo.IncrRedirectTime, uint64(urlInfo.LastRedirectTs.Unix()))
		}
	})
	obj.total, obj.hint, obj.miss = 0, 0, 0

	go obj.grLoop(obj.ctx)
}

//Get 获取urlinfo
func (obj *localCacheMgrType) Info() string {
	str := ""
	if obj.total != 0 {
		str = fmt.Sprintf("currNum=%d, visit(total=%d,hint=%d,miss=%d,rate=%.2f)",
			obj.cache.ItemCount(),
			obj.total, obj.hint, obj.miss,
			float64(obj.hint)/float64(obj.total))
	} else {
		str = fmt.Sprintf("currNum=%d, visit(total=%d,hint=%d,miss=%d,rate=%f)",
			obj.cache.ItemCount(),
			obj.total, obj.hint, obj.miss,
			0.0)
	}

	return str
}

//grLoop loop
func (obj *localCacheMgrType) grLoop(ctx context.Context) {
	secTick := time.NewTicker(time.Second)
	defer secTick.Stop()
	done := false
	for !done {
		select {
		case <-secTick.C:
			{
				//做任务的每秒回调
				if utils.Now()%100 == 0 {
					logger.MainLogger.Error(obj.Info())
				}
			}
		case <-ctx.Done():
			{
				done = true
			}
		}
	}
}

//Add 添加新缓存
func (obj *localCacheMgrType) Add(urlInfo *data.URLInfo) {
	maxCache := viper.GetInt("url.localCacheNum")
	if obj.cache.ItemCount() >= maxCache {
		obj.cache.DeleteExpired() //主动做一次过期
		if obj.cache.ItemCount() >= maxCache {
			logger.MainLogger.Error("cant add shortCode=" + urlInfo.ShortCode + " cuz nospace")
			return
		}
	}

	//如果已经存在，则删除，并且用新的替换
	_, exist := obj.Get(urlInfo.ShortCode)
	if exist {
		obj.Del(urlInfo.ShortCode)
	}

	//此处做深拷贝，使得cache内部可以使用对象池来管理URLInfo对象，其释放在cache.OnEvicted函数中
	u := data.URLInfoPool.Get().(*data.URLInfo)
	*u = *urlInfo

	expireSec := time.Second * time.Duration(viper.GetInt("url.localCacheExpire"))
	obj.cache.Set(u.ShortCode, u, expireSec)
	logger.MainLogger.Debug("key=" + urlInfo.ShortCode + " add")
	return
}

//Get 获取urlinfo
func (obj *localCacheMgrType) Get(shortCode string) (urlInfo *data.URLInfo, exist bool) {
	tmp, exist := obj.cache.Get(shortCode)
	obj.total++
	if exist {
		urlInfo = tmp.(*data.URLInfo) //此处要不要深拷贝一份出去呢？因为有可能外部的调用函数使用期间被过期回收了（比如外部调用hold住这个变量30s）
		obj.hint++
		logger.MainLogger.Debug("key=" + shortCode + " hint")
	} else {
		obj.miss++
		logger.MainLogger.Debug("key=" + shortCode + " miss")
		obj.cache.DeleteExpired() //如果没有命中，先跑一次过期，因为有可能这个key是过期的了，此举是为了对象池做对象回收
	}
	return
}

//UpdateRedirect 更新重定向次数
func (obj *localCacheMgrType) UpdateRedirect(shortCode string) {
	tmp, exist := obj.cache.Get(shortCode)
	if exist {
		//重新设置
		urlInfo := tmp.(*data.URLInfo)
		urlInfo.RedirectTime++
		urlInfo.LastRedirectTs = time.Now()
		urlInfo.IncrRedirectTime++
		expireSec := time.Second * time.Duration(viper.GetInt("url.localCacheExpire"))
		obj.cache.Set(shortCode, urlInfo, expireSec)
		logger.MainLogger.Debug("key=" + urlInfo.ShortCode + " update")
	}
	return
}

//UpdateRedirect 更新重定向次数
func (obj *localCacheMgrType) Del(shortCode string) {
	logger.MainLogger.Debug("localCacheMgrType delete shortCode=" + shortCode)
	obj.cache.Delete(shortCode)
	return
}
