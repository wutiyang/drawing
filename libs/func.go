package libs

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

//获取uuid
func GetUuid(num string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num1 := fmt.Sprintf("%d", r.Intn(10000))
	num2 := fmt.Sprintf("%d", r.Intn(10000))
	num3 := fmt.Sprintf("%d", time.Now().UnixNano())

	uuid := num3 + num1 + num2 + num
	//加密
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(uuid))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

//获取主机名
func GetHostName() string {
	host, err := os.Hostname()
	if err != nil {
		return "127.0.0.1"
	}
	return host
}

//服务状态
func CloseServer(appName string, args ...int) bool {
	hostName := GetHostName()
	if r, _ := RedisInfo.Get("isclose:" + hostName + ":" + appName); r == "1" {
		var isok bool = true
		if len(args) > 0 {
			for _, v := range args {
				if v > 0 {
					isok = false
					break
				}
			}
		}
		if isok == true {
			time.Sleep(60 * time.Second)
			Log.Info("CloseInfo", "已关闭服务")
			os.Exit(0)
		}
	}
	return false
}

//返回json
func ReturnInfo(code float64, msg string) string {
	ri := map[string]interface{}{"code": code, "msg": msg}
	buf, _ := ffjson.Marshal(ri)
	return string(buf)
}

//返回json
func ReturnData(code float64, data interface{}) string {
	ri := map[string]interface{}{"code": code, "data": data}
	buf, _ := ffjson.Marshal(ri)
	return string(buf)
}

// logtype = 1 正常日志，其他是错误日志
func PrintlnLog(logType int8, filename string, message interface{}) {

	if logType == 1 {

		Log.Info(filename, message.(string))
	} else {
		Log.Error(filename, message)
	}

}

func U2s(form string) (to string, err error) {
	//	fmt.Println("===1", form)
	form = strings.TrimLeft(form, `"`)
	form = strings.TrimRight(form, `"`)

	formstr1 := strings.Replace(form, `\"`, `'`, -1)
	//	fmt.Println("===2", formstr1)
	formstr := strings.Replace(formstr1, `\/`, `/`, -1)
	//	fmt.Println("===3", formstr)
	to, err = strconv.Unquote(`"` + formstr + `"`)

	return
}

func In_array(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

//判断是否在redis的map中
func In_redis_map() {

}

type TokenInfo struct {
	Code        int    `json:"code"`
	Errcode     int    `json:"errcode"`
	Msg         string `json:"msg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Type        string `json:"type"`
}

func GetToken(appid string) string {
	var tokenJson string
	var tokenInfo TokenInfo
	var err error
	//获取redis缓存中的access_token json

	redislink := NewRedisConnect(ServerConf.TokenRedis.Host, ServerConf.TokenRedis.Port, ServerConf.TokenRedis.Password, ServerConf.TokenRedis.Db)

	if tokenJson, err = redislink.Get(fmt.Sprintf(ServerConf.Keys.TokenKey, appid)); err != nil {
		Log.Error("redis", fmt.Sprintf("access_token get error:", err))
	}

	if err = ffjson.Unmarshal([]byte(tokenJson), &tokenInfo); err != nil {
		Log.Error("redis", fmt.Sprintf("access_token format error:", err))
	}
	return tokenInfo.AccessToken
	//url := fmt.Sprintf(ReplyUrl, tokenInfo.AccessToken)
}

func GetTokenByHttp(appid string) string {

	h := &HttpRequest{NetCount: 3, Ctime: 5, Rwtime: 5, Interval: 3}
	url := fmt.Sprintf(ServerConf.Url.TokenUrl, appid)
	msg, err := h.HttpGet(url, nil, "json")

	fmt.Println("############2", url, msg, err)
	if err != nil {
		Log.Error("HttpNetError", err)
		return ""
	}

	var tokenInfo TokenInfo
	if err := ffjson.Unmarshal([]byte(msg), &tokenInfo); err != nil {
		Log.Error("ParseError", fmt.Sprintf("TokenInfo Json Unmarshal Str:%s Error %s", string(msg), err))

	}

	//Log.Error("test", appid)

	if tokenInfo.Code == 200 {
		return tokenInfo.AccessToken
	} else {
		return ""
	}
}
