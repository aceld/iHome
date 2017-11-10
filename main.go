package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"iHome/models"
	_ "iHome/routers"
)

func ormInit() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// set default database
	orm.RegisterDataBase("default", "mysql", "root:mysql@tcp(127.0.0.1:3306)/ihome_go?charset=utf8", 30)

	//注册model
	orm.RegisterModel(new(models.User), new(models.House), new(models.Area), new(models.Facility), new(models.HouseImage), new(models.OrderHouse))

	// create table
	//第二个参数是强制更新数据库
	//第三个参数是如果没有则同步
	orm.RunSyncdb("default", false, true)
}

func main() {
	ormInit()
	beego.SetStaticPath("/group1/M00", "fastdfs/storage_data/data")
	//beego.Run(":9998")
	beego.Run()
}
