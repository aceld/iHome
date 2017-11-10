package utils

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
)

/* 将url加上 http://IP:PROT/  前缀 */
func AddDomain2Url(url string) (domain_url string) {
	//从配置文件读取到port和addr
	appconf, err := config.NewConfig("ini", "./conf/app.conf")
	if err != nil {
		beego.Debug(err)
		return ""
	}
	port := appconf.String("httpport")
	addr := appconf.String("httpaddr")

	//beego.Debug("port:", port, " addr:", addr)
	domain_url = "http://" + addr + ":" + port + "/" + url

	return domain_url
}
