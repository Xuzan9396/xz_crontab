package xz_crontab

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorhill/cronexpr"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// 任务调度计划表
type JobSchedulerPlan struct {
	Job      *Job
	Expr     *cronexpr.Expression // 解析好的cronnxpr 表达式
	NextTime time.Time
}

type Job struct {
	Name     string // 任务名
	Par      string // 额外参数
	CronExpr string // cron 表达式
	IsOpen   bool
	IsSkip   bool // 如果为true 忽视重复 false 默认只会开启一个
	Callback func(par ...interface{}) (err error)
	Once     bool // true 常驻只执行一次
}

// 执行的结果
type JobResult struct {
	Name      string
	Err       error // 错误
	StartTing time.Time
	EndTime   time.Time
}

type Scheduler struct {
	jobPlanTable     map[string]*JobSchedulerPlan // 执行计划表
	jobPlanTableInit map[string]*Job              // 只会执行一次的脚本
	is_stop          bool
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	nextCh chan string
}

var g_jobexecuting map[string]string
var g_JobResult_chan chan *JobResult

func init() {
	//log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func InitCrontab(jobs []Job) *Scheduler {

	g_jobexecuting = make(map[string]string)
	g_JobResult_chan = make(chan *JobResult, 100)
	model := &Scheduler{
		jobPlanTable:     make(map[string]*JobSchedulerPlan),
		jobPlanTableInit: make(map[string]*Job),
		is_stop:          false,
		nextCh:           make(chan string),
	}

	model.ctx, model.cancel = context.WithCancel(context.Background())

	for _, job := range jobs {
		if job.IsOpen {
			if job.Once == false {
				model.jobPlanTable[job.Name], _ = BuildSchedulerPlan(job)
			} else {
				model.once(job)
			}
		}

	}

	go model.SchedulerLoop()
	return model
}
func (c *Scheduler) NextChGet() chan string {
	return c.nextCh
}

func (c *Scheduler) once(job Job) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(errors.New("灾难错误"), r, string(debug.Stack()))
			}
		}()
		job.Callback(job.Name, job.Par, c.ctx)
	}()
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
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
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

		select {
		case scheduler.nextCh <- fmt.Sprintf("%s,下次执行的时间:%s", jobPlan.Job.Name, datetime):
		default:

		}
		//log.Println(jobPlan.Job.Name, "下次执行的时间", datetime)
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {

			if scheduler.getStop() {
				//log.Println("脚本停止了，请检查数据是否跑完!")
				goto LOOP
			}
			// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！
			if jobPlan.Job.IsSkip == false {
				// 如果任务正在执行，跳过本次调度
				if _, jobExecuting := g_jobexecuting[jobPlan.Job.Name]; jobExecuting {
					log.Printf("尚未退出,跳过执行:%s", jobPlan.Job.Name)
					goto LOOP
				}
			}
			if jobPlan.Job.Callback == nil {
				goto LOOP
			}
			// 保存执行状态
			g_jobexecuting[jobPlan.Job.Name] = jobPlan.Job.Name

			go func(jobPlan *JobSchedulerPlan) { // go 需要传值进来
				defer func() {
					if r := recover(); r != nil {
						log.Println(errors.New("灾难错误"), r, string(debug.Stack()))

					}
				}()

				startTing := time.Now()
				err := jobPlan.Job.Callback(jobPlan.Job.Name, jobPlan.Job.Par, datetime)
				if err != nil {
					log.Println(jobPlan.Job.Name, err)
				}
				endTime := time.Now()

				pushg_JobResult_chan(jobPlan.Job.Name, startTing, endTime, err)

			}(jobPlan)
		LOOP:

			// 更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)

		}

		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 睡眠多少时间
	schedulerAfter = (*nearTime).Sub(now)
	//log.Println("schedulerAfter",schedulerAfter)
	return
}

func pushg_JobResult_chan(name string, startTime, endTime time.Time, err error) {
	g_JobResult_chan <- &JobResult{
		Name:      name,
		StartTing: startTime,
		EndTime:   endTime,
		Err:       err,
	}
}

func dealResult(result *JobResult) {
	delete(g_jobexecuting, result.Name)
	//log.Println("执行时间删除",result.EndTime.Unix() - result.EndTime.Unix(),result.Name)
}

// 关闭脚本
func (scheduler *Scheduler) getStop() bool {
	scheduler.RLock()
	defer scheduler.RUnlock()
	return scheduler.is_stop
}

// 关闭脚本
func (scheduler *Scheduler) Stop() {
	scheduler.Lock()
	defer scheduler.Unlock()
	scheduler.is_stop = true
	// 取消一次性脚本
	scheduler.cancel()
}

func (scheduler *Scheduler) Start() {
	scheduler.Lock()
	defer scheduler.Unlock()
	scheduler.is_stop = false
}
