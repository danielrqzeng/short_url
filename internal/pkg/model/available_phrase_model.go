package model

import (
	"iyfiysi.com/short_url/internal/pkg/data"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sort"
	"sync"
)

//实例
var (
	AvailablePhraseModelMgrInstance *AvailablePhraseModelMgrType
	AvailablePhraseModelMgrOnce     sync.Once
)

//AvailablePhraseMgr 获取单例
func AvailablePhraseMgr() *AvailablePhraseModelMgrType {
	AvailablePhraseModelMgrOnce.Do(func() {
		AvailablePhraseModelMgrInstance = &AvailablePhraseModelMgrType{}
		AvailablePhraseModelMgrInstance.Init()
	})
	return AvailablePhraseModelMgrInstance
}

//AvailablePhraseModelMgrType 实例定义
type AvailablePhraseModelMgrType struct {
	sync.Mutex
	//双缓存策略，使用时候只使用读缓存，更新时候使用写缓存，并且交换读写缓存的指针
	//更新策略为5min(mysql.updateInterval中配置)更新一次数据
	readCache    map[string]map[string]*TAvailablePhraseInfo //[pk][sk]	= TAvailablePhraseInfo
	writeCache   map[string]map[string]*TAvailablePhraseInfo //[pk][sk]	= TAvailablePhraseInfo
	dataSign     string                                      //数据签名
	lastDataSign string                                      //前一个版本的数据签名
}

//PK primary key
func (mgr *AvailablePhraseModelMgrType) PK(m *TAvailablePhraseInfo) (pk string) {
	pk = m.Phrase
	return
}

//SK sub key
func (mgr *AvailablePhraseModelMgrType) SK(m *TAvailablePhraseInfo) (sk string) {
	sk = "sk"
	return
}

//Init init
func (mgr *AvailablePhraseModelMgrType) Init() {
	mgr.readCache = make(map[string]map[string]*TAvailablePhraseInfo)
	mgr.writeCache = make(map[string]map[string]*TAvailablePhraseInfo)

}

//Name table name
func (mgr *AvailablePhraseModelMgrType) Name() (name string) {
	m := &TAvailablePhraseInfo{}
	name = m.TableName()
	return
}

//BeforeLoad action before load
func (mgr *AvailablePhraseModelMgrType) BeforeLoad() {
}

//ResetCache reset write cache
func (mgr *AvailablePhraseModelMgrType) ResetCache() {
	mgr.writeCache = make(map[string]map[string]*TAvailablePhraseInfo)
}

//ReloadFromDB reload from db
func (mgr *AvailablePhraseModelMgrType) Reload() (err error) {
	var l []TAvailablePhraseInfo
	dba := DB().Table(&l)
	err = DoQuery(dba)
	if err != nil {
		return
	}
	for _, m := range l {
		mgr.AddToCache(m)
	}

	return
}

//Swap swap read&Write
func (mgr *AvailablePhraseModelMgrType) Swap() {
	mgr.CalcDataSign()
	mgr.writeCache, mgr.readCache = mgr.readCache, mgr.writeCache
	mgr.ResetCache()
}

//AfterLoad action after reload
func (mgr *AvailablePhraseModelMgrType) AfterLoad() {
}

//AddToCache 将db数据更新到writeCache中
func (mgr *AvailablePhraseModelMgrType) AddToCache(m TAvailablePhraseInfo) {
	pk := mgr.PK(&m)
	sk := mgr.SK(&m)
	if _, ok := mgr.writeCache[pk]; !ok {
		mgr.writeCache[pk] = make(map[string]*TAvailablePhraseInfo)
	}

	mgr.writeCache[pk][sk] = &m

}

//CalcDataSign 计算数据签名
func (mgr *AvailablePhraseModelMgrType) CalcDataSign() {
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
func (mgr *AvailablePhraseModelMgrType) GetDataSign() string {
	return mgr.dataSign
}

//IsDataChange 是否数据有变动
func (mgr *AvailablePhraseModelMgrType) IsDataChange() bool {
	return mgr.lastDataSign != mgr.dataSign
}

//GetRegexList 获取为正则表达式的短语
func (mgr *AvailablePhraseModelMgrType) GetRegexList() (regexList []string) {
	regexList = make([]string, 0)
	for _, infos := range mgr.readCache {
		for _, info := range infos {
			if info.PhraseType == data.PhraseTypeRegexp {
				regexList = append(regexList, info.Phrase)
			}
		}
	}
	return
}

//GetPhraseList 获取为文字类型（字母）的短语
func (mgr *AvailablePhraseModelMgrType) GetPhraseList() (regexList []string) {
	regexList = make([]string, 0)
	for _, infos := range mgr.readCache {
		for _, info := range infos {
			if info.PhraseType == data.PhraseTypeWord {
				regexList = append(regexList, info.Phrase)
			}
		}
	}
	return
}
