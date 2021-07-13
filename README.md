## crontab
**1 crontab 规则，设置脚本定时时间**

[我的博客地址](https://blog.csdn.net/qq_36517296/article/details/118692303)

```go
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

例如
- 格式 秒 分 小时 天 月 星期(0代表星期天) 年
- */10 * * * * * *   每10s执行一次
- 20 30 3 * * * *  每天凌晨3点30分20s执行
- 15 0 0 * * 1 *   每周1凌晨0点0分15s执行
- 5 0 0,6,10,12,18 16-28 6 * 2021    2021年6月份 16号到28号(0,6,10,12,18点)0分5秒执行

**2 示例代码**

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
	// par[0] 代表脚本名称， par[1] 代表Par额外参数
	log.Println("回调参数", par[0], par[1])
	return
}
```

**3.运行结果**
![在这里插入图片描述](https://img-blog.csdnimg.cn/20210713093339409.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzM2NTE3Mjk2,size_16,color_FFFFFF,t_70)
