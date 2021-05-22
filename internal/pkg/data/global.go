package data

const (
	//url类型
	URLTypeTypeNone         = 0 //url类型-无
	URLTypeCreateWithID     = 1 //url类型-使用自增id创建的
	URLTypeCreateWithPhrase = 2 //url类型-使用短语创建的

	//url info状态
	URLStatusNone     = 0 //url状态-无
	URLStatusEnable   = 1 //url状态-启用中
	URLStatusBan      = 2 //url状态-禁用
	URLStatusNotExist = 3 //url状态-不存在，此处是为了做“缓存穿透”（redis和mysql都不存在数据，设置标识表示其不存在）

	//短语类型
	PhraseTypeNone   = 0 //短语类型-无
	PhraseTypeWord   = 1 //短语类型-短语
	PhraseTypeRegexp = 2 //短语类型-正则表达式

	IndexRequestErr = "redirect to index"
)
