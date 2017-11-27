/**
* @file register.go
* @brief  注册model
* @author

Aceld(LiuDanbing)

email: danbing.at@gmail.com
Blog: http://www.gitbook.com/@aceld

* @version 1.0
* @date 2017-11-05
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"iHome/models"
	"iHome/utils"
)

//注册回执数据
type RegResp struct {
	Errno  string `json:"errno"`
	Errmsg string `json:"errmsg"`
}

//注册请求数据
type RegReq struct {
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
	Sms_code string `json:"sms_code"`
}

type RegController struct {
	beego.Controller
}

func (c *RegController) RetData(rep interface{}) {
	c.Data["json"] = rep
	c.ServeJSON()
}

func (c *RegController) Post() {
	var req RegReq
	rep := RegResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer c.RetData(&rep)

	//request
	json.Unmarshal(c.Ctx.Input.RequestBody, &req)

	fmt.Printf("Reg RequestInfo:%+v\n", req)
	/*
		fmt.Println("mobile:", req.Mobile)
		fmt.Println("password:", req.Password)
		fmt.Println("sms_code:", req.Sms_code)
	*/

	if req.Mobile == "" || req.Password == "" || req.Sms_code == "" {
		rep.Errno = utils.RECODE_PARAMERR
		rep.Errmsg = utils.RecodeText(utils.RECODE_PARAMERR)
		return
	}
	//对短信验证码的校验

	user := models.User{}

	user.Mobile = req.Mobile
	user.Password_hash = req.Password

	//将user存入mysql数据库
	o := orm.NewOrm()
	id, err := o.Insert(&user)
	if err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	beego.Debug("reg insert id =", id, " succ!")

	//存入session
	c.SetSession("user_id", user.Id)
	if user.Name == "" {
		c.SetSession("name", user.Mobile)
	} else {
		c.SetSession("name", user.Name)
	}
	c.SetSession("mobile", user.Mobile)

	return
}
