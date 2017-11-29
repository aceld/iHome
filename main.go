package main

import (
	"github.com/astaxie/beego"
	_ "iHome/models" //加载orm
	_ "iHome/routers"
	_ "iHome/utils" //加载配置文件
)

func main() {
	beego.SetStaticPath("/group1/M00", "fastdfs/storage_data/data")
	//beego.Run(":9998")
	beego.Run()
}
