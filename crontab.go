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
	Name     string      // 任务名
	Par  string   // 额外参数
	CronExpr string // cron 表达式
	IsOpen   bool
	IsSkip bool  // 如果为true 忽视重复
	Callback  func(par ...interface{}) (err error)
}

// 执行的结果
type JobResult struct {
	Name string 
	Err          error // 错误
	StartTing    time.Time
	EndTime      time.Time
}


type Scheduler struct {
	jobPlanTable map[string]*JobSchedulerPlan // 执行计划表
}


var g_jobexecuting map[string]string
var g_JobResult_chan chan *JobResult

func InitCrontab(jobs []Job) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	g_jobexecuting = make(map[string]string)
	g_JobResult_chan = make(chan *JobResult,100)

	model := &Scheduler{
		jobPlanTable: make(map[string]*JobSchedulerPlan),
	}



	for _, job := range jobs {
		if job.IsOpen {
			model.jobPlanTable[job.Name], _ = BuildSchedulerPlan(job)
		}

	}

	model.SchedulerLoop()
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
		case result := <-g_JobResult_chan:
			dealResult(result)
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

			// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！

			if jobPlan.Job.IsSkip == false{
				// 如果任务正在执行，跳过本次调度
				if _, jobExecuting := g_jobexecuting[jobPlan.Job.Name]; jobExecuting {
					log.Printf("尚未退出,跳过执行:%s", jobPlan.Job.Name)
					continue
				}
			}

			// 保存执行状态
			g_jobexecuting[jobPlan.Job.Name] = jobPlan.Job.Name


			if jobPlan.Job.Callback == nil {
				continue
			}
			go func() {
				defer func() {
						if r := recover(); r != nil {
							log.Println(errors.New("灾难错误"), r)
						}
					}()
			    startTing :=   time.Now()
				err := jobPlan.Job.Callback(jobPlan.Job.Name,jobPlan.Job.Par,datetime)
				if err != nil {
					log.Println(jobPlan.Job.Name,err)
				}
				endTime :=     time.Now()

				pushg_JobResult_chan(jobPlan.Job.Name,startTing,endTime,err)
				
			}()


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

func pushg_JobResult_chan(name string,startTime,endTime time.Time,err error )  {
	g_JobResult_chan <- &JobResult{
		Name: name,
		StartTing: startTime,
		EndTime: endTime,
		Err: err,
	}
}

func dealResult(result *JobResult)  {
	delete(g_jobexecuting,result.Name)
	log.Println("执行时间",result.EndTime.Unix() - result.EndTime.Unix())
}