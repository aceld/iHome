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
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
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
	Data   interface{} `json:"data,omitempty"`
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

	//得到order_id
	order_id := this.Ctx.Input.Param(":id")

	//得到当前用户id
	user_id := this.GetSession("user_id").(int)
	var req map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	//得到请求指令
	action := req["action"]
	if action != "accept" && action != "reject" {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	o := orm.NewOrm()

	order := models.OrderHouse{}
	if err := o.QueryTable("order_house").Filter("id", order_id).Filter("status", models.ORDER_STATUS_WAIT_ACCEPT).One(&order); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	if _, err := o.LoadRelated(&order, "House"); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	house := order.House
	/*
		fmt.Printf("house = %+v\n", house)
		fmt.Printf("house.user_id = %d\n", house.User.Id)
	*/
	if house.User.Id != user_id {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = "订单用户不匹配,操作无效"
		return
	}

	if action == "accept" {
		//如果是接受订单,将订单状态变成待评价状态
		order.Status = models.ORDER_STATUS_WAIT_COMMENT
		beego.Debug("action = accpet!")
	} else if action == "reject" {
		//如果是拒绝接单, 尝试获得拒绝原因,并把拒单原因保存起来
		order.Status = models.ORDER_STATUS_REJECTED
		reason := req["reason"]
		order.Comment = reason.(string)
		beego.Debug("action = reject!, reason is ", reason)
	}

	//将order订单重新入库

	if _, err := o.Update(&order); err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	return
}

// /orders/:id/comment [PUT]
func (this *OrdersController) OrderComment() {
	rep := OrderResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	//获得用户id
	user_id := this.GetSession("user_id").(int)

	//得到订单id
	order_id := this.Ctx.Input.Param(":id")

	//获得参数
	var req map[string]interface{}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	comment := req["comment"].(string)
	//检验评价信息是否合法 确保不为空
	if comment == "" {
		rep.Errno = utils.RECODE_PARAMERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//查询数据库，订单必须存在，订单状态必须为WAIT_COMMENT待评价状态
	order := models.OrderHouse{}
	o := orm.NewOrm()
	if err := o.QueryTable("order_house").Filter("id", order_id).Filter("status", models.ORDER_STATUS_WAIT_COMMENT).One(&order); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//关联查询order订单所关联的user信息
	if _, err := o.LoadRelated(&order, "User"); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//确保订单所关联的用户和该用户是同一个人
	if user_id != order.User.Id {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = "该订单并不属于本人"
		return
	}

	//关联查询order订单所关联的House信息
	if _, err := o.LoadRelated(&order, "House"); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	house := order.House
	//	fmt.Printf("=========== > house = %+v\n", house)

	//更新order的status为COMPLETE
	order.Status = models.ORDER_STATUS_COMPLETE
	order.Comment = comment

	//将房屋订单成交量+1
	house.Order_count++

	//将order和house更新数据库
	if _, err := o.Update(&order, "status", "comment"); err != nil {
		beego.Error("update order status, comment error, err = ", err)
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	if _, err := o.Update(house, "order_count"); err != nil {
		beego.Error("update house order_count error, err = ", err)
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//将house_info_[house_id]的缓存key删除 （因为已经修改订单数量）
	redis_config_map := map[string]string{
		"key":   "ihome_go",
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)

	cache_conn, err := cache.NewCache("redis", string(redis_config))
	if err != nil {
		beego.Debug("connect cache error")
	}
	house_info_key := "house_info_" + strconv.Itoa(house.Id)
	if err := cache_conn.Delete(house_info_key); err != nil {
		beego.Error("delete ", house_info_key, "error , err = ", err)
	}

	return
}
