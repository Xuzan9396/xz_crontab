package xz_crontab

import (
	"context"
	"log"
	"testing"
	"time"
)

func Test_crontab(t *testing.T)  {

	jobs := []Job{
		{
			Name:     "test",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			CronExpr: "*/5 * * * * * *", // 5s执行一次
			IsOpen: false, // true 开启脚本 false 关闭脚本
			Callback: callback,  // 设置你调用的函数
		},

		{
			Name:     "test2",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			CronExpr: "*/11 * * * * * *", // 5s执行一次
			IsOpen: false, // true 开启脚本 false 关闭脚本
			Callback: callback,  // 设置你调用的函数
		},

		{
			Name:     "test3",
			Par:  "1",
			IsOpen: true, // true 开启脚本 false 关闭脚本
			Callback: OnceTest2,  // 设置你调用的函数
			Once: true, // 只执行一次
		},

	}
	model := InitCrontab(jobs)

	time.Sleep(20*time.Second)
	model.Stop()

	select {

	}


}

func OnceTest(par ...interface{})(err error)  {
	log.Println("只执行一次")
	return nil
}


func OnceTest2(par ...interface{})(err error)  {
	res := par[2].(context.Context)
	for{
		select {
		case <-res.Done():
			log.Println("停止了!")
			return
		default:
			log.Println("只执行一次")
			time.Sleep(time.Second)

		}
	}
	return
}


func callback(par ...interface{})(err error )  {
	log.Println("回调参数",par[0],par[1])
	time.Sleep(6*time.Second)
	return
}