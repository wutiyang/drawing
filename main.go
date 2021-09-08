package main

import (
	"github.com/wutiyang/drawing/libs"
	_ "github.com/wutiyang/drawing/routers"

	//"github.com/wutiyang/drawing/draw"

	"github.com/astaxie/beego"
)

var (
	path string = "conf/httpconfig.toml"
	log  string = "runtime/logs.log"
)

func init() {
	libs.InitSingleInfo(path)

}

func main() {
	//init conf info
	//info := map[string]string{"transapp": path}
	//draw.Run("wxc441d1e8dcfffc91")

	beego.BeeLogger.SetLogger("file", `{"filename":"`+log+`","daily":true,"maxdays":1}`)
	beego.Run()

}
