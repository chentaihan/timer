package timer

/**
定时器管理器
 */
import "time"

type ITimer interface {
	//添加定时任务
	//fun：要执行的方法
	//interval：执行时间间隔
	//once：true，只执行一次，false：执行多次，每interval执行一次
	//return 任务id，可通过这个id重定时器中将这个任务移除
	Add(fun RunFun, interval time.Duration, once bool) int64

	//通过任务id将任务从定时器中移除
	//timerId为Add方法的返回值
	Remove(timerId int64)

	//启动定时器，只需调用一次，如果Stop后，需要在调用Start重新启动定时器
	Start()

	//停止定时器
	//会清空所有还没执行的任务
	//当所有任务执行完成后才会返回，适合在进程准备推出的时候调用
	Stop()

	//返回正在执行的任务数量
	GetRunningCount() int64

	//返回已经执行成功的任务数量
	GetFinishCount() int64

	//返回还没开始执行的任务数量
	GetTaskCount() int
}
