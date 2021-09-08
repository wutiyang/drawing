// hello project loger.go
package libs

import (
	"fmt"

	"github.com/astaxie/beego/logs"
)

var Log *Loger

type Loger struct {
	Logpath   string
	LogMapObj map[string]*logs.BeeLogger
}

func init() {

}

func NewLoger(conf *S1) *Loger {
	Log = &Loger{}
	Log.Logpath = conf.LogPath
	Log.LogMapObj = make(map[string]*logs.BeeLogger, 0)
	return Log
}

func (l *Loger) Info(fileName string, message string) {
	logObj := l.logInit(fileName, "Info")
	logObj.Info(message)
}

func (l *Loger) Error(fileName string, message interface{}) {
	logObj := l.logInit(fileName, "Error")
	switch message.(type) {
	case string:
		logObj.Error(message.(string))
	default:
		logObj.Error(fmt.Sprintf("", message.(error)))
	}
}

func (l *Loger) logInit(fileName string, level string) (logObj *logs.BeeLogger) {
	if _, ok := l.LogMapObj[fileName]; !ok {
		logObj = logs.NewLogger(100000)
		filePath := fmt.Sprintf("%s.%s.log", l.Logpath+fileName, level)
		//		logObj.SetLogger("file", `{"filename":"`+filePath+`"}`)
		logObj.SetLogger("file", `{"filename":"`+filePath+`","daily":true,"maxdays":1}`)
		l.LogMapObj[fileName] = logObj
	} else {
		logObj = l.LogMapObj[fileName]
	}
	return logObj
}
