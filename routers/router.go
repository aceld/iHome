/**
* @file router.go
* @brief  iHome-go 路由设置
* @author

Aceld(LiuDanbing)

email: danbing.at@gmail.com
Blog: http://www.gitbook.com/@aceld

* @version 1.0
* @date 2017-11-05
*/
package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"iHome/controllers"
	"net/http"
	"strings"
)

func init() {
	ignoreStaticPath()

	//beego.Router("/", &controllers.MainController{})
	//	beego.Router("/api/v1.0/users/", &controllers.RegController{})

	ns :=
		beego.NewNamespace("/api",
			beego.NSCond(func(ctx *context.Context) bool {
				if ctx.Input.Domain() == "101.200.170.171" {
					return true
				}
				beego.Debug("now domain is ", ctx.Input.Domain(), " not 101.200.170.171")
				return false
			}),
			//beego.NSBefore(auth),
			beego.NSNamespace("/v1.0",
				//注册
				beego.NSRouter("/users", &controllers.RegController{}),
				//登陆
				beego.NSRouter("/sessions", &controllers.LoginController{}),
				//请求地理区域信息
				beego.NSRouter("/areas", &controllers.AreaController{}),
				//验证用户是否已经注册
				beego.NSRouter("/session", &controllers.SessionController{}),
				//请求当前用户信息
				beego.NSRouter("/user", &controllers.UserController{}),

				//用户上传用户头像
				beego.NSRouter("/user/avatar", &controllers.UserController{}, "post:Avatar"),
				//用户更新用户名
				beego.NSRouter("/user/name", &controllers.UserController{}, "put:UpdateName"),
				//请求用户身份认证信息, 上传用户身份信息
				beego.NSRouter("/user/auth", &controllers.UserController{}, "get:Get;post:AuthPost"),
				//请求当前用户的所有发布的房源信息列表
				beego.NSRouter("/user/houses", &controllers.UserController{}, "get:GetHouses"),
				//请求当前用户提交的订单列表信息
				beego.NSRouter("/user/orders", &controllers.UserController{}, "get:GetOrders"),

				//用户发布房源信息
				beego.NSRouter("/houses", &controllers.HousesController{}, "post:Post;get:Get"),
				//用户上传房源图片
				beego.NSRouter("/houses/:id/images", &controllers.HousesController{}, "post:UploadHouseImage"),
				//用户请求房源详细信息
				beego.NSRouter("/houses/:id", &controllers.HousesController{}, "get:GetOneHouseInfo"),
				//用户请求房源首页列表信息
				beego.NSRouter("/houses/index", &controllers.HousesController{}, "get:IndexHouses"),

				//用户下单请求
				beego.NSRouter("/orders", &controllers.OrdersController{}, "post:Post"),
				//房东用户接受/拒绝 订单请求
				beego.NSRouter("/orders/:id/status", &controllers.OrdersController{}, "put:OrderStatus"),
				//用户发送订单评价信息
				beego.NSRouter("/orders/:id/comment", &controllers.OrdersController{}, "put:OrderComment"),
			),
		)

	//注册 namespace
	beego.AddNamespace(ns)
}

func ignoreStaticPath() {

	//透明static

	beego.InsertFilter("/", beego.BeforeRouter, TransparentStatic)
	beego.InsertFilter("/*", beego.BeforeRouter, TransparentStatic)
}

func TransparentStatic(ctx *context.Context) {
	/*
		if strings.Index(ctx.Request.URL.Path, "v1/") >= 0 {
			return
		}
	*/
	orpath := ctx.Request.URL.Path
	beego.Debug("request url: ", orpath)
	//如果请求uri还有api字段,说明是指令应该取消静态资源路径重定向
	if strings.Index(orpath, "api") >= 0 {
		return
	}
	http.ServeFile(ctx.ResponseWriter, ctx.Request, "static/html/"+ctx.Request.URL.Path)
}
