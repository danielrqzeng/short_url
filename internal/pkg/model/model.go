//nolint:golint,lll,dupl,structcomment,funccomment
//为何要对这个做nolint呢，是因为这个文件是gorose根据db生成的，并非人写的，若是认为修改，下次再次生成还得改一波，是以做nolint

package model

import "time"

type TAvailablePhraseInfo struct {
	Id           int       `gorose:"id" json:"id"`                         // id
	PhraseType   int       `gorose:"phrase_type" json:"phrase_type"`       // 短语类型，0:none,1:短语,2:正则表达式
	Phrase       string    `gorose:"phrase" json:"phrase"`                 // 短语|正则式
	CreateTs     int       `gorose:"create_ts" json:"create_ts"`           // 创建时间
	Version      int       `gorose:"version" json:"version"`               // 数据版本,cas update辅佐作用
	LastUpdateTs time.Time `gorose:"last_update_ts" json:"last_update_ts"` // 上次更新时间戳
}

func (*TAvailablePhraseInfo) TableName() string {
	return "t_available_phrase_info"
}

type TForbidPhraseInfo struct {
	Id           int       `gorose:"id" json:"id"`                         // id
	PhraseType   int       `gorose:"phrase_type" json:"phrase_type"`       // 短语类型，0:none,1:短语,2:正则表达式
	Phrase       string    `gorose:"phrase" json:"phrase"`                 // 短语|正则式
	CreateTs     int       `gorose:"create_ts" json:"create_ts"`           // 创建时间
	Version      int       `gorose:"version" json:"version"`               // 数据版本,cas update辅佐作用
	LastUpdateTs time.Time `gorose:"last_update_ts" json:"last_update_ts"` // 上次更新时间戳
}

func (*TForbidPhraseInfo) TableName() string {
	return "t_forbid_phrase_info"
}

type TKVInfo struct {
	KeyInfo      string    `gorose:"key_info" json:"key_info"`             // key_info
	ValStrInfo   string    `gorose:"val_str_info" json:"val_str_info"`     // val,字符串格式
	ValIntInfo   uint64    `gorose:"val_int_info" json:"val_int_info"`     // val,整数格式
	DescInfo     string    `gorose:"desc_info" json:"desc_info"`           // desc
	CreateTs     int       `gorose:"create_ts" json:"create_ts"`           // 创建时间
	Version      int       `gorose:"version" json:"version"`               // 数据版本,cas update辅佐作用
	LastUpdateTs time.Time `gorose:"last_update_ts" json:"last_update_ts"` // 上次更新时间戳
}

func (*TKVInfo) TableName() string {
	return "t_kv_info"
}

type TURLInfo struct {
	Id             uint64    `gorose:"id" json:"id"`                             // id
	ShortCode      string    `gorose:"short_code" json:"short_code"`             // 短码
	ShortUrl       string    `gorose:"short_url" json:"short_url"`               // 短网址
	RawUrl         string    `gorose:"raw_url" json:"raw_url"`                   // 原始网址
	UrlType        int       `gorose:"url_type" json:"url_type"`                 // 类型，0:none,1:incr id,2:phrase
	Status         int       `gorose:"status" json:"status"`                     // 状态，0:none,1:enable,2:ban,3:notexist
	BanAt          int       `gorose:"ban_at" json:"ban_at"`                     // 被禁时间戳
	UnbanAt        int       `gorose:"unban_at" json:"unban_at"`                 // 解禁时间戳
	BanCuz         string    `gorose:"ban_cuz" json:"ban_cuz"`                   // 禁止原因
	BanBy          string    `gorose:"ban_by" json:"ban_by"`                     // 被谁禁止，sys为系统自动禁止
	RedirectTime   int       `gorose:"redirect_time" json:"redirect_time"`       // 重定向次数
	LastRedirectTs int       `gorose:"last_redirect_ts" json:"last_redirect_ts"` // 最近一次重定向的时间戳
	CreateTs       int       `gorose:"create_ts" json:"create_ts"`               // 创建时间
	Version        int       `gorose:"version" json:"version"`                   // 数据版本,cas update辅佐作用
	LastUpdateTs   time.Time `gorose:"last_update_ts" json:"last_update_ts"`     // 上次更新时间戳
}

func (*TURLInfo) TableName() string {
	return "t_url_info"
}
