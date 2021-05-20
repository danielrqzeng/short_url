package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/pkg/data"
	"iyfiysi.com/short_url/internal/pkg/db"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/model"
	"sync"
	"time"
)

//内部变量的定义
var (
	urlInfoMgrInstance *urlInfoMgrType
	urlInfoMgrOnce     sync.Once
)

func ToTURLInfo(u *data.URLInfo) (tu *model.TURLInfo) {
	tu = &model.TURLInfo{}
	tu.Id = u.IncrID
	tu.RawUrl = u.RawUrl
	tu.ShortUrl = u.ShortUrl
	tu.ShortCode = u.ShortCode

	tu.CreateTs = int(u.CreateAt.Unix())
	tu.UrlType = int(u.URLType)
	tu.Status = int(u.Status)

	tu.BanAt = int(u.BanAt.Unix())
	tu.UnbanAt = int(u.UnBanAt.Unix())
	tu.BanCuz = u.BanCuz
	tu.BanBy = u.BanBy

	tu.RedirectTime = int(u.RedirectTime)
	tu.LastRedirectTs = int(u.LastRedirectTs.Unix())
	tu.LastUpdateTs = time.Now()
	return
}

func FromTURLInfo(tu *model.TURLInfo) (u *data.URLInfo) {
	u = &data.URLInfo{}
	u.IncrID = tu.Id
	u.RawUrl = tu.RawUrl
	u.ShortUrl = tu.ShortUrl
	u.ShortCode = tu.ShortCode
	u.CreateAt = time.Unix(int64(tu.CreateTs), 0)
	u.URLType = uint64(tu.UrlType)
	u.Status = uint64(tu.Status)
	u.BanAt = time.Unix(int64(tu.BanAt), 0)
	u.UnBanAt = time.Unix(int64(tu.UnbanAt), 0)
	u.BanCuz = tu.BanCuz
	u.BanBy = tu.BanBy
	u.RedirectTime = uint64(tu.RedirectTime)
	u.LastRedirectTs = time.Unix(int64(tu.LastRedirectTs), 0)
	return
}

//UrlInfoMgr 实例义单例
func UrlInfoMgr() *urlInfoMgrType {
	urlInfoMgrOnce.Do(func() {
		urlInfoMgrInstance = &urlInfoMgrType{}
		urlInfoMgrInstance.Init()
	})
	return urlInfoMgrInstance
}

//urlInfoMgrType 实例定义
type urlInfoMgrType struct {
	ctx    context.Context
	cancel context.CancelFunc
}

//Init init
func (obj *urlInfoMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	obj.InitPubSub(obj.ctx)
	go obj.grLoop(obj.ctx)
}

//InitPubSub 订阅消息通知
func (obj *urlInfoMgrType) InitPubSub(ctx context.Context) {
	psKey := viper.GetString("url.pubsubKey")
	onError := func(err error, channels ...string) {
		if err != nil {
			//网络出错之类的，
			obj.InitPubSub(ctx)
		}
	}
	go db.RedisSub(ctx, obj.PubSubMsgBack, onError, psKey)
}

//grLoop loop
func (obj *urlInfoMgrType) grLoop(ctx context.Context) {
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

//Add 落地&缓存
func (obj *urlInfoMgrType) Add(incrID uint64, rawUrl, shortUrl, shortCode string) (err error) {
	u := data.URLInfoPool.Get().(*data.URLInfo)
	defer func() {
		data.URLInfoPool.Put(u)
	}()

	u.IncrID = incrID
	u.ShortCode = shortCode
	u.RawUrl = rawUrl
	u.ShortUrl = shortUrl
	u.CreateAt = time.Now()
	u.URLType = data.URLTypeCreateWithID
	u.Status = data.URLStatusEnable
	u.RedirectTime = 0
	u.LastRedirectTs = time.Now()

	err = obj.AddDB(u)
	if err != nil {
		return
	}
	_ = obj.AddRedis(u) //此处不成功也木关系
	_ = obj.AddLocalCache(u)
	return
}

//AddPhrase 落地&缓存,短语类型的短码
func (obj *urlInfoMgrType) AddPhrase(incrID uint64, rawUrl, shortUrl, shortCode string) (err error) {
	u := data.URLInfoPool.Get().(*data.URLInfo)
	defer func() {
		data.URLInfoPool.Put(u)
	}()

	u.IncrID = incrID
	u.ShortCode = shortCode
	u.RawUrl = rawUrl
	u.ShortUrl = shortUrl
	u.CreateAt = time.Now()
	u.URLType = data.URLTypeCreateWithPhrase
	u.Status = data.URLStatusEnable
	u.RedirectTime = 0
	u.LastRedirectTs = time.Now()

	err = obj.AddDB(u)
	if err != nil {
		return
	}
	_ = obj.AddRedis(u)      //redis层不成功也木关系
	_ = obj.AddLocalCache(u) //本地缓存层不成功也没有关系，家常便饭
	PhraseMgr().OnPhraseUsed(shortCode)
	return
}

//AddDB 三层缓存之db层
func (obj *urlInfoMgrType) AddDB(urlInfo *data.URLInfo) (err error) {
	m := ToTURLInfo(urlInfo)
	err = model.URLMgr().Save(m)
	return
}

//AddRedis 三层缓存之redis层
func (obj *urlInfoMgrType) AddRedis(urlInfo *data.URLInfo) (err error) {
	itemPrefix := viper.GetString("url.itemPrefix")
	itemExpireInRedis := viper.GetInt("url.itemExpireInRedis")
	urlInfoStr, err := json.Marshal(urlInfo)
	if err != nil {
		return
	}
	err = db.RedisSetEx(itemPrefix+urlInfo.ShortCode, itemExpireInRedis, urlInfoStr)
	return
}

//AddLocalCache 三层缓存之本地缓存
func (obj *urlInfoMgrType) AddLocalCache(urlInfo *data.URLInfo) (err error) {
	LocalCacheMgr().Add(urlInfo)
	return
}

//ReloadFromDB 从db开始，重建缓存
func (obj *urlInfoMgrType) ReloadFromDB(incrID uint64) (err error) {
	//是否是命中缓存穿透的
	turlInfo, err := model.URLMgr().Get(incrID)
	if err != nil {
		return
	}
	if turlInfo == nil {
		return
	}

	urlInfo := &data.URLInfo{}
	urlInfo = FromTURLInfo(turlInfo)

	obj.AddRedis(urlInfo)
	obj.AddLocalCache(urlInfo)

	return
}

//ReloadFromRedis 从redis开始，重建缓存
func (obj *urlInfoMgrType) ReloadFromRedis(shortCode string) (err error) {
	itemPrefix := viper.GetString("url.itemPrefix")
	itemKey := itemPrefix + shortCode
	exist, err := db.RedisExist(itemKey)
	if err != nil {
		return
	}
	if !exist {
		return
	}
	urlInfoStr := db.RedisGet(itemKey)
	urlInfo := &data.URLInfo{}
	err = json.Unmarshal([]byte(urlInfoStr), urlInfo)
	if err != nil {
		return
	}
	obj.AddLocalCache(urlInfo)
	return
}

//Get 通过短码获取原始url信息（会被加入到重定向统计中）
func (obj *urlInfoMgrType) Get(shortCode string) (urlInfo *data.URLInfo, err error) {
	//--------三层缓存之从本地缓存获取------------------
	//本地缓存存在，直接返回
	urlInfo, exist := LocalCacheMgr().Get(shortCode)
	if exist {
		LocalCacheMgr().UpdateRedirect(shortCode)
		return
	}

	//--------三层缓存之从redis缓存获取------------------
	//本地缓存不存在，去redis加载
	err = obj.ReloadFromRedis(shortCode)
	if err != nil {
		return
	}
	//再次去本地缓存获取
	urlInfo, exist = LocalCacheMgr().Get(shortCode)
	if exist {
		LocalCacheMgr().UpdateRedirect(shortCode)
		return
	}

	//--------三层缓存之从db获取------------------
	//本地缓存不存在，reload
	incrID, err := IDMgr().ToIncrID(shortCode)
	if err != nil {
		return
	}
	err = obj.ReloadFromDB(incrID)
	if err != nil {
		return
	}
	//再次去本地缓存获取
	urlInfo, exist = LocalCacheMgr().Get(shortCode)
	if exist {
		LocalCacheMgr().UpdateRedirect(shortCode)
		return
	}
	//--------三层缓存之数据不存在，做不存在标识（只做在本地缓存层）------------------
	//不存在，证明这是一个不存在的shortcode，新建状态为不存在，且加入到本地缓存中
	urlInfo = &data.URLInfo{}
	urlInfo.ShortCode = shortCode
	urlInfo.Status = data.URLStatusNotExist
	obj.AddLocalCache(urlInfo)

	return
}

//OnChange urlinfo 有变更，此时需要发布消息通知所有服务实例，通过redis的发布功能实现
func (obj *urlInfoMgrType) Ban(shortCode string) (err error) {
	logger.MainLogger.Debug("OnChange ,shortCode=" + shortCode)
	psKey := viper.GetString("url.pubsubKey")
	err = db.RedisPub(psKey, shortCode)
	//是否有这个url
	urlInfo, err := obj.Get(shortCode)
	if err != nil {
		return
	}
	if urlInfo.Status == data.URLStatusBan {
		err = fmt.Errorf("shortCode=" + shortCode + " ban aready")
		return
	}
	//更新db
	incrID, err := IDMgr().ToIncrID(shortCode)
	if err != nil {
		return
	}
	err = model.URLMgr().UpdateStatus(incrID, data.URLStatusBan)
	//删除redis
	itemPrefix := viper.GetString("url.itemPrefix")
	err = db.RedisDel(itemPrefix + shortCode)
	if err != nil {
		return
	}
	//通知每个服务进程，删除本地缓存
	err = obj.OnChange(shortCode)
	return
}

//OnChange urlinfo 有变更，此时需要发布消息通知所有服务实例，通过redis的发布功能实现
func (obj *urlInfoMgrType) OnChange(shortCode string) (err error) {
	logger.MainLogger.Debug("OnChange ,shortCode=" + shortCode)
	psKey := viper.GetString("url.pubsubKey")
	err = db.RedisPub(psKey, shortCode)
	return
}

//PubSubMsgBack 监控到通道消息，回调之
func (obj *urlInfoMgrType) PubSubMsgBack(channel string, byteData []byte) {
	shortCode := string(byteData)
	logger.MainLogger.Debug("PubSubMsgBack ,shortCode=" + shortCode)
	obj.GotChangeNotify(shortCode)
}

//GotChangeNotify 收到通知，某个urlinfo信息变更了，此时需要删除本地缓存（不需要去db重新拉取），通过redis的订阅功能实现
func (obj *urlInfoMgrType) GotChangeNotify(shortCode string) (err error) {
	LocalCacheMgr().Del(shortCode)
	return
}
