package service

import (
	"context"
	"github.com/adamzy/cedar-go"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/pkg/data"
	"iyfiysi.com/short_url/internal/pkg/db"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/model"
	"regexp"
	"strings"
	"sync"
	"time"
)

//内部变量的定义
var (
	phraseMgrInstance *phraseMgrType
	phraseMgrOnce     sync.Once
)

//PhraseMgr 短语单例
func PhraseMgr() *phraseMgrType {
	phraseMgrOnce.Do(func() {
		phraseMgrInstance = &phraseMgrType{}
		phraseMgrInstance.Init()
	})
	return phraseMgrInstance
}

/*
phraseMgrType 短语单例定义
短语库三个
	* 禁用短语库：在生成短语时候调用，id生成或者短语生成方式都需要调用，以免使用了涉敏的短语，不如短语fuck明显不能用
	* 可用短语库：在id生成时候调用，主要是使得id生成时候，余空出来这些短语给短语生成的方式使用，主要使用的牛津英文词典和新华字典生成的库，保存在t_available_phrase_info表中
	* 已用短语库：对已用的短语（短语生成方式生成的），将会在redis建立一个集合，以判断那些已经被用了，此主要是为了提高解码时候的合法性判断性能
*/
type phraseMgrType struct {
	sync.RWMutex
	ready                   bool
	availablePhraseReg      *regexp.Regexp
	availablePhraseWordTrie *cedar.Cedar
	forbidPhraseReg         *regexp.Regexp
	forbidPhraseWordTrie    *cedar.Cedar
	ctx                     context.Context
	cancel                  context.CancelFunc
}

//Init init
func (obj *phraseMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	obj.ready = false
	go obj.grLoop(obj.ctx)
	go model.ModelMgr().RegisterCallbackFunc(9, OnLoadDataDone) //此处使用goroutine去做下注册，否则会导致死锁
}

//grLoop loop
func (obj *phraseMgrType) grLoop(ctx context.Context) {
	secTick := time.NewTicker(time.Second)
	defer secTick.Stop()
	done := false
	for !done {
		select {
		case <-secTick.C:
			{
				//每秒回调
			}
		case <-ctx.Done():
			{
				done = true
			}
		}
	}
}

//Ready 短语库是否已经可用（其需要去db拉取数据，构建完毕之后才算是可用）
func (obj *phraseMgrType) Ready() bool {
	return obj.ready
}

//ReBuild 短语库有变动，需要重建
//这个重建比较耗时，一般看数据量，默认库(6w的available和2k的forbid）里面数据量，大约需要30s才能重建完毕
func (obj *phraseMgrType) ReBuild() {
	//重建待用短语库
	regList := model.AvailablePhraseMgr().GetRegexList()
	wordListStr := strings.Join(regList, "|")
	logger.MainLogger.Debug("wordListStr=" + wordListStr)
	availableReg := regexp.MustCompile(wordListStr) //正则匹配

	phraseList := model.AvailablePhraseMgr().GetPhraseList()
	availableTrie := cedar.New() //精准匹配
	for k, v := range phraseList {
		err := availableTrie.Insert([]byte(v), k)
		logger.MainLogger.Error("ReBuild err=" + err.Error())
		return
	}

	//重建禁用短语库
	regList = model.ForbidPhraseInfoMgr().GetRegexList()
	wordListStr = strings.Join(regList, "|")
	logger.MainLogger.Debug("wordListStr=" + wordListStr)
	forbidReg := regexp.MustCompile(wordListStr) //正则匹配

	phraseList = model.ForbidPhraseInfoMgr().GetPhraseList()
	forbidTrie := cedar.New() //精准匹配
	for k, v := range phraseList {
		err := forbidTrie.Insert([]byte(v), k)
		logger.MainLogger.Error("ReBuild err=" + err.Error())
		return
	}

	obj.Lock()
	defer obj.Unlock()
	//使用新值
	obj.availablePhraseReg = availableReg
	obj.availablePhraseWordTrie = availableTrie
	obj.forbidPhraseReg = forbidReg
	obj.forbidPhraseWordTrie = forbidTrie
	obj.ready = true
	logger.MainLogger.Debug("Rebuild phrase success")
	return
}

//IsForbid 是否短语被禁用
func (obj *phraseMgrType) IsForbid(phrase string) (isForbid bool, err error) {
	obj.RLock()
	defer obj.RUnlock()
	isForbid = false

	//精准匹配
	t := obj.forbidPhraseWordTrie
	if _, existErr := t.Get([]byte(phrase)); existErr == nil {
		logger.MainLogger.Debug("mark phrase=" + phrase)
		isForbid = true
		return
	}
	//正则匹配-匹配成功
	r := obj.forbidPhraseReg
	//logger.MainLogger.Debug(r.String())
	matchStrs := r.FindAllString(phrase, -1)
	if len(matchStrs) != 0 {
		for _, v := range matchStrs {
			logger.MainLogger.Debug("#" + v + "#")
		}
		logger.MainLogger.Debug("mark phrase=" + phrase + ",match=" + strings.Join(matchStrs, "-"))
		isForbid = true
		return
	}
	return
}

//IsBeenUsedByID 是否短语已经被用掉了（无论是否是id生成还是短语生成的），其代价比较大，有可能会去db查找
func (obj *phraseMgrType) IsBeenUsed(phrase string) (beenUsed bool, err error) {
	beenUsed = false
	urlInfo, err := UrlInfoMgr().Get(phrase)
	if err != nil {
		return
	}
	if urlInfo.Status == data.URLStatusNotExist {
		return
	}
	//已经被用掉了
	beenUsed = true
	return
}

//IsBeenUsedByPhrase 是否短语已经被短语生成的方式用掉了（只检查被短语生成的方式用掉），代价比较小，只在redis查找，后续优化可以直接放内存
func (obj *phraseMgrType) IsBeenUsedByPhrase(phrase string) (beenUsed bool, err error) {
	beenUsed = false
	phraseSetKey := viper.GetString("phraseID.phraseSetKey")
	beenUsed, err = db.RedisSISMEMBER(phraseSetKey, phrase)
	return
}

//OnPhraseUsed 通知短语被用了
func (obj *phraseMgrType) OnPhraseUsed(phrase string) {
	phraseSetKey := viper.GetString("phraseID.phraseSetKey")
	err := db.RedisSAdd(phraseSetKey, phrase)
	if err != nil {
		logger.MainLogger.Error(err.Error())
		//TODO 这里失败了，会有些问题，告警处理
	}
	return
}

//IsAvailable 是否短语是待用,这里只代表其在待用短语库中
func (obj *phraseMgrType) IsAvailable(phrase string) (isAvailable bool, err error) {
	obj.RLock()
	defer obj.RUnlock()
	isAvailable = false

	//精准匹配
	t := obj.availablePhraseWordTrie
	if _, existErr := t.Get([]byte(phrase)); existErr == nil {
		logger.MainLogger.Debug("match " + phrase)
		isAvailable = true
		return
	}
	//正则匹配-匹配成功
	r := obj.availablePhraseReg
	matchStrs := r.FindAllString(phrase, -1)
	if len(matchStrs) != 0 {
		isAvailable = true
		logger.MainLogger.Debug("matchStrs=" + strings.Join(matchStrs, "-"))
		return
	}
	return
}

//OnLoadDataDone db数据加载完毕的回调，db每5min重新加载一次数据，之后做回调
func OnLoadDataDone() {
	change := false
	if !change {
		change = model.AvailablePhraseMgr().IsDataChange()
	}
	if !change {
		change = model.ForbidPhraseInfoMgr().IsDataChange()
	}
	//数据没有变动过，不需要更改
	if !change {
		return
	}
	logger.MainLogger.Debug("OnLoadDataDone")

	//数据变动了，重建
	PhraseMgr().ReBuild()
}
