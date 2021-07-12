package xz_crontab



import (
	"errors"
	"fmt"
	"github.com/gorhill/cronexpr"
	"log"
	"time"
)

// 任务调度计划表
type JobSchedulerPlan struct {
	Job      *Job
	Expr     *cronexpr.Expression // 解析好的cronnxpr 表达式
	NextTime time.Time

}

type Job struct {
	Name     string `json:"name"`     // 任务名
	Par  string `json:"Par"`  // 额外参数
	CronExpr string `json:"cronExpr"` // cron 表达式
	IsOpen   bool
	Callback  func(par ...interface{}) (err error)
}

type Scheduler struct {
	jobPlanTable map[string]*JobSchedulerPlan // 执行计划表
}

func InitCrontab(jobs []Job) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	model := &Scheduler{
		jobPlanTable: make(map[string]*JobSchedulerPlan),
	}

	// 可以从配置项里面获取
	//jobs := []Job{
	//	{
	//		Name:     "split_file",
	//		Par:  "1",
	//		CronExpr: "45 59 23 * * * *", // 23 点 59分 45 秒
	//		IsOpen: true,
	//		//Callback: xxxx  // 设置你调用的函数
	//	},
	//
	//}

	for _, job := range jobs {
		if job.IsOpen {
			model.jobPlanTable[job.Name], _ = BuildSchedulerPlan(job)
		}

	}
	//log.Println(model.jobPlanTable);

	model.SchedulerLoop()
	//return model
}

// 构建任务执行计划
func BuildSchedulerPlan(job Job) (jobSchedulerPlan *JobSchedulerPlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		fmt.Println(err, "解析错误了")
		return
	}
	nextNow := expr.Next(time.Now())
	now := time.Now()
	if nextNow.Before(now) {
		err = errors.New("时间过期了")
		return
	}

	jobSchedulerPlan = &JobSchedulerPlan{
		Job:      &job,
		Expr:     expr,
		NextTime: nextNow,
	}
	return

}

// 调度协程
func (scheduler *Scheduler) SchedulerLoop() {
	// 定时任务
	var (
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
	)

	// 计算调度的时间
	schedulerAfter = scheduler.TrySchedule()
	// 调度延时器
	schedulerTimer = time.NewTimer(schedulerAfter)
	// 调度延迟事件

	for {
		select {
		case <-schedulerTimer.C:

		}

		// 重新调度一次任务
		schedulerAfter = scheduler.TrySchedule()
		// 重置任务定时器
		schedulerTimer.Reset(schedulerAfter)
	}
}

// 尝试遍历所有任务
func (scheduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		//jobPlan  *JobSchedulerPlan
		now      time.Time
		nearTime *time.Time
	)

	// 没有任务睡一s
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}

	now = time.Now()

	for key, jobPlan := range scheduler.jobPlanTable {

		if jobPlan.NextTime.Unix() < 0 {
			// 过期的删除
			log.Println("我删除了", jobPlan.Job.Name)
			delete(scheduler.jobPlanTable, key)
			continue
		}
		timeLayout := "2006-01-02 15:04:05" //转化所需模板
		datetime := time.Unix(jobPlan.NextTime.Unix(), 0).Format(timeLayout)
		log.Println(jobPlan.Job.Name, "下次执行的时间", datetime)
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {


			if jobPlan.Job.Callback == nil {
				continue
			}

			err := jobPlan.Job.Callback(jobPlan.Job.Name,jobPlan.Job.Par)
			if err != nil {
				log.Println(jobPlan.Job.Name,err)
			}


			// 更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)

		}

		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 睡眠多少时间
	schedulerAfter = (*nearTime).Sub(now)

	return
}
