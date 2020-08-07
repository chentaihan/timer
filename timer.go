package timer

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Cmd int

const (
	CmdAdd    Cmd = 1
	CmdRemove Cmd = 2
	CmdClose  Cmd = 3
)

type RunFun func()

type ITimer interface {
	Add(fun RunFun, interval time.Duration, once bool) int64
	Remove(timerId int64)
}

type task struct {
	id       int64
	fun      RunFun
	interval time.Duration
	once     bool
	runTime  time.Time
}

func (tk *task) GetHashCode() int {
	return int(tk.runTime.UnixNano())
}

func (tk *task) GetValue() int {
	return int(tk.id)
}

type taskCmd struct {
	cmd      Cmd
	data     interface{}
	response chan *taskCmd
}

func newTaskCmd(cmd Cmd, data interface{}) *taskCmd {
	return &taskCmd{
		cmd:      cmd,
		data:     data,
		response: make(chan *taskCmd),
	}
}

type Timer struct {
	Id           uint64
	mTask        map[int64]*task
	heap         *SmallHeap
	status       int64
	chTask       chan *taskCmd
	finishCount  int64 //执行完成数量
	runningCount int64 //正在执行的方法数量
}

func NewTimer() *Timer {
	return &Timer{
		Id:     0,
		mTask:  make(map[int64]*task),
		heap:   NewSmallHeap(4),
		status: 0,
		chTask: make(chan *taskCmd),
	}
}

func (t *Timer) nextId() uint64 {
	return atomic.AddUint64(&t.Id, 1)
}

func (t *Timer) Add(fun RunFun, interval time.Duration, once bool) int64 {
	tk := &task{
		id:       int64(t.nextId()),
		fun:      fun,
		interval: interval,
		once:     once,
	}
	cmd := newTaskCmd(CmdAdd, tk)
	t.chTask <- cmd
	<-cmd.response
	return tk.id
}

func (t *Timer) Remove(timerId int64) {
	if atomic.LoadInt64(&t.status) == 0 {
		return
	}
	cmd := newTaskCmd(CmdRemove, timerId)
	t.chTask <- cmd
	<-cmd.response
}

func (t *Timer) Start() {
	if atomic.LoadInt64(&t.status) == 1 {
		return
	}
	fmt.Println("Timer.Start in")
	atomic.StoreInt64(&t.status, 1)
	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()
	isRun := true
	for isRun {
		select {
		case <-ticker.C:
			t.checkAndRunTask()
		case task := <-t.chTask:
			switch task.cmd {
			case CmdAdd:
				t.addTask(task)
			case CmdRemove:
				t.removeTask(task)
			case CmdClose:
				isRun = false
				t.close(task)
			}
		}
	}
	fmt.Println("Timer.Start out")
}

func (t *Timer) Stop() {
	if atomic.LoadInt64(&t.status) == 0 {
		return
	}
	atomic.StoreInt64(&t.status, 0)
	cmd := newTaskCmd(CmdClose, nil)
	t.chTask <- cmd
	<-cmd.response
}

func (t *Timer) close(taskCmd *taskCmd) {
	t.mTask = make(map[int64]*task)
	t.heap = NewSmallHeap(0)
	taskCmd.response <- taskCmd
}

func (t *Timer) addTask(taskCmd *taskCmd) {
	task, _ := taskCmd.data.(*task)
	task.runTime = time.Now()
	t.mTask[task.id] = task
	t.heap.Push(task)
	taskCmd.response <- taskCmd
}

func (t *Timer) removeTask(taskCmd *taskCmd) {
	taskId, _ := taskCmd.data.(int64)
	if task := t.mTask[taskId]; task != nil {
		t.heap.Remove(task)
		delete(t.mTask, taskId)
	}
	taskCmd.response <- taskCmd
}

func (t *Timer) checkAndRunTask() {
	var list []*task
	curTime := time.Now()
	for t.heap.Len() > 0 {
		task := t.heap.Peek().(*task)
		if curTime.Sub(task.runTime) >= 0 {
			list = append(list, task)
			t.heap.Pop()
		} else {
			break
		}
	}
	for _, task := range list {
		t.run(task)
	}
}

func (t *Timer) run(task *task) {
	go func() {
		atomic.AddInt64(&t.runningCount, 1)
		defer func() {
			atomic.AddInt64(&t.finishCount, 1)
			atomic.AddInt64(&t.runningCount, -1)
		}()
		task.fun()
	}()
	//map 和 heap保存的是同一个task，当是once task时需要重map中删除，非once任务map不需要动，重新将task加入heap
	if !task.once {
		task.runTime = time.Now().Add(task.interval)
		t.heap.Push(task)
	} else {
		delete(t.mTask, task.id)
	}
}

func (t *Timer) GetRunningCount() int64 {
	return atomic.LoadInt64(&t.runningCount)
}

func (t *Timer) GetFinishCount() int64 {
	return atomic.LoadInt64(&t.finishCount)
}

func (t *Timer) GetTaskCount() int {
	return len(t.mTask)
}
