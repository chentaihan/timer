# timer

# 定时器管理器

## 实现原理
* 1：按照时间排序维护一个最小堆，堆顶元素就是最近要执行的任务
* 2：每100毫秒检查一次，是否有到期的任务，启动子协程执行所有到期的任务
* 3：维护一个任务id和任务对应关系的map，可通过任务id删除任务
* 4：开始执行一个任务时，会将多次执行的任务更新任务的执行时间，出堆之后再次添加到堆中，等待下次执行，如果是只执行一次的任务，出堆之后不再添加到堆中

## 实现方式
* 方案1：通过channel的方式实现定时器，所有的动作都是发一个消息，在同一个协程中处理这些消息，无需加锁
* 方案2：通过加锁方式实现定时器，所有变更都需要加锁

## 性能对比
在处理20000个定时任务时：方案2的性能是方案1的5倍，而且在短时间处理大量定时器的时候，方案1会崩溃：Panic: too many concurrent operations on a single file or socket。
通过加锁的方式比使用channel的方式性能更好。加锁多协程局部加锁，而channel是单协程，性能自然慢一些 


## 定时器接口说明
```
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
```

## 实例
```
方案一：通过channel的方式实现定时器，性能差
tm := NewTimer()
go tm.Start()
for i := 0; i < 100; i++ {
    timerId := tm.Add(func() {
        fmt.Println(i)
    }, time.Second, false)
    fmt.Println("timerId=", timerId)
}
time.Sleep(time.Second * 3)
fmt.Println("finishCount: ", tm.GetFinishCount())
fmt.Println("runningCount: ", tm.GetRunningCount())
fmt.Println("taskCount: ", tm.GetTaskCount())
tm.Remove(10)
tm.Stop()
fmt.Println("taskCount: ", tm.GetTaskCount())

方案二：通过加锁的方式实现定时器，性能高
tm := NewTimerLock()
go tm.Start()
for i := 0; i < 100; i++ {
    timerId := tm.Add(func() {
        fmt.Println(i)
    }, time.Second, false)
    fmt.Println("timerId=", timerId)
}
time.Sleep(time.Second * 3)
fmt.Println("finishCount: ", tm.GetFinishCount())
fmt.Println("runningCount: ", tm.GetRunningCount())
fmt.Println("taskCount: ", tm.GetTaskCount())
tm.Remove(10)
tm.Stop()
fmt.Println("taskCount: ", tm.GetTaskCount())

```