package service

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"sync"
	"time"
)

//内部变量的定义
var (
	phraseEncodeMgrInstance *phraseEncodeMgrType
	phraseEncodeMgrOnce     sync.Once
)

//PhraseEncodeMgr 短语编码单例
func PhraseEncodeMgr() *phraseEncodeMgrType {
	phraseEncodeMgrOnce.Do(func() {
		phraseEncodeMgrInstance = &phraseEncodeMgrType{}
		phraseEncodeMgrInstance.Init()
	})
	return phraseEncodeMgrInstance
}

//phraseEncodeMgrType 短语编码实例定义，通过短语来生成短码
type phraseEncodeMgrType struct {
	ctx    context.Context
	cancel context.CancelFunc
}

//Init init
func (obj *phraseEncodeMgrType) Init() {
	obj.ctx, obj.cancel = context.WithCancel(context.Background())
	go obj.grLoop(obj.ctx)
}

//grLoop loop
func (obj *phraseEncodeMgrType) grLoop(ctx context.Context) {
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

//Encode 将长网址编码为短网址
func (obj *phraseEncodeMgrType) Encode(phrase, rawUrl string) (shortUrl string, err error) {
	//是否网址是合法的
	valid, err := LinkMgr().CheckLink(rawUrl)
	if err != nil {
		return
	}
	if !valid {
		err = fmt.Errorf("url=" + rawUrl + " not a good website")
		return
	}

	//检测短语是否合法
	//短语合法之-长度
	phraseMinLen := viper.GetInt("phraseID.minLen")
	phraseMaxLen := viper.GetInt("phraseID.maxLen")
	if len(phrase) < phraseMinLen {
		err = fmt.Errorf("param not valid,min=%d,got=%d", phraseMinLen, len(phrase))
		return
	}
	if len(phrase) > phraseMaxLen {
		err = fmt.Errorf("param not valid,max=%d,got=%d", phraseMaxLen, len(phrase))
		return
	}

	//短语合法之-是否禁用
	isForbid, err := PhraseMgr().IsForbid(phrase)
	if err != nil {
		return
	}
	if isForbid {
		err = fmt.Errorf("phrase '%s' been forbid", phrase)
		return
	}

	//短语合法之-已经被用掉了
	used, err := PhraseMgr().IsBeenUsed(phrase)
	if err != nil {
		return
	}
	if used {
		err = fmt.Errorf("phrase '%s' been used", phrase)
		return
	}

	incrID, err := IDMgr().ToIncrID(phrase)
	if err != nil {
		return
	}

	shortCode := phrase
	shortUrl = viper.GetString("url.domain") + shortCode
	err = UrlInfoMgr().AddPhrase(incrID, rawUrl, shortUrl, shortCode)
	if err != nil {
		return
	}
	return
}
