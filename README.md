# xz_crontab
定时脚本任务

## 使用

1. 设置脚本定时时间

crontab 规则
https://github.com/gorhill/cronexpr
```
Field name     Mandatory?   Allowed values    Allowed special characters
----------     ----------   --------------    --------------------------
Seconds        No           0-59              * / , -
Minutes        Yes          0-59              * / , -
Hours          Yes          0-23              * / , -
Day of month   Yes          1-31              * / , - L W
Month          Yes          1-12 or JAN-DEC   * / , -
Day of week    Yes          0-6 or SUN-SAT    * / , - L #
Year           No           1970–2099         * / , -
```

2. 例子：
```go
package main

import (
	"github.com/Xuzan9396/xz_crontab"
	"log"
)

func main()  {
	jobs := []xz_crontab.Job{
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
	xz_crontab.InitCrontab(jobs)
}


func callback(par ...interface{})(err error ) {
	log.Println("回调参数", par[0], par[1])
	return
}
```