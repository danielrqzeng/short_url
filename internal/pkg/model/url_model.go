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
	URLModelMgrInstance *URLModelMgrType
	URLModelMgrOnce     sync.Once
)

//ModMgr 获取单例实例
func URLMgr() *URLModelMgrType {
	URLModelMgrOnce.Do(func() {
		URLModelMgrInstance = &URLModelMgrType{}
		URLModelMgrInstance.Init()
	})
	return URLModelMgrInstance
}

//URLModelMgrType 实例定义
type URLModelMgrType struct {
	sync.Mutex
	//双缓存策略，使用时候只使用读缓存，更新时候使用写缓存，并且交换读写缓存的指针
	//更新策略为5min(mysql.updateInterval中配置)更新一次数据
	readCache    map[string]map[string]*TURLInfo //[pk][sk]	= TURLInfo
	writeCache   map[string]map[string]*TURLInfo //[pk][sk]	= TURLInfo
	dataSign     string                          //数据签名
	lastDataSign string                          //前一个版本的数据签名
}

//PK primary key
func (mgr *URLModelMgrType) PK(m *TURLInfo) (pk string) {
	pk = utils.Num2Str(m.Id)
	return
}

//SK sub key
func (mgr *URLModelMgrType) SK(m *TURLInfo) (sk string) {
	sk = m.ShortCode
	return
}

//Init init
func (mgr *URLModelMgrType) Init() {
	mgr.readCache = make(map[string]map[string]*TURLInfo)
	mgr.writeCache = make(map[string]map[string]*TURLInfo)

}

//Name table name
func (mgr *URLModelMgrType) Name() (name string) {
	m := &TURLInfo{}
	name = m.TableName()
	return
}

//BeforeLoad action before load
func (mgr *URLModelMgrType) BeforeLoad() {
}

//ResetCache reset write cache
func (mgr *URLModelMgrType) ResetCache() {
	mgr.writeCache = make(map[string]map[string]*TURLInfo)
}

//ReloadFromDB reload from db
func (mgr *URLModelMgrType) Reload() (err error) {
	return
}

//Swap swap read&Write
func (mgr *URLModelMgrType) Swap() {
	mgr.CalcDataSign()
	mgr.writeCache, mgr.readCache = mgr.readCache, mgr.writeCache
	mgr.ResetCache()
}

//AfterLoad action after reload
func (mgr *URLModelMgrType) AfterLoad() {
}

//AddToCache 将db数据更新到writeCache中
func (mgr *URLModelMgrType) AddToCache(m TURLInfo) {
	pk := mgr.PK(&m)
	sk := mgr.SK(&m)
	if _, ok := mgr.writeCache[pk]; !ok {
		mgr.writeCache[pk] = make(map[string]*TURLInfo)
	}

	mgr.writeCache[pk][sk] = &m

}

//CalcDataSign 计算数据签名
func (mgr *URLModelMgrType) CalcDataSign() {
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
func (mgr *URLModelMgrType) GetDataSign() string {
	return mgr.dataSign
}

//IsDataChange 是否数据有变动
func (mgr *URLModelMgrType) IsDataChange() bool {
	return mgr.lastDataSign != mgr.dataSign
}

//GetPages 获取page,limit&offset参数里面不做校验，调用方自个保证合法
func (mgr *URLModelMgrType) Get(incrID uint64) (m *TURLInfo, err error) {
	var l []TURLInfo
	dba := DB().Table(&l).
		Where("id", incrID)
	err = DoQuery(dba)
	if err != nil {
		return
	}
	if len(l) != 1 {
		m = nil
		return
	}
	m = &l[0]
	return
}

//GetPages 获取page,limit&offset参数里面不做校验，调用方自个保证合法
func (mgr *URLModelMgrType) Save(m *TURLInfo) (err error) {
	_, err = DoInsert(m)
	if err != nil {
		return
	}
	return
}

//IncrRedirectTime 更新重定向次数和时间
func (mgr *URLModelMgrType) IncrRedirectTime(incrID, incrTime, lastRedirectTime uint64) (err error) {
	begin := time.Now()
	dba := DB()
	affectRow, err := dba.Execute(
		"update t_url_info set `redirect_time`=`redirect_time`+?, `last_redirect_ts`=? where `id`=?",
		incrTime,
		lastRedirectTime,
		incrID)
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

//GetPhraseTypeData 获取短语类型的url数据
func (mgr *URLModelMgrType) GetPhraseTypeData() (ml []*TURLInfo, err error) {
	var l []TURLInfo
	dba := DB().Table(&l).Where("url_type", 2) //加载
	err = DoQuery(dba)
	if err != nil {
		return
	}
	ml = make([]*TURLInfo, 0)
	for _, v := range l {
		m := v
		ml = append(ml, &m)
	}
	return
}

//GetPhraseTypeData 获取短语类型的url数据
func (mgr *URLModelMgrType) UpdateStatus(incID uint64, status int) (err error) {
	m := &TURLInfo{}
	dba := DB().Table(m).
		Data(map[string]interface{}{"status": status}).
		Where("id", incID)

	affectRow, err := DoUpdate(dba)
	if err != nil {
		return
	}
	if affectRow == 0 {
		err = fmt.Errorf("UpdateStatus incID=" + utils.Num2Str(incID) + ",affectRow=0")
		return
	}
	return
}
