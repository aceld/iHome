/**
* @file houses.go
* @brief  房屋model相关controller 房屋的查询,上传,图片的上传,查询等
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
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"github.com/astaxie/beego/orm"
	"iHome/models"
	"iHome/utils"
	"path"
	"strconv"
	"time"
)

type House_id struct {
	House_id int64 `json:"house_id"`
}

// /houses 请求的返回数据
type HousesResp struct {
	Errno  string   `json:"errno"`
	Errmsg string   `json:"errmsg"`
	Data   House_id `json:"data"`
}

type HouseInfo struct {
	Area_id    string   `json:"area_id"`    //归属地的区域编号
	Title      string   `json:"title"`      //房屋标题
	Price      string   `json:"price"`      //单价,单位:分
	Address    string   `json:"address"`    //地址
	Room_count string   `json:"room_count"` //房间数目
	Acreage    string   `json:"acreage"`    //房屋总面积
	Unit       string   `json:"unit"`       //房屋单元,如 几室几厅
	Capacity   string   `json:"capacity"`   //房屋容纳的总人数
	Beds       string   `json:"beds"`       //房屋床铺的配置
	Deposit    string   `json:"deposit"`    //押金
	Min_days   string   `json:"min_days"`   //最好入住的天数
	Max_days   string   `json:"max_days"`   //最多入住的天数 0表示不限制
	Facilities []string `json:"facility"`   //房屋设施
}

type Image_url struct {
	Url string `json:"url"`
}

// /houses/:id/images 请求的回复数据
type HouseImageResp struct {
	Errno  string    `json:"errno"`
	Errmsg string    `json:"errmsg"`
	Data   Image_url `json:"data"`
}

type HouseOneResp struct {
	Errno  string      `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

type HousesController struct {
	beego.Controller
}

func (this *HousesController) RetData(rep interface{}) {
	this.Data["json"] = rep
	this.ServeJSON()
}

// /houese?aid=1&sd=2017-11-09&ed=2017-11-11&sk=new&p=1 [GET]
func (this *HousesController) Get() {
	rep := HouseOneResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	beego.Debug()
	var aid int
	this.Ctx.Input.Bind(&aid, "aid")
	var sd string
	this.Ctx.Input.Bind(&sd, "sd")
	var ed string
	this.Ctx.Input.Bind(&ed, "ed")
	var sk string
	this.Ctx.Input.Bind(&sk, "sk")
	var page int
	this.Ctx.Input.Bind(&page, "p")

	beego.Debug(aid, sd, ed, sk, page)

	//把时间从str转换成字符串格式

	//校验开始时间一定要早于结束时间

	//判断page的合法性 一定是大于0的整数

	//尝试从redis中获取数据, 有则返回

	//如果没有 从mysql中查询
	houses := []models.House{}

	o := orm.NewOrm()

	qs := o.QueryTable("house")

	num, err := qs.Filter("area_id", aid).All(&houses)
	if err != nil {
		rep.Errno = utils.RECODE_PARAMERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	total_page := int(num)/models.HOUSE_LIST_PAGE_CAPACITY + 1
	house_page := 1

	house_list := []interface{}{}
	for _, house := range houses {
		o.LoadRelated(&house, "Area")
		o.LoadRelated(&house, "User")
		o.LoadRelated(&house, "Images")
		o.LoadRelated(&house, "Facilities")
		house_list = append(house_list, house.To_house_info())
	}

	data := map[string]interface{}{}
	data["houses"] = house_list
	data["total_page"] = total_page
	data["current_page"] = house_page

	rep.Data = data

	return
}

// /houese [POST]
func (this *HousesController) Post() {
	rep := HousesResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	var req HouseInfo
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug(err)
		return
	}

	//校验数据合法性
	fmt.Println("%+v\n", req)

	house := models.House{}

	house.Room_count, _ = strconv.Atoi(req.Room_count)
	house.Title = req.Title
	house.Acreage, _ = strconv.Atoi(req.Acreage)
	house.Unit = req.Unit
	house.Deposit, _ = strconv.Atoi(req.Deposit)
	house.Deposit = house.Deposit * 100 //单位转换
	house.Address = req.Address
	house.Price, _ = strconv.Atoi(req.Price)
	house.Price = house.Price * 100 //单位转换
	house.Capacity, _ = strconv.Atoi(req.Capacity)
	house.Beds = req.Beds
	house.Min_days, _ = strconv.Atoi(req.Min_days)
	house.Max_days, _ = strconv.Atoi(req.Max_days)
	user := models.User{Id: this.GetSession("user_id").(int)}
	area_id, _ := strconv.Atoi(req.Area_id)
	area := models.Area{Id: area_id}
	house.User = &user
	house.Area = &area

	o := orm.NewOrm()
	//将单个house插入到house表中
	house_id, err := o.Insert(&house)
	if err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	beego.Debug("house insert id =", house_id, " succ!")

	//多对多 m2m插入,将facilities 一起关联插入到表中
	facilities := []*models.Facility{}
	for _, fid := range req.Facilities {
		id, _ := strconv.Atoi(fid)
		facility := &models.Facility{Id: id}
		facilities = append(facilities, facility)
	}

	// 第一个参数的对象，主键必须有值
	// 第二个参数为对象需要操作的 M2M 字段
	// QueryM2Mer 的 api 将作用于 Id 为 1 的 House
	m2mhouse_facility := o.QueryM2M(&house, "Facilities")

	num, err := m2mhouse_facility.Add(facilities)
	if err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	beego.Debug("house m2m facility insert num =", num, " succ!")
	rep.Data = House_id{House_id: house_id}

	return
}

// /houses/:id/images [POST]
func (this *HousesController) UploadHouseImage() {
	rep := HouseImageResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	house_id := this.Ctx.Input.Param(":id")

	file, header, err := this.GetFile("house_image")
	if err != nil {
		rep.Errno = utils.RECODE_REQERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("get file error")
		return
	}
	defer file.Close()

	//将文件存到fastDFS上
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

	beego.Debug("house_id", house_id)
	house := models.House{}
	house.Id, _ = strconv.Atoi(house_id)

	o := orm.NewOrm()
	if err := o.Read(&house); err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("fdfs upload file error")
		return
	}
	//根据house_id 查询house_image 是否为空
	if house.Index_image_url == "" {
		//如果为空 那么就用当前image_url为house的主image_url
		house.Index_image_url = fileId
		beego.Debug("set index_image_url ", fileId)
	}

	house_image := models.HouseImage{House: &house, Url: fileId}

	//将house_image 和hosue相关联
	house.Images = append(house.Images, &house_image)

	//将house_image入库
	if _, err := o.Insert(&house_image); err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("insert house image error")
		return

	}

	//将house更新入库
	if _, err := o.Update(&house); err != nil {
		rep.Errno = utils.RECODE_DBERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		beego.Debug("update house error")
		return
	}

	//返回前端图片url
	image_url := Image_url{Url: utils.AddDomain2Url(fileId)}
	rep.Data = image_url
}

// /houses/:id [GET]
func (this *HousesController) GetOneHouseInfo() {
	rep := HouseOneResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)
	data := make(map[string]interface{})

	user_id := this.GetSession("user_id")
	beego.Debug("user_id = ", user_id)
	if user_id == nil {
		user_id = -1
	}
	house_id := this.Ctx.Input.Param(":id")

	//先从缓存中获取房屋数据,将缓存数据返回前端即可
	cache_conn, err := cache.NewCache("redis", `{"key": "ihome_go", "conn":"127.0.0.1:6380", "dbNum":"8"}`)
	if err != nil {
		beego.Debug("connect cache error")
	}
	house_info_key := fmt.Sprintf("house_info_%s", house_id)
	house_info_value := cache_conn.Get(house_info_key)
	if house_info_value != nil {
		beego.Debug("======= get house info desc  from CACHE!!! ========")
		data["user_id"] = user_id
		house_info := map[string]interface{}{}
		json.Unmarshal(house_info_value.([]byte), &house_info)
		data["house"] = house_info
		rep.Data = data
		return
	}

	beego.Debug("======= no house info desc CACHE!!!  SAVE house desc to CACHE !========")
	//如果缓存没有房屋数据,那么从数据库中获取数据,再存入缓存中,然后返回给前端
	//1 根据house_id 关联查询数据库
	o := orm.NewOrm()

	//根据house_id查询house
	/*
		// ---- 方法1  关联关系查询 ----
			if err := o.QueryTable("house").Filter("id", house_id).RelatedSel().One(&house); err != nil {
				rep.Errno = utils.RECODE_NODATA
				rep.Errmsg = utils.RecodeText(rep.Errno)
				return
			}
			//查询house_id为 id的house_image
			if _, err := o.QueryTable("house_image").Filter("House", house_id).RelatedSel().All(&house.Images); err != nil {

			}

			//查询house_id为 id的Facility
			if _, err := o.QueryTable("facility").Filter("Houses__House__Id", house_id).All(&house.Facilities); err != nil {

			}
	*/
	// --- 方法2  载入关系查询 -----
	house := models.House{}
	house.Id, _ = strconv.Atoi(house_id)
	o.Read(&house)
	o.LoadRelated(&house, "Area")
	o.LoadRelated(&house, "User")
	o.LoadRelated(&house, "Images")
	o.LoadRelated(&house, "Facilities")

	//2 将该房屋的json格式数据保存在redis缓存数据库
	house_info_value, _ = json.Marshal(house.To_one_house_desc())
	cache_conn.Put(house_info_key, house_info_value, 3600*time.Second)

	//3 返回数据
	data["user_id"] = user_id
	data["house"] = house.To_one_house_desc()

	rep.Data = data
	return
}

type HouseIndexResp struct {
	Errno  string      `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

// /houses/index [GET]
func (this *HousesController) IndexHouses() {
	rep := HouseIndexResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)
	data := []interface{}{}

	beego.Debug("Index Houses....")

	//1 从缓存服务器中请求 "home_page_data" 字段,如果有值就直接返回
	//先从缓存中获取房屋数据,将缓存数据返回前端即可
	cache_conn, err := cache.NewCache("redis", `{"key": "ihome_go", "conn":"127.0.0.1:6380", "dbNum":"8"}`)
	if err != nil {
		beego.Debug("connect cache error")
	}
	house_page_key := "home_page_data"
	house_page_value := cache_conn.Get(house_page_key)
	if house_page_value != nil {
		beego.Debug("======= get house page info  from CACHE!!! ========")
		json.Unmarshal(house_page_value.([]byte), &data)
		rep.Data = data
		return
	}

	houses := []models.House{}

	//2 如果缓存没有,需要从数据库中查询到房屋列表
	o := orm.NewOrm()

	if _, err := o.QueryTable("house").Limit(models.HOME_PAGE_MAX_HOUSES).All(&houses); err == nil {
		for _, house := range houses {
			o.LoadRelated(&house, "Area")
			o.LoadRelated(&house, "User")
			o.LoadRelated(&house, "Images")
			o.LoadRelated(&house, "Facilities")
			data = append(data, house.To_house_info())
		}

	}

	//将data存入缓存数据
	house_page_value, _ = json.Marshal(data)
	cache_conn.Put(house_page_key, house_page_value, 3600*time.Second)

	//返回前端data
	rep.Data = data
	return
}
