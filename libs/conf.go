package libs

import (
	//	"fmt"
	"time"

	"github.com/koding/multiconfig"
)

var ServerConf *S1

type ServerArray struct {
	serverList []S1
}

type (
	S1 struct {
		Name         string
		Enabled      bool
		OpenWeb      bool
		RedisKey     string
		RedisYinman  string
		BackRedisKey string
		TaskListNum  int
		TaskNum      int
		LogPath      string `default:"runtime/"`
		UrlType      int    `default:0`
		PicPath      string
		Concurrent   time.Duration
		Url          Url
		Callback     Callback
		Redis        Redis
		TokenRedis   TokenRedis
		Httpinfo     Httpinfo
		Keys         Keys
		Other        Other
		Wx           Wx
		CX           CX
		Qiniu        Qiniu
	}

	Redis struct {
		Host     string
		Port     string `default:"6379"`
		Db       string `default:"0"`
		Password string
		Pre      string
	}
	TokenRedis struct {
		Host     string
		Port     string `default:"6379"`
		Db       string `default:"0"`
		Password string
		Pre      string
	}

	Httpinfo struct {
		Ctime    int8 `default:30`
		Rwtime   int8 `default:5`
		NetCount int8 `default:0`
	}

	Callback struct {
		TagsUrl    string
		TempUrl    string
		FehoursUrl string
	}
	Keys struct {
		CustomerServiceStatus string
		QueueMap              string
		QueueName             string
		TokenKey              string
		QueueMapInstant       string
	}

	Url struct {
		WxUrl    string
		TokenUrl string
	}

	Other struct {
		Eventids                             string
		MarketingAutomationEventIds          string
		MarketingAutomationCallbackSendTmpId string
		MarketingAutomation                  string
		MarketingAutomationTmp               string
	}
	CX struct {
		CXEventIds string
		CXName     string
		CXRedisKey string
		CXType     int8
	}
	Wx struct {
		PicUrlPre string
	}
	Qiniu struct {
		ACCESS_KEY string
		SECRET_KEY string
		BUCKET     string
		QINIUHOST  string
	}
)

// init args.
func init() {
	//	m := multiconfig.NewWithPath("config/config.toml") // supports TOML and JSON
	//	// Get an empty struct for your configuration
	//	ServerConf = new(S1)
	//	// Populated the serverConf struct
	//	m.MustLoad(ServerConf) // Check for error
	//	fmt.Println("After Loading: ")
	//	fmt.Printf("%+v\n", ServerConf)
	//	if ServerConf.Enabled {
	//		fmt.Println("Enabled field is set to true")
	//	} else {
	//		fmt.Println("Enabled field is set to false")
	//	}
}

func NewConfMan() *ServerArray {
	return &ServerArray{make([]S1, 0)}
}

func NewConf() *S1 {
	return &S1{}
}

func Add() {}

func ReloadConf(confPath string, num time.Duration, c *S1) {

	timer := time.NewTicker(num * time.Second)
	for {
		select {
		case <-timer.C:
			c = StartConf(confPath)
		}
	}
}

func StartConf(confPath string) *S1 {
	m := multiconfig.NewWithPath(confPath) // supports TOML and JSON
	// Get an empty struct for your configuration
	serverConf := new(S1)
	// Populated the serverConf struct
	m.MustLoad(serverConf) // Check for error
	//	fmt.Println("After Loading: ")
	//	fmt.Printf("%+v\n", serverConf)
	return serverConf
}
