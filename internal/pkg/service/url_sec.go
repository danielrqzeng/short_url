package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//OnRequestDone log when http request done
func OnRequestDone(begin time.Time, retCode int, msg string, url, rBody, wBody string) {
	elasped := time.Since(begin).String()
	logger.APILogger.Info(url + "|" + elasped + "|" + utils.Num2Str(retCode) + "|" + msg + "|" + rBody + "|" + wBody)
}

/*
此服务是想检查是否这个url是安全的，比如涉黄，涉控都是不被允许的
目前有的检测网址合法性的有
1.安全联盟：https://www.anquan.org，这个是政府牵头，多家大企业参与的，权威可信，还提供了易用的api给大家乱用
2.腾讯网址安全检测：https://urlsec.qq.com/check.html，这个调用很麻烦，不鸟它了
3.孟坤工具箱，一个包装腾讯网址安全检测的检测：http://tool.mkblog.cn/UrlSecurity/，这个可以调用
4.腾讯110：https://110.qq.com/，这个有api调用，可以使用
5.极强检测，一个综合n多检测的网址：https://user.urlzt.com/comdet?url=pornhub.com，这个可以参考有哪些提供了网址检测
*/

/*--------------孟坤工具箱----------------*/
func getSession() (session string) {
	const homePage = "http://tool.mkblog.cn/webscan/"
	rsp, _ := http.Get(homePage)
	defer rsp.Body.Close()
	body, _ := ioutil.ReadAll(rsp.Body)
	_ = body
	cs := rsp.Cookies()
	for _, c := range cs {
		if c.Name == "PHPSESSID" {
			session = c.Value
			return
		}
	}
	return
}

//UrlSecurityResponse 孟坤工具箱检测结果
//jQuery3110313769350759209_1621215738028({"data":{"retcode":0,"results":{"url":"qq.com","whitetype":3,"WordingTitle":"","Wording":"","detect_time":"1514189828","eviltype":"0","certify":0,"isDomainICPOk":1,"Orgnization":"\u6df1\u5733\u5e02\u817e\u8baf\u8ba1\u7b97\u673a\u7cfb\u7edf\u6709\u9650\u516c\u53f8","ICPSerial":"\u7ca4B2-20090059-5"}},"reCode":0})
type UrlSecurityResponse struct {
	Data struct {
		RetCode int `json:"retcode"`
		Results struct {
			Url           string `json:"url"`
			WhiteType     int    `json:"whitetype"`    //检测结果，1:未知,2:网站存在风险,3:安全网站,4:腾讯官方网站,13:可信度低,2001:可能涉及钱财，交易需谨慎
			WordingTitle  string `json:"WordingTitle"` //当whitetype==2时候，危险原因
			Wording       string `json:"Wording"`      //当whitetype==2时候，危险描述
			DetectTime    string `json:"detect_time"`  //检出时间,是一个时间戳，代表这个网址什么时候做的检测，显然其会定期检测
			Eviltype      string `json:"eviltype"`
			Certify       int    `json:"certify"`
			IsDomainICPOk int    `json:"isDomainICPOk"` //是否已经备案，0:no,1:yes
			Orgnization   string `json:"Orgnization"`   //备案：主办方
			ICPSerial     string `json:"ICPSerial"`     //备案：备案号
		} `json:"results"`
	} `json:"data"`
	ReCode int `json:"reCode"`
}

//CheckByMK 孟坤工具箱检测
func CheckByMK(targetUrl string) (IsSecurity bool, err error) {
	begin := time.Now()
	nlpServerCode := -1
	rBody := ""
	wBody := ""
	IsSecurity = false
	callbackTag := fmt.Sprintf("jQuery3110313769350759209_%d", utils.NowMs())
	vals := url.Values{}
	vals.Add("url", targetUrl)
	vals.Add("callback", callbackTag)
	vals.Add("types", "qq")
	vals.Add("_", "1621215738030")
	urlPath := fmt.Sprintf("http://tool.mkblog.cn/webscan/?%s", vals.Encode())

	//urlPath := fmt.Sprintf(
	//	"http://tool.mkblog.cn/webscan/?callback=%s&types=qq&url=%s&_=1621215738030",
	//	callbackTag,
	//	targetUrl)
	req, _ := http.NewRequest("GET", urlPath, nil)

	req.Header.Add(
		"accept",
		"text/javascript, application/javascript, application/ecmascript, application/x-ecmascript, */*; q=0.01")
	req.Header.Add(
		"user-agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("referer", "http://tool.mkblog.cn/UrlSecurity/")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7,id;q=0.6")
	req.Header.Add("cookie", fmt.Sprintf("PHPSESSID=%s,", getSession()))
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	logger.MainLogger.Debug(string(body))
	if err != nil {
		logger.MainLogger.Debug(err.Error())
		OnRequestDone(begin, nlpServerCode, err.Error(), urlPath, rBody, wBody)
		return
	}
	if !strings.HasPrefix(string(body), callbackTag) {
		OnRequestDone(begin, nlpServerCode, "strings.HasPrefix fail", urlPath, rBody, wBody)
		return
	}
	nlpServerCode = 200
	wBody = string(body)
	jsonStr := strings.TrimPrefix(string(body), callbackTag+"(")
	jsonStr = strings.TrimSuffix(jsonStr, ")")

	rspStruct := &UrlSecurityResponse{}
	err = json.Unmarshal([]byte(jsonStr), rspStruct)
	if err != nil {
		OnRequestDone(begin, nlpServerCode, err.Error(), urlPath, rBody, wBody)
		return
	}

	if rspStruct.ReCode != 0 {
		err = fmt.Errorf("rspStruct.ReCode=" + utils.Num2Str(rspStruct.ReCode))
		OnRequestDone(begin, nlpServerCode, err.Error(), urlPath, rBody, wBody)
		return
	}

	if rspStruct.Data.RetCode != 0 {
		err = fmt.Errorf("rspStruct.Data.RetCode=" + utils.Num2Str(rspStruct.Data.RetCode))
		OnRequestDone(begin, nlpServerCode, err.Error(), urlPath, rBody, wBody)
		return
	}

	OnRequestDone(begin, nlpServerCode, "success", urlPath, rBody, wBody)
	if rspStruct.Data.Results.WhiteType == 3 || rspStruct.Data.Results.WhiteType == 4 {
		IsSecurity = true
		return
	}

	return
}

/*--------------孟坤工具箱 end----------------*/

/*--------------安全联盟----------------*/
//AnQuanLianMenResponse ...
//{"code":1002,"success":false,"msg":"域名正常","results":null}
//{"code":1001,"success":true,"msg":"域名被拉黑","results":"色情网站"}
type AnQuanLianMenResponse struct {
	Code    int     `json:"code"`    //1002:正常,1004:申诉被冻结,1005:已在申诉处理中,其他不正常，不正常原因在results中写明
	Success bool    `json:"success"` //不需要管这个，一般false代表网址正常，true代表网址有问题
	Msg     string  `json:"msg"`     //说明结果，比如：域名正常，域名被拉黑
	Results *string `json:"results"` //说明原因，只有有问题的网址才有有这个字段，否则是空的
}

//CheckByAnQuanLianMen 通过调用安全联盟检测
func CheckByAnQuanLianMen(targetUrl string) (IsSecurity bool, err error) {
	begin := time.Now()
	nlpServerCode := -1
	rBody := ""
	wBody := ""
	//curl 'https://www.anquan.org/intercept/web/check/?domain_name=www.youtube.com'
	IsSecurity = false
	vals := url.Values{}
	vals.Add("domain_name", targetUrl)
	urlPath := fmt.Sprintf("https://www.anquan.org/intercept/web/check/?%s", vals.Encode())
	req, _ := http.NewRequest("GET", urlPath, nil)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	jsonStr, err := ioutil.ReadAll(res.Body)
	nlpServerCode = 200
	wBody = string(jsonStr)

	rspStruct := &AnQuanLianMenResponse{}
	err = json.Unmarshal([]byte(jsonStr), rspStruct)
	if err != nil {
		OnRequestDone(begin, nlpServerCode, err.Error(), urlPath, rBody, wBody)
		return
	}
	OnRequestDone(begin, nlpServerCode, "success", urlPath, rBody, wBody)
	switch rspStruct.Code {
	case 1002:
		IsSecurity = true
		return
	default:
		return
	}
	return
}

//IsUrlSecurity 是否url是安全的
//@TODO 性能优化：这个其实还可以在先做一层缓存,黑白名单之类的，以免要调用api（这个api很耗时，150ms每次调用）
func IsUrlSecurity(targetUrl string) (IsSecurity bool, err error) {
	return CheckByAnQuanLianMen(targetUrl) //调用安全联盟api来检测
}
