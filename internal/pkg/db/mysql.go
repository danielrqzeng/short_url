package db

import (
	"iyfiysi.com/short_url/internal/pkg/model"
)

func MysqlInit() {
	model.InitDBInstance()
}
