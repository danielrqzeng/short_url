package service

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/pkg/db"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/model"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sync"
	"time"
)

//内部变量的定义
var (
	idMgrInstance *idMgrType
	idMgrOnce     sync.Once
)

//idMgr 实例义单例
func IDMgr() *idMgrType {
	idMgrOnce.Do(func() {
		idMgrInstance = &idMgrType{}
		idMgrInstance.Init()
	})
	return idMgrInstance
}

/*idMgrType 自增id管理实例
对于id生成的短码
* 先建立id池
	* id池在没有id时候会调用主动触发
	* id池也会被协程时时扫描，少于阀值即触发建立id池
	* id池建立时候，一般会建立n个，然后缓存在本地（用光没用光无所谓，因为我们并不保证id是连续的）
	* id池建立时候，会被检测是否禁用，是否要保留给短语使用等步骤（比如新增的10个id，有2个是被禁用的，3个是待用短语库的，那么只剩下5个可用）
* id生成
	* 在使用id生成短码时候，会调用以使用id关联短码
	* id使用了之后，会落地db关联短码
	* id是从已有的id池里面直接取出来使用
*/
type idMgrType struct {
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	idMaking bool     //是否正在生成id
	idPool   []uint64 //缓存的id池
	idMax    uint64
}

//Init init
func (obj *idMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	obj.idMax = 0
	obj.idMaking = false
	go obj.grLoop(obj.ctx)
}

//grLoop loop
func (obj *idMgrType) grLoop(ctx context.Context) {
	secTick := time.NewTicker(time.Second)
	defer secTick.Stop()
	done := false
	for !done {
		select {
		case <-secTick.C:
			{
				//做任务的每秒回调

				//更新最大id
				incIDKey := viper.GetString("incID.incIDKey")
				currID, err := db.RedisGetInt(incIDKey)
				if err == nil {
					obj.idMax = currID
				}

				//检查id pool是否到库存告急，需要重新装库
				reloadWhen := viper.GetInt("incID.reloadWhen")
				if len(obj.idPool) < reloadWhen {
					obj.GenIDPool()
				}
			}
		case <-ctx.Done():
			{
				done = true
			}
		}
	}
}

func (obj *idMgrType) ToIncrID(shortCode string) (incrID uint64, err error) {
	showID, err := utils.Base62Decode(shortCode)
	if err != nil {
		return
	}

	incrID, err = NumShuffleMgr().Decode(showID)
	if err != nil {
		return
	}
	return
}
func (obj *idMgrType) ToShortCode(incrID uint64) (shortCode string, err error) {
	showID, err := NumShuffleMgr().Encode(incrID)
	if err != nil {
		return
	}
	shortCode = utils.Base62Encode(showID)
	return
}

func (obj *idMgrType) IsIDPoolEmpty() bool {
	if len(obj.idPool) == 0 {
		return true
	}
	return false
}

func (obj *idMgrType) PushToIDPool(ids ...uint64) {
	obj.Lock()
	defer obj.Unlock()
	obj.idPool = append(obj.idPool, ids...)
	logger.MainLogger.Debug("PushToIDPool num=" + utils.Num2Str(len(ids)))
}

func (obj *idMgrType) PopFromIDPool() (id uint64, err error) {
	obj.Lock()
	defer obj.Unlock()
	if len(obj.idPool) <= 0 {
		err = fmt.Errorf("ip pool is empty")
		return
	}
	id = obj.idPool[0]
	obj.idPool = obj.idPool[1:]
	return
}

//GenIDPool 生成id库存
// 	1. 当没有id使用的时候，会去做生成
// 	2. 会有一个gorounie检查，若是低于生成量的1/4时候，也会去生成
func (obj *idMgrType) GenIDPool() {
	//logger.MainLogger.Debug("GenIDPool")
	if !PhraseMgr().Ready() {
		return
	}
	logger.MainLogger.Debug("GenIDPool")
	//正在生成id中，只能等待
	if obj.idMaking {
		return
	}
	logger.MainLogger.Debug("GenIDPool start")
	obj.idMaking = true
	defer func() { obj.idMaking = false }()

	//分布式锁
	incIDLockKey := viper.GetString("incID.incIDLockKey")
	incIDLockMS := viper.GetInt("incID.incIDLockMS")
	lock, err := db.RedLock(incIDLockKey, incIDLockMS)
	if err != nil {
		logger.MainLogger.Debug("GenIDPool,err" + err.Error())
		logger.MainLogger.Error("fail to RedLock")
		return
	}
	logger.MainLogger.Debug("GenIDPool")
	err = lock.Lock()
	if err != nil {
		logger.MainLogger.Error("fail to RedLock")
		return
	}
	logger.MainLogger.Debug("GenIDPool")

	//从db中，增加id
	incIDKey := viper.GetString("incID.incIDKey")
	incrBy := viper.GetUint64("incID.incBy")
	err = model.KVMgr().Incr(incIDKey, incrBy)
	if err != nil {
		lock.Unlock()
		return
	}

	kvInfo, err := model.KVMgr().Get(incIDKey)
	if err != nil {
		lock.Unlock()
		return
	}
	//主动释放分布式锁，因为下面检测id是否可用的比较耗时，得先放掉分布式锁以便其他的进程可以用
	lock.Unlock()

	//更新id信息
	idTo := kvInfo.ValIntInfo
	idFrom := idTo - incrBy
	obj.idMax = idTo
	logger.MainLogger.Debug("id from " + utils.Num2Str(idFrom) + " to " + utils.Num2Str(idTo))

	idPool := make([]uint64, 0)
	for id := idFrom; id < idTo; id++ {
		shortCode, err := obj.ToShortCode(id)
		if err != nil {
			continue
		}

		s := time.Now()
		//禁用的短码
		isforbid, err := PhraseMgr().IsForbid(shortCode)
		if err != nil {
			continue
		}
		if isforbid {
			logger.MainLogger.Debug("id=" + utils.Num2Str(id) + ",shortCode=" + shortCode + " is forbid")
			continue
		}
		logger.MainLogger.Debug(shortCode + " check forbid elapsed=" + time.Now().Sub(s).String())

		//待用短语库的短码（需要保留给短语生成接口）
		s = time.Now()
		isAvailable, err := PhraseMgr().IsAvailable(shortCode)
		if err != nil {
			continue
		}
		if isAvailable {
			logger.MainLogger.Debug("id=" + utils.Num2Str(id) + ",shortCode=" + shortCode + " not available for id gen")
			continue
		}
		logger.MainLogger.Debug(shortCode + " check available elapsed=" + time.Now().Sub(s).String())

		//已经被用了的短码
		isUsed, err := PhraseMgr().IsBeenUsedByPhrase(shortCode)
		if err != nil {
			continue
		}
		if isUsed {
			logger.MainLogger.Debug("id=" + utils.Num2Str(id) + ",shortCode=" + shortCode + " been used")
			continue
		}
		logger.MainLogger.Debug("id=" + utils.Num2Str(id) + ",shortCode=" + shortCode + " succ")
		idPool = append(idPool, id)
	}
	logger.MainLogger.Debug("Done")

	obj.PushToIDPool(idPool...)
	//同步id到redis
	err = db.RedisSet(incIDKey, idTo)
	if err != nil {
		return
	}
	logger.MainLogger.Debug("GenIDPool Done")
}

//IDInit 初始化id，会检查db是否有，没有则初始化db值，否则跳过
func (obj *idMgrType) IDInit() (err error) {
	//先获取锁
	incIDLockKey := viper.GetString("incID.incIDLockKey")
	incIDLockMS := viper.GetInt("incID.incIDLockMS")
	lock, err := db.RedLock(incIDLockKey, incIDLockMS)
	if err != nil {
		return
	}
	err = lock.Lock()
	if err != nil {
		return
	}
	defer lock.Unlock()

	//检查db是否存在了id
	incIDKey := viper.GetString("incID.incIDKey")
	kvInfo, err := model.KVMgr().Get(incIDKey)
	if err != nil {
		return
	}
	//incID在db还没有存在，需要做初始化
	if kvInfo == nil {
		kvInfo = &model.TKVInfo{}
		kvInfo.KeyInfo = incIDKey
		kvInfo.ValIntInfo = viper.GetUint64("incID.incIDStartAt")
		kvInfo.DescInfo = "inc ID"
		err = model.KVMgr().Add(kvInfo)
		if err != nil {
			return
		}
	}
	return
}

//IsIDValid 是否id是合法的，判断id范围
func (obj *idMgrType) IsIDValid(id uint64) (valid bool, err error) {
	if obj.idMax == 0 {
		err = fmt.Errorf("max id not available,pls wait")
		return
	}
	valid = false
	minID := viper.GetUint64("incID.incIDStartAt")
	maxID := obj.idMax
	if id >= minID && id <= maxID {
		valid = true
	}
	return
}

//MakeID 制造一个id
func (obj *idMgrType) MakeID() (id uint64, err error) {
	logger.MainLogger.Debug("MadeID")
	if obj.IsIDPoolEmpty() {
		obj.GenIDPool()
	}
	logger.MainLogger.Debug("MadeID")

	if obj.IsIDPoolEmpty() {
		err = fmt.Errorf("make id fail,plase retry later")
		return
	}
	logger.MainLogger.Debug("MadeID")
	id, err = obj.PopFromIDPool()
	return
}
