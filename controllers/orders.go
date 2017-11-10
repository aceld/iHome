/**
* @file orders.go
* @brief  订单model相关controller
* @author

Aceld(LiuDanbing)

email: danbing.at@gmail.com
Blog: http://www.gitbook.com/@aceld

* @version 1.0
* @date 2017-11-10
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"iHome/models"
	"iHome/utils"
	"strconv"
	"time"
)

type OrderRequest struct {
	House_id   string `json:"house_id"`   //下单的房源id
	Start_date string `json:"start_date"` //订单开始时间
	End_date   string `json:"end_date"`   //订单结束时间
}

type OrderResp struct {
	Errno  string      `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

type OrdersController struct {
	beego.Controller
}

func (this *OrdersController) RetData(rep interface{}) {
	this.Data["json"] = rep
	this.ServeJSON()
}

// /orders [POST]
func (this *OrdersController) Post() {
	rep := OrderResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	//得到当前用户id
	user_id := this.GetSession("user_id")

	//获得客户端请求数据
	var req OrderRequest
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	fmt.Printf("req = %+v\n", req)

	//用户参数做合法判断
	if req.House_id == "" || req.Start_date == "" || req.End_date == "" {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = "请求参数为空"
		return
	}
	//格式化日期时间
	start_date_time, _ := time.Parse("2006-01-02 15:04:05", req.Start_date+" 00:00:00")
	end_date_time, _ := time.Parse("2006-01-02 15:04:05", req.End_date+" 00:00:00")
	//确保end_date 在 start_date之后
	if end_date_time.Before(start_date_time) {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = "结束时间在开始时间之前"
		return
	}
	fmt.Printf("start_date_time = %+v, end_date_time = %+v\n", start_date_time, end_date_time)
	//得到入住天数
	days := end_date_time.Sub(start_date_time).Hours()/24 + 1
	fmt.Printf("days = %f\n", days)

	//根据house_id 得到房屋信息
	house_id, _ := strconv.Atoi(req.House_id)
	house := models.House{Id: house_id}
	o := orm.NewOrm()
	if err := o.Read(&house); err != nil {
		rep.Errno = utils.RECODE_NODATA
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	o.LoadRelated(&house, "User")
	//房东不能够预定自己的房子
	if user_id == house.User.Id {
		rep.Errno = utils.RECODE_ROLEERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//TODO 确保用户选择的房屋未被预定,日期没有冲突

	amount := days * float64(house.Price)
	order := models.OrderHouse{}
	order.House = &house
	user := models.User{Id: user_id.(int)}
	order.User = &user
	order.Begin_date = start_date_time
	order.End_date = end_date_time
	order.Days = int(days)
	order.House_price = house.Price
	order.Amount = int(amount)
	order.Status = models.ORDER_STATUS_WAIT_ACCEPT

	fmt.Printf("order = %+v\n", order)

	if _, err := o.Insert(&order); err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	this.SetSession("user_id", user_id)
	data := map[string]interface{}{}
	data["order_id"] = order.Id
	rep.Data = data
	return
}

// /orders/:id/status [PUT]
func (this *OrdersController) OrderStatus() {
	rep := OrderResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	return
}
