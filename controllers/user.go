/**
* @file user.go
* @brief  用户model相关controller 更新用户名,实名认证,查询用户关联房屋,上传用户头像等
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
	"path"
)

//  /user 请求的回复数据
type UserResp struct {
	Errno  string      `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   models.User `json:"data"`
}

type Avatar struct {
	Url string `json:"avatar_url"`
}

// /user/avatar 请求的回复数据
type AvatarResp struct {
	Errno  string `json:"errno"`
	Errmsg string `json:"errmsg"`
	Data   Avatar `json:"data"`
}

type Name struct {
	Name string `json:"name"`
}

// /user/name 请求的回复数据
type NameResp struct {
	Errno  string `json:"errno"`
	Errmsg string `json:"errmsg"`
	Data   Name   `json:"data"`
}

type AuthInfo struct {
	RealName string `json:"real_name"`
	Id_card  string `json:"id_card"`
}

// /user/auth [POST] 请求的回复数据
type AuthResp struct {
	Errno  string `json:"errno"`
	Errmsg string `json:"errmsg"`
}

/*
type HouseInfo struct {
	House_id    int    `json:"house_id"`
	Title       string `json:"title"`
	Price       int    `json:"Price"`
	Area_name   string `json:"area_name"`
	Img_url     string `json:"img_url"`
	Room_count  int    `json:"room_count"`
	Order_count int    `json:"Order_count"`
	Address     string `json:"address"`
	User_avatar string `json:"user_avatar"`
	Ctime       string `json:"ctime"`
}
*/

type Houses struct {
	Houses []interface{} `json:"houses"`
}

// /user/Houses [GET] 请求的回复数据
type UserHousesResp struct {
	Errno  string `json:"errno"`
	Errmsg string `json:"errmsg"`

	Data Houses `json:"data"`
}

type UserController struct {
	beego.Controller
}

func (this *UserController) RetData(rep interface{}) {
	this.Data["json"] = rep
	this.ServeJSON()
}

// /user/name [PUT]
func (this *UserController) UpdateName() {
	rep := NameResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	//从session得到user_id
	user_id := this.GetSession("user_id")

	//request post data
	var req Name
	//得到客户端请求数据
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	if req.Name == "" {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = "name is Empty!"
		return
	}

	//更新数据库 User 的 name字段
	o := orm.NewOrm()
	user := models.User{Id: user_id.(int), Name: req.Name}

	if _, err := o.Update(&user, "name"); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug(err)
		return
	}

	//更新Session
	this.SetSession("user_id", user_id)
	this.SetSession("name", req.Name)

	//response data
	rep.Data = req
}

// /user/avatar [POST]
func (this *UserController) Avatar() {
	rep := AvatarResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	file, header, err := this.GetFile("avatar")
	if err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("get file error")
		return
	}
	defer file.Close()

	//保存文件到本地
	//this.SaveToFile("avatar", "static/"+header.Filename)

	//将文件存在fastDFS上
	fileBuffer := make([]byte, header.Size)
	_, err = file.Read(fileBuffer)
	if err != nil {
		rep.Errno = utils.RECODE_IOERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("read file error")
		return
	}
	//获得文件名后缀 suffix = ".png"
	suffix := path.Ext(header.Filename)

	groupName, fileId, err := utils.FDFSUploadByBuffer(fileBuffer, suffix[1:])
	if err != nil {
		rep.Errno = utils.RECODE_IOERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("fdfs upload file error")
		return
	}

	beego.Debug("groupName:", groupName, " fileId:", fileId)

	//将文件名入mysql数据库中的user表中
	user_id := this.GetSession("user_id")
	o := orm.NewOrm()
	user := models.User{Id: user_id.(int), Avatar_url: fileId}

	if _, err := o.Update(&user, "avatar_url"); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug(err)
		return
	}

	//将fileid和服务器域名拼接,返回给前端
	rep.Data.Url = utils.AddDomain2Url(fileId)

	return
}

// /user [GET]
func (this *UserController) Get() {
	rep := UserResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	user_id := this.GetSession("user_id")

	if user_id == nil {
		rep.Errno = utils.RECODE_SESSIONERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//根据user_id查询当前用户信息
	user := models.User{Id: user_id.(int)}
	o := orm.NewOrm()
	err := o.Read(&user)

	if err == orm.ErrNoRows {
		//没有此数据
		beego.Debug(err)
		rep.Errno = utils.RECODE_NODATA
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	} else if err == orm.ErrMissPK {
		//找不到主键
		beego.Debug(err)
		rep.Errno = utils.RECODE_NODATA
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	beego.Debug("user id", user.Id, "name:", user.Name)

	//更改user的avatar_url路径 加上服务器前缀
	user.Avatar_url = utils.AddDomain2Url(user.Avatar_url)

	rep.Data = user
	return
}

// /user/auth [GET]
func (this *UserController) AuthGet() {
	// 由于/user/auth [GET] 和 /user 业务相同
	// 所以使用同一个业务
}

// /user/auth [POST]
func (this *UserController) AuthPost() {
	rep := AuthResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	//从session得到user_id
	user_id := this.GetSession("user_id")

	//request post data
	var req AuthInfo
	//得到客户端请求数据
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	if req.RealName == "" || req.Id_card == "" {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = "name is Empty!"
		return
	}

	//更新数据库 User 的 real_name, id_card 字段
	o := orm.NewOrm()
	user := models.User{Id: user_id.(int), Real_name: req.RealName, Id_card: req.Id_card}

	if _, err := o.Update(&user, "real_name", "id_card"); err != nil {
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug(err)
		return
	}

	//更新Session
	this.SetSession("user_id", user_id)

	return
}

// /user/houses [GET]
func (this *UserController) GetHouses() {
	rep := UserHousesResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	user_id := this.GetSession("user_id").(int)

	o := orm.NewOrm()
	qs := o.QueryTable("house")

	house_list := []models.House{}
	//将house相关联的User和Area一并查询
	qs.Filter("user_id", user_id).RelatedSel().All(&house_list)

	houses_rep := Houses{}
	for _, houseinfo := range house_list {
		fmt.Printf("house.user = %+v\n", houseinfo.User)
		fmt.Printf("house.area = %+v\n", houseinfo.Area)
		houses_rep.Houses = append(houses_rep.Houses, houseinfo.To_house_info())
	}
	fmt.Printf("houses_rep = %+v\n", houses_rep)

	rep.Data = houses_rep

	return
}

type OrdersResp struct {
	Errno  string      `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

// /user/orders [GET]
func (this *UserController) GetOrders() {
	rep := OrderResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)
	//得到用户id
	user_id := this.GetSession("user_id").(int)
	//得到用户角色
	var role string
	this.Ctx.Input.Bind(&role, "role")

	if role == "" {
		rep.Errno = utils.RECODE_ROLEERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	o := orm.NewOrm()
	orders := []models.OrderHouse{}
	order_list := []interface{}{}

	if "landlord" == role {
		//角色为房东
		//现找到自己目前已经发布了哪些房子
		landLordHouses := []models.House{}
		o.QueryTable("house").Filter("user__id", user_id).All(&landLordHouses)
		housesIds := []int{}
		for _, house := range landLordHouses {
			housesIds = append(housesIds, house.Id)
		}
		//在从订单中找到房屋id为自己房源的id
		o.QueryTable("order_house").Filter("house__id__in", housesIds).OrderBy("-ctime").All(&orders)
	} else {
		//角色为租客
		o.QueryTable("order_house").Filter("user__id", user_id).OrderBy("-ctime").All(&orders)
	}

	for _, order := range orders {
		o.LoadRelated(&order, "User")
		o.LoadRelated(&order, "House")
		order_list = append(order_list, order.To_order_info())
	}

	data := map[string]interface{}{}
	data["orders"] = order_list

	rep.Data = data

	return
}
