/**
* @file login.go
* @brief  登陆model
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

//登陆回执数据
type LoginResp struct {
	Errno  string `json:"errno"`
	Errmsg string `json:"errmsg"`
}

//登陆请求数据
type LoginReq struct {
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
}

type LoginController struct {
	beego.Controller
}

func (this *LoginController) RetData(rep interface{}) {
	this.Data["json"] = rep
	this.ServeJSON()
}

func (this *LoginController) Post() {
	var req LoginReq

	rep := LoginResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	//request
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &req)
	if err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(utils.RECODE_DATAERR)
		return
	}
	fmt.Printf("Login RequestInfo:%+v\n", req)

	//校验信息
	if req.Mobile == "" || req.Password == "" {
		rep.Errno = utils.RECODE_PARAMERR
		rep.Errmsg = utils.RecodeText(utils.RECODE_PARAMERR)
		return
	}

	//查询数据库
	var user models.User
	o := orm.NewOrm()
	err = o.QueryTable("user").Filter("mobile", req.Mobile).One(&user)
	if err == orm.ErrNoRows {
		//没有该用户
		rep.Errno = utils.RECODE_LOGINERR
		rep.Errmsg = utils.RecodeText(utils.RECODE_LOGINERR)
		return
	}

	if user.Password_hash != req.Password {
		//密码错误
		rep.Errno = utils.RECODE_PWDERR
		rep.Errmsg = utils.RecodeText(utils.RECODE_PWDERR)
		return
	}

	//存入session
	this.SetSession("user_id", user.Id)
	if user.Name == "" {
		this.SetSession("name", user.Mobile)
	} else {
		this.SetSession("name", user.Name)
	}
	this.SetSession("mobile", user.Mobile)
	return
}
