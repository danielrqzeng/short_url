package model

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" //导入mysql
	"github.com/gohouse/gorose/v2"
	"github.com/json-iterator/go"
	"github.com/spf13/viper"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var engin *gorose.Engin
var once sync.Once

//SQLSuccessMsg success desc
const SQLSuccessMsg = "success"

//InitDBInstance 初始化db
func InitDBInstance() (err error) {

	c := &gorose.Config{
		Driver: "mysql",
		//username:password@protocol(address)/dbname?param=value
		//root:root@tcp(localhost:3306)/test?charset=utf8&parseTime=true
		Dsn: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true",
			viper.GetString("mysql.user"),
			viper.GetString("mysql.pass"),
			viper.GetString("mysql.host"),
			viper.GetString("mysql.port"),
			viper.GetString("mysql.db"),
		),
		SetMaxOpenConns: 0,
		SetMaxIdleConns: 5,
	}

	once.Do(func() {
		var err error
		engin, err = gorose.Open(c)
		if err != nil {
			panic(err)
		}
		_ = engin
		engin.GetPrefix()
	})
	return
}

//Init ...
func Init() (err error) {
	err = InitDBInstance()
	if err != nil {
		return
	}
	//model注册
	ModelMgr().Register(URLMgr())
	ModelMgr().Register(KVMgr())
	ModelMgr().Register(AvailablePhraseMgr())
	ModelMgr().Register(ForbidPhraseInfoMgr())

	//model管理类初始化&gr
	ModelMgr().ModelInit()
	go ModelMgr().grLoop()
	return
}

//DB get orm instance
func DB() gorose.IOrm {
	engin.SetPrefix("")
	return engin.NewOrm()
}

//DBWithTablePrefix set table prefix
func DBWithTablePrefix(prefix string) gorose.IOrm {
	engin.SetPrefix(prefix)
	return engin.NewOrm()
}

//OnSQLRun db日志
//result=success 代表成功，其他都是失败
func OnSQLRun(sql string, begin time.Time, result, retResultMsg string, selectNum int) {
	logger.DBLogger.Info("DB|" + result + "|" +
		utils.Num2Str(utils.Elapsed(begin)) + "ms|" +
		sql + "|" + retResultMsg + "|" +
		utils.Num2Str(selectNum))
	//fmt.Println("DB|" + result + "|" +
	//	utils.Num2Str(utils.Elapsed(begin)) + "ms|" +
	//	sql + "|" + retResultMsg + "|" +
	//	utils.Num2Str(selectNum))
	//pl_boot.AppCtx.GetQueryLogger().Info("DB|" + result + "|" +
	//	utils.Num2Str(utils.Elapsed(begin)) + "ms|" +
	//	sql + "|" + retResultMsg + "|" +
	//	utils.Num2Str(selectNum))
}

//DoQuery 统一query
func DoQuery(dba gorose.IOrm) (err error) {
	start := time.Now()
	var sqlRunInfo [3]string //{success|error,sql,other}
	var sql string

	err = dba.Select()
	sql = dba.LastSql()
	if err != nil {
		//error
		sqlRunInfo[0] = err.Error()
		sqlRunInfo[1] = sql
		sqlRunInfo[2] = ""
	} else {
		//success
		sqlRunInfo[0] = SQLSuccessMsg
		sqlRunInfo[1] = sql
		sqlRunInfo[2] = ""
	}
	OnSQLRun(sqlRunInfo[1], start, sqlRunInfo[0], sqlRunInfo[2], 0)
	return
}

//DoInsert 统一插入调用
func DoInsert(m interface{}) (lastInsertID int64, err error) {
	start := time.Now()
	var sqlRunInfo [3]string //{success|error,sql,other}
	dba := DB().Data(m)

	fields := utils.GetKeysByTag(m, "gorose")
	var affectRow int64
	affectRow, err = dba.ExtraCols(fields...).Insert()
	sql := dba.LastSql()
	lastInsertID = dba.LastInsertId()
	if err != nil {
		sqlRunInfo[0] = err.Error()
		sqlRunInfo[1] = sql
		sqlRunInfo[2] = ""
	} else {
		sqlRunInfo[0] = SQLSuccessMsg
		sqlRunInfo[1] = sql
		sqlRunInfo[2] = ""
	}
	OnSQLRun(sqlRunInfo[1], start, sqlRunInfo[0], sqlRunInfo[2], int(affectRow))
	return
}

//DoUpdate 统一更新调用
func DoUpdate(dba gorose.IOrm) (affectRow int64, err error) {
	start := time.Now()
	var sqlRunInfo [3]string //{success|error,sql,other}

	affectRow = 0
	affectRow, err = dba.Update()
	sql := dba.LastSql()
	if err != nil {
		//error
		sqlRunInfo[0] = err.Error()
		sqlRunInfo[1] = sql
		sqlRunInfo[2] = ""
	} else {
		sqlRunInfo[0] = SQLSuccessMsg
		sqlRunInfo[1] = sql
		sqlRunInfo[2] = ""
	}
	OnSQLRun(sqlRunInfo[1], start, sqlRunInfo[0], sqlRunInfo[2], int(affectRow))
	return
}
