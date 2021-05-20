package model

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"sync"
)

//IModelMgr 由此文件统一管理数据更新加载等一系列操作
//各个mgr实现具体的逻辑和各自的数据管理
type IModelMgr interface {
	Name() string        //此model的唯一标识
	Init()               //model初始化
	BeforeLoad()         //去db拉取数据之前回调
	ResetCache()         //重置writeCache
	Reload() (err error) //去db拉数据
	Swap()               //调换数据，readCache<->writeCache
	AfterLoad()          //所有model拉完数据之后回调告知该model
}

//内部变量
var modelMgrOnce sync.Once
var modelMgrInstance *ModelMgrType

//MaxModelCallbackfuncPriority 最大回调次数
const MaxModelCallbackfuncPriority = 10

//ModelMgr orm管理实例
func ModelMgr() *ModelMgrType {
	modelMgrOnce.Do(func() {
		modelMgrInstance = &ModelMgrType{}
		modelMgrInstance.Init()
	})
	return modelMgrInstance
}

//ModelMgrType 实例定义
type ModelMgrType struct {
	//第一次加载完毕，做这个是为了使得对那些有回调需求，但是回调又在第一次加载完毕之后再RegisterCallbackFunc的业务做的回调
	firstLoadDone bool
	models        map[string]IModelMgr
	cbOnLoadDone  [][]func() //{[priority]={cb1,cb2,cb3....}},其中有10个priority=0...9,数值越小，优先级越高
}

//Init init
func (modelMgr *ModelMgrType) Init() {
	modelMgr.firstLoadDone = false
	modelMgr.models = make(map[string]IModelMgr)
	modelMgr.cbOnLoadDone = make([][]func(), MaxModelCallbackfuncPriority)
	for idx := range modelMgr.cbOnLoadDone {
		modelMgr.cbOnLoadDone[idx] = make([]func(), 0)
	}
}

//RegisterCallbackFunc 注册orm
// priority越低，优先级越高，取值范围为0~9
func (modelMgr *ModelMgrType) RegisterCallbackFunc(priority int, f func()) {
	if priority < 0 || priority > MaxModelCallbackfuncPriority {
		return
	}
	modelMgr.cbOnLoadDone[priority] = append(modelMgr.cbOnLoadDone[priority], f)
	if modelMgr.firstLoadDone {
		f()
	}
}

//Register 注册
func (modelMgr *ModelMgrType) Register(model IModelMgr) {
	name := model.Name()
	modelMgr.models[name] = model
}

//ModelInit 首次初始化
func (modelMgr *ModelMgrType) ModelInit() {
	for _, model := range modelMgr.models {
		model.Init()
	}
	modelMgr.BeforeLoad()
	modelMgr.ResetCache()
	modelMgr.Reload()
	modelMgr.Swap()
	modelMgr.AfterLoad()
}

//BeforeLoad action before load
func (modelMgr *ModelMgrType) BeforeLoad() {
	for _, model := range modelMgr.models {
		model.BeforeLoad()
	}
}

//ResetCache reset cache
func (modelMgr *ModelMgrType) ResetCache() {
	for _, model := range modelMgr.models {
		model.ResetCache()
	}
}

//ReloadFromDB reload
func (modelMgr *ModelMgrType) Reload() {
	for _, model := range modelMgr.models {
		err := model.Reload()
		if err != nil {
		}
	}
}

//Swap swap
func (modelMgr *ModelMgrType) Swap() {
	for _, model := range modelMgr.models {
		model.Swap()
	}
}

//AfterLoad action after load
func (modelMgr *ModelMgrType) AfterLoad() {
	modelMgr.firstLoadDone = true
	for _, model := range modelMgr.models {
		model.AfterLoad()
	}

	//全部数据加载完毕，通知需要通知的callback
	for _, cbs := range modelMgr.cbOnLoadDone {
		for _, cb := range cbs {
			cb()
		}
	}
}

func (modelMgr *ModelMgrType) grLoop() {
	c := cron.New()

	reloadCron := viper.GetString("mysql.reloadCron")

	//_, err := c.AddFunc("*/5 * * * *", func() {
	_, err := c.AddFunc(reloadCron, func() {
		modelMgr.BeforeLoad()
		modelMgr.ResetCache()
		modelMgr.Reload()
		modelMgr.Swap()
		modelMgr.AfterLoad()
		//DumpCPU()
		//DumpHeap()
		//DumpMemStat()
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}
