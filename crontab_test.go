package xz_crontab

import (
	"log"
	"testing"
)

func Test_crontab(t *testing.T)  {

	jobs := []Job{
		{
			Name:     "脚本名字1",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			CronExpr: "*/5 * * * * * *", // 5s执行一次
			IsOpen: true, // true 开启脚本 false 关闭脚本
			Callback: callback,  // 设置你调用的函数
		},

		{
			Name:     "脚本名字2",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			CronExpr: "*/5 * * * * * *", // 5s执行一次
			IsOpen: true, // true 开启脚本 false 关闭脚本
			Callback: callback,  // 设置你调用的函数
		},

	}
	InitCrontab(jobs)
}

func callback(par ...interface{})(err error )  {
	log.Println("回调参数",par[0],par[1])
	return
}