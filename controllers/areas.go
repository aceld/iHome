/**
* @file areas.go
* @brief  区域请求controller 目前只支持北京市内区域
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
	"time"
)

type AreasResp struct {
	Errno  string      `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

type AreaController struct {
	beego.Controller
}

func (this *AreaController) Get() {
	rep := AreasResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&rep)

	//1 从redis读地域信息,如果有直接返回
	areas_info_key := "area_info"
	redis_config_map := map[string]string{
		"key":   "ihome_go",
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)

	cache_conn, err := cache.NewCache("redis", string(redis_config))
	if err != nil {
		beego.Debug("connect cache error")
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	areas_info_value := cache_conn.Get(areas_info_key)
	if areas_info_value != nil {
		beego.Debug("=== get AreaInfo from CACHE!===")
		var areas_info interface{}
		json.Unmarshal(areas_info_value.([]byte), &areas_info)
		rep.Data = areas_info
		return
	}

	//2 如果没有应该从mysql中查到,然后拷贝到Redis中
	o := orm.NewOrm()

	var areas []models.Area
	num, err := o.QueryTable("area").All(&areas)
	if err != nil {
		//数据库出错
		beego.Debug(err)
		rep.Errno = utils.RECODE_DATAERR
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}
	if num == 0 {
		//查询数据为0行
		rep.Errno = utils.RECODE_NODATA
		rep.Errmsg = utils.RecodeText(rep.Errno)
		return
	}

	//3 将areas数据返回前端
	area_list := []models.Area{}
	for _, area := range areas {
		fmt.Printf("%+v\n", area)
		area_list = append(area_list, area)
	}
	//将区域信息存入缓存
	areas_info_value, _ = json.Marshal(area_list)
	cache_conn.Put(areas_info_key, areas_info_value, 3600*time.Second)

	rep.Data = area_list
	return
}

func (this *AreaController) RetData(rep interface{}) {
	this.Data["json"] = rep
	this.ServeJSON()
}
