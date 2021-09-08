package routers

import (
	"github.com/wutiyang/drawing/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/tasks/drawing/article", &controllers.TasksController{}, "*:ArticelDraw")
	beego.Router("/tasks/drawing/draw", &controllers.TasksController{}, "*:Draw")
	beego.Router("/", &controllers.TasksController{}, "*:Index")

	//beego.AutoRouter(&controllers.TasksController{})
	//beego.AutoRouter(&controllers.RegisterController{})
}
