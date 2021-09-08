package libs

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	//	"io/ioutil"
	"strings"

	"net/url"
	"os"
	"time"

	"github.com/astaxie/beego/httplib"
)

type HttpRequest struct {
	Hubkey   bool //开启url连接后面带一个唯一key
	NetCount int8 //网络错误(包含404 500等)或超时重复次数
	Ctime    int8 //请求地址超时(单位秒)
	Rwtime   int8 //数据读取超时(单位秒)
	Interval int8 //重发时间间隔(单位秒)
}

type Callfunc func(url string, obj map[string]interface{}, dataType string) (string, error)

//post 请求
func (h *HttpRequest) HttpPost(url string, obj map[string]interface{}, dataType string) (string, error) {
	return h.parseHttpRequest(func(url string, obj map[string]interface{}, dataType string) (string, error) {
		return h.http(url, obj, "post", dataType)
	}, h.httpBuildQuery(url, obj), obj, dataType)
}

//get 请求
func (h *HttpRequest) HttpGet(url string, obj map[string]interface{}, dataType string) (string, error) {
	return h.parseHttpRequest(func(url string, obj map[string]interface{}, dataType string) (string, error) {
		return h.http(url, nil, "get", dataType)
	}, h.httpBuildQuery(url, obj), nil, dataType)
}

//重复机制回调
func (h *HttpRequest) parseHttpRequest(callfunc Callfunc, url string, obj map[string]interface{},
	dataType string) (result string, err error) {
	result, err = callfunc(url, obj, dataType)
	if err != nil && h.NetCount > 1 {
		if h.Interval > 0 {
			time.Sleep(time.Duration(h.Interval) * time.Second)
		}
		h.NetCount--
		//result, err = h.parseHttpRequest(callfunc, url, obj, dataType)
	}
	return result, err
}

//把body中get参数合并到url上
func (h *HttpRequest) httpBuildQuery(sendurl string, obj map[string]interface{}) string {
	if obj["get"] != nil {
		u, _ := url.Parse(sendurl)
		q := u.Query()
		for k, v := range obj["get"].(map[string]interface{}) {
			str, b := v.(string)
			if b {
				q.Set(k, str)
			} else {
				b, _ := json.Marshal(v)
				q.Set(k, string(b))
			}
		}
		if h.Hubkey != false {
			q.Set("hubkey", GetUuid(""))
		}
		u.RawQuery = q.Encode()
		sendurl = u.String()
	}
	return sendurl
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

//http请求
func (h *HttpRequest) http(url string, obj map[string]interface{}, httpType string, dataType string) (string, error) {

	//	fmt.Println(obj)

	var req *httplib.BeegoHTTPRequest

	fmt.Println(httpType)
	if dataType == "postfile" {
		req = httplib.Post(url)
		//req.Param("access_token", "16_uo3VU_3FzD56VWDYIXqJgGiRWYFVVyPp4wGd1qqp_v-qLaj7nNVPm7HqcR0dl940tk1JxVl5Q9ctOK4tDQ5MTAa3ZKv09AY1Zg9dIRnt3a6JpBC7MGwJ3DgFLvepgoDjvFabOukq2oTTa82vUGWfAIABND")
		//req.Param("type", "image")

		//fmt.Println(obj["file"].(string))

		//req.PostFile("media", "\r\n")
		//file, _ := os.Open(obj["file"].(string))
		//defer file.Close()

		//req.Header("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		//			escapeQuotes("media"), escapeQuotes(obj["file"].(string))))
		//req.Header("Content-Type", "image/png")

		req.PostFile("media", obj["file"].(string))

		//bt, _ := ioutil.ReadFile(obj["file"].(string))
		//req.Body(bt)
		//req.Debug(true)
		fmt.Println("@@@@@@@@@@@", req)

		_, e := req.Response()
		a, _ := req.String()

		fmt.Println(a, e)
		os.Exit(0)
	} else if httpType == "post" {
		req = httplib.Post(url)
		if obj["post"] != nil {
			for k, v := range obj["post"].(map[string]interface{}) {
				str, b := v.(string)
				if b {
					req.Param(k, str)
				} else {
					b, _ := json.Marshal(v)
					req.Param(k, string(b))
				}
			}
		}
		if obj["body"] != nil {
			if dataType == "json" {
				//为什么使用 jsonbody  ---需要考虑下 好像是48小时客服的数据需要的是json格式接受
				req.JSONBody(obj["body"])
			} else {
				req.Body(obj["body"])
				//初始化设置header，如不设置，系统使用默认，当触发json数据类型时，header被更改，再次使用form格式则没被有被更改，则post的form数据接收方收不到
				req.Header("Content-Type", "application/x-www-form-urlencoded; param=value")
			}
		}
	} else {
		req = httplib.Get(url)
	}

	req.SetTimeout(time.Duration(h.Ctime)*time.Second, time.Duration(h.Rwtime)*time.Second)
	req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := req.Response()
	//fmt.Println("^^^^^^^^^^^^^", err)
	if err != nil {
		return "", err
	}
	msg, err := req.String()

	if err != nil || res.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Error:%s Url:%s StatusCode:%d", err, url, res.StatusCode))
		return "", err
	}
	return msg, nil
}
