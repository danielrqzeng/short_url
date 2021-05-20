package model

import (
	"fmt"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sort"
	"sync"
	"time"
)

//实例
var (
	KVModelMgrInstance *KVModelMgrType
	KVModelMgrOnce     sync.Once
)

//KVMgr 获取单例实例
func KVMgr() *KVModelMgrType {
	KVModelMgrOnce.Do(func() {
		KVModelMgrInstance = &KVModelMgrType{}
		KVModelMgrInstance.Init()
	})
	return KVModelMgrInstance
}

//KVModelMgrType 实例定义
type KVModelMgrType struct {
	sync.Mutex
	//双缓存策略，使用时候只使用读缓存，更新时候使用写缓存，并且交换读写缓存的指针
	//更新策略为5min(mysql.updateInterval中配置)更新一次数据
	readCache    map[string]map[string]*TKVInfo //[pk][sk]	= TKVInfo
	writeCache   map[string]map[string]*TKVInfo //[pk][sk]	= TKVInfo
	dataSign     string                         //数据签名
	lastDataSign string                         //前一个版本的数据签名
}

//PK primary key
func (mgr *KVModelMgrType) PK(m *TKVInfo) (pk string) {
	pk = m.KeyInfo
	return
}

//SK sub key
func (mgr *KVModelMgrType) SK(m *TKVInfo) (sk string) {
	sk = "sk"
	return
}

//Init init
func (mgr *KVModelMgrType) Init() {
	mgr.readCache = make(map[string]map[string]*TKVInfo)
	mgr.writeCache = make(map[string]map[string]*TKVInfo)

}

//Name table name
func (mgr *KVModelMgrType) Name() (name string) {
	m := &TKVInfo{}
	name = m.TableName()
	return
}

//BeforeLoad action before load
func (mgr *KVModelMgrType) BeforeLoad() {
}

//ResetCache reset write cache
func (mgr *KVModelMgrType) ResetCache() {
	mgr.writeCache = make(map[string]map[string]*TKVInfo)
}

//ReloadFromDB reload from db
func (mgr *KVModelMgrType) Reload() (err error) {
	return
}

//Swap swap read&Write
func (mgr *KVModelMgrType) Swap() {
	mgr.CalcDataSign()
	mgr.writeCache, mgr.readCache = mgr.readCache, mgr.writeCache
	mgr.ResetCache()
}

//AfterLoad action after reload
func (mgr *KVModelMgrType) AfterLoad() {
}

//AddToCache 将db数据更新到writeCache中
func (mgr *KVModelMgrType) AddToCache(m TKVInfo) {
	pk := mgr.PK(&m)
	sk := mgr.SK(&m)
	if _, ok := mgr.writeCache[pk]; !ok {
		mgr.writeCache[pk] = make(map[string]*TKVInfo)
	}

	mgr.writeCache[pk][sk] = &m

}

//CalcDataSign 计算数据签名
func (mgr *KVModelMgrType) CalcDataSign() {
	mgr.lastDataSign = mgr.dataSign
	datas := make([]string, 0)
	for _, infos := range mgr.writeCache {
		for _, info := range infos {
			tmp, err := json.Marshal(info)
			if err != nil {
				logger.MainLogger.Error(err.Error())
				continue
			}
			datas = append(datas, utils.Md5sum(tmp))
		}
	}
	mgr.dataSign = ""
	sort.Strings(datas)
	for _, d := range datas {
		mgr.dataSign = utils.Md5sum([]byte(mgr.dataSign + d))
	}
}

//GetDataSign 获取数据签名
func (mgr *KVModelMgrType) GetDataSign() string {
	return mgr.dataSign
}

//IsDataChange 是否数据有变动
func (mgr *KVModelMgrType) IsDataChange() bool {
	return mgr.lastDataSign != mgr.dataSign
}

//Get 获取某个key，如果不存在，kvInfo为nil
func (mgr *KVModelMgrType) Get(key string) (kvInfo *TKVInfo, err error) {
	var l []TKVInfo
	dba := DB().Table(&l).
		Where("key_info", key)
	err = DoQuery(dba)
	if err != nil {
		return
	}

	if len(l) <= 0 {
		kvInfo = nil
		return
	}
	kvInfo = &l[0]
	return
}

//Get 获取某个key，如果不存在，kvInfo为nil
func (mgr *KVModelMgrType) Add(kvInfo *TKVInfo) (err error) {
	_, err = DoInsert(kvInfo)
	if err != nil {
		return
	}
	return
}

//Get 获取某个key，如果不存在，kvInfo为nil
func (mgr *KVModelMgrType) Incr(key string, incrby uint64) (err error) {
	begin := time.Now()
	dba := DB()
	affectRow, err := dba.Execute("update t_kv_info set `val_int_info`=`val_int_info`+? where `key_info`=?", incrby, key)
	lastSql := dba.LastSql()
	if err != nil {
		OnSQLRun(lastSql, begin, err.Error(), "", int(affectRow))
		return
	}
	if affectRow != 1 {
		err = fmt.Errorf("affectRow!=1")
		OnSQLRun(lastSql, begin, err.Error(), "", int(affectRow))
	}
	OnSQLRun(lastSql, begin, SQLSuccessMsg, "", int(affectRow))
	return
}
