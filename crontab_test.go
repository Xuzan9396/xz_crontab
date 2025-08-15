package xz_crontab_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/Xuzan9396/xz_crontab"
	"github.com/Xuzan9396/zlog"
	"github.com/gorhill/cronexpr"
)

func Test_crontab(t *testing.T) {

	jobs := []xz_crontab.Job{
		{
			Name: "test",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			//CronExpr:  "*/30 * * * * * *", // 5s执行一次
			CronExpr:  "0 55 14 * * * *", // 5s执行一次
			IsOpen:    true,              // true 开启脚本 false 关闭脚本
			Callback:  callback,          // 设置你调用的函数
			ShowNextN: 3,
		},

		//{
		//	Name: "test2",
		//	Par:  "1",
		//	//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
		//	CronExpr: "*/11 * * * * * *", // 5s执行一次
		//	IsOpen:   true,               // true 开启脚本 false 关闭脚本
		//	Callback: callback,           // 设置你调用的函数
		//},
		//
		//{
		//	Name:     "test3",
		//	Par:      "1",
		//	IsOpen:   false,     // true 开启脚本 false 关闭脚本
		//	Callback: OnceTest2, // 设置你调用的函数
		//	Once:     true,      // 只执行一次
		//},
	}
	model := xz_crontab.InitCrontab(jobs)
	go func() {
		for {
			select {
			case nextTime := <-model.NextChGet():
				log.Println(nextTime)
			}
		}
	}()

	time.Sleep(20 * time.Second)
	model.Stop()

	select {}

}

func OnceTest(par ...interface{}) (err error) {
	log.Println("只执行一次")
	return nil
}

func OnceTest2(par ...interface{}) (err error) {
	res := par[2].(context.Context)
	for {
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

func callback(par ...interface{}) (err error) {
	time.Sleep(time.Second)
	log.Println("回调参数", par[0], par[1])
	return
}

func Test_crontabLoc(t *testing.T) {

	jobs := []xz_crontab.Job{
		{
			Name: "test",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			//CronExpr:  "*/30 * * * * * *", // 5s执行一次
			CronExpr:  "0 55 16 * * * *", // 5s执行一次
			IsOpen:    true,              // true 开启脚本 false 关闭脚本
			Callback:  callback,          // 设置你调用的函数
			ShowNextN: 3,
		},
	}
	model := xz_crontab.InitCrontab(jobs, xz_crontab.WithLoc("Asia/Kolkata"))
	go func() {
		for {
			select {
			case nextTime := <-model.NextChGet():
				log.Println(nextTime)
			}
		}
	}()

	time.Sleep(20 * time.Second)
	model.Stop()

	// 给goroutine一点时间来处理Stop信号
	time.Sleep(2 * time.Second)
	log.Println("测试完成")

}

// 解析过期
func Test_crontabParse(t *testing.T) {
	var (
		expr *cronexpr.Expression
		err  error
	)
	if expr, err = cronexpr.Parse("31 46 0 14 3 * 2025"); err != nil {
		fmt.Println(err, "解析错误了")
		return
	}
	nowT := time.Now().In(zlog.LOC)
	nextNow := expr.Next(nowT)
	now := time.Now().In(zlog.LOC)
	if nextNow.Before(now) {
		t.Error("时间过期了")
		//return
	}
	t.Log(nextNow.Unix())

	var nextN []time.Time
	nextN = expr.NextN(now, 3)
	if len(nextN) == 0 {
		t.Error("后面没数据时间过期了")
		//return
	}
	for _, v := range nextN {
		fmt.Println(v)
	}
}

func Test_crontab_loop(t *testing.T) {

	jobs := []xz_crontab.Job{
		{
			Name: "test",
			Par:  "1",
			//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
			//CronExpr:  "*/30 * * * * * *", // 5s执行一次
			IsOpen:   true, // true 开启脚本 false 关闭脚本
			Once:     true, // 只执行一次
			LoopBool: true,
			LoopTime: 2 * time.Second,
			Callback: callback, // 设置你调用的函数
		},

		//{
		//	Name: "test2",
		//	Par:  "1",
		//	//CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
		//	CronExpr: "*/11 * * * * * *", // 5s执行一次
		//	IsOpen:   true,               // true 开启脚本 false 关闭脚本
		//	Callback: callback,           // 设置你调用的函数
		//},
		//
		//{
		//	Name:     "test3",
		//	Par:      "1",
		//	IsOpen:   false,     // true 开启脚本 false 关闭脚本
		//	Callback: OnceTest2, // 设置你调用的函数
		//	Once:     true,      // 只执行一次
		//},
	}
	model := xz_crontab.InitCrontab(jobs)
	go func() {
		for {
			select {
			case nextTime := <-model.NextChGet():
				log.Println(nextTime)
			}
		}
	}()

	time.Sleep(20 * time.Second)
	model.Stop()

	// 给goroutine一点时间来处理Stop信号
	time.Sleep(2 * time.Second)
	log.Println("测试完成")

}
