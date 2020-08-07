package timer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type CmdType int

const (
	CmdAdd    CmdType = 1
	CmdRemove CmdType = 2
	CmdClose  CmdType = 3
)

type RunFun func()

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

type Cmd struct {
	cmd      CmdType
	data     interface{}
	response chan *Cmd
}

func newTaskCmd(cmd CmdType, data interface{}) *Cmd {
	return &Cmd{
		cmd:      cmd,
		data:     data,
		response: make(chan *Cmd),
	}
}

type Timer struct {
	Id           uint64
	mTask        map[int64]*task
	heap         *SmallHeap
	status       int64
	chCmd        chan *Cmd
	finishCount  int64 //执行完成数量
	runningCount int64 //正在执行的方法数量
	wg           sync.WaitGroup
}

func NewTimer() ITimer {
	return &Timer{
		Id:     0,
		mTask:  make(map[int64]*task),
		heap:   NewSmallHeap(4),
		status: 0,
		chCmd:  make(chan *Cmd),
	}
}

func (t *Timer) nextId() uint64 {
	return atomic.AddUint64(&t.Id, 1)
}

func (t *Timer) Add(fun RunFun, interval time.Duration, once bool) int64 {
	if interval < 0 {
		interval = 0
	}
	tk := &task{
		id:       int64(t.nextId()),
		fun:      fun,
		interval: interval,
		once:     once,
	}
	cmd := newTaskCmd(CmdAdd, tk)
	t.chCmd <- cmd
	<-cmd.response
	return tk.id
}

func (t *Timer) Remove(timerId int64) {
	if atomic.LoadInt64(&t.status) == 0 {
		return
	}
	cmd := newTaskCmd(CmdRemove, timerId)
	t.chCmd <- cmd
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
		case task := <-t.chCmd:
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
	fmt.Println("Timer.Stop in")
	if atomic.LoadInt64(&t.status) == 0 {
		return
	}
	atomic.StoreInt64(&t.status, 0)
	cmd := newTaskCmd(CmdClose, nil)
	t.chCmd <- cmd
	<-cmd.response
	t.wg.Wait()
	fmt.Println("Timer.Stop out")
}

func (t *Timer) close(taskCmd *Cmd) {
	t.mTask = make(map[int64]*task)
	t.heap = NewSmallHeap(0)
	taskCmd.response <- taskCmd
}

func (t *Timer) addTask(taskCmd *Cmd) {
	task, _ := taskCmd.data.(*task)
	task.runTime = time.Now()
	t.mTask[task.id] = task
	t.heap.Push(task)
	taskCmd.response <- taskCmd
}

func (t *Timer) removeTask(taskCmd *Cmd) {
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
		t.wg.Add(1)
		atomic.AddInt64(&t.runningCount, 1)
		defer func() {
			atomic.AddInt64(&t.finishCount, 1)
			atomic.AddInt64(&t.runningCount, -1)
			t.wg.Done()
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
