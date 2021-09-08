package controllers

import (
	"fmt"

	"github.com/wutiyang/drawing/draw"
	"github.com/wutiyang/drawing/libs"

	//	"fmt"

	"github.com/astaxie/beego"
)

var (
	//网络重试次数
	retry  int8 = 3
	expire int  = 86400
)

// Tasks API
type TasksController struct {
	beego.Controller
}

func (c *TasksController) Draw() {

	//	var ob map[string]interface{}

	appid := c.GetString("appid")

	//	if err := ffjson.Unmarshal(c.Ctx.Input.RequestBody, &ob); err != nil {
	//		c.Ctx.WriteString(libs.ReturnInfo(10002, "解析body错误"))
	//		return
	//	}
	//	if len(ob) == 0 {
	//		c.Ctx.WriteString(libs.ReturnInfo(10001, "缺少body参数"))
	//		return
	//	}

	media_id, create_at := draw.Run(appid, c.Ctx.Input.RequestBody)

	info := map[string]interface{}{
		"media_id":  media_id,
		"create_at": create_at,
	}
	c.Ctx.WriteString(libs.ReturnData(0, info))
}

func (c *TasksController) Index() {
	c.Ctx.WriteString("index")
}

//Receive info by Http,Filter info push queue.
/*
*bat -body="[{\"post\":{\"touser\":[\"oddklw8VCtb5pnAk2_qQyZNDbkc4\"],\"format\":{\"msgtype\":\"text\",\"text\":{\"content\":\"aaaaaaaaaaaa\"}}},\"appid\":\"wx312fb00864252191\",\"remark\":\"\",\"source\":\"2\"}]"  POST :/tasks/post?task_type=fehours
 */

// 海报合成(升级)
func (c *TasksController) ArticelDraw() {
	appid := c.GetString("appid")
	fmt.Println(appid)

	media_id, create_at := draw.ArticleRun(appid, c.Ctx.Input.RequestBody)

	// 返回客户端信息
	info := map[string]interface{}{
		"media_id":  media_id,
		"create_at": create_at,
	}
	c.Ctx.WriteString(libs.ReturnData(0, info))
}
