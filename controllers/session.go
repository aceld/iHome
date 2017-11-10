/**
* @file session.go
* @brief  会话认证model
* @author

Aceld(LiuDanbing)

email: danbing.at@gmail.com
Blog: http://www.gitbook.com/@aceld

* @version 1.0
* @date 2017-11-05
*/
package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"iHome/utils"
)

type SessionUsername struct {
	Name string `json:"name"`
}

type SessionResp struct {
	Errno  string          `json:"errno"`
	Errmsg string          `json:"errmsg"`
	Data   SessionUsername `json:"data"`
}

type SessionController struct {
	beego.Controller
}

func (this *SessionController) Get() {
	name := this.GetSession("name")

	fmt.Printf("name = %+v\n", name)

	if name == nil {
		this.Data["json"] = &SessionResp{Errno: utils.RECODE_SESSIONERR, Errmsg: utils.RecodeText(utils.RECODE_SESSIONERR)}
	} else {
		data := SessionUsername{Name: fmt.Sprintf("%s", name)}
		this.Data["json"] = &SessionResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK), Data: data}
	}

	this.ServeJSON()
}

// /session  [Delete]
func (this *SessionController) Delete() {
	this.DelSession("user_id")
	this.DelSession("name")
	this.Data["json"] = &SessionResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	this.ServeJSON()
}
