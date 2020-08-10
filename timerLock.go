package timer

//通过加锁方式实现定时器，所有变更都需要加锁

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type TimerLock struct {
	Id           uint64
	mTask        map[int64]*task
	heap         *SmallHeap
	status       int64
	chCmd        chan *Cmd
	finishCount  int64 //执行完成数量
	runningCount int64 //正在执行的方法数量
	wg           sync.WaitGroup
	lock         sync.Mutex
}

func NewTimerLock() ITimer {
	return &TimerLock{
		Id:     0,
		mTask:  make(map[int64]*task),
		heap:   NewSmallHeap(4),
		status: 0,
		chCmd:  make(chan *Cmd),
	}
}

func (t *TimerLock) nextId() uint64 {
	return atomic.AddUint64(&t.Id, 1)
}

func (t *TimerLock) Add(fun RunFun, interval time.Duration, once bool) int64 {
	tk := &task{
		id:       int64(t.nextId()),
		fun:      fun,
		interval: interval,
		once:     once,
		runTime:  time.Now(),
	}
	t.lock.Lock()
	t.mTask[tk.id] = tk
	t.heap.Push(tk)
	t.lock.Unlock()
	return tk.id
}

func (t *TimerLock) Remove(timerId int64) {
	if atomic.LoadInt64(&t.status) == 0 {
		return
	}
	t.lock.Lock()
	if task := t.mTask[timerId]; task != nil {
		t.heap.Remove(task)
		delete(t.mTask, timerId)
	}
	t.lock.Unlock()
}

func (t *TimerLock) Start() {
	if atomic.LoadInt64(&t.status) == 1 {
		return
	}
	fmt.Println("TimerLock.Start in")
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
			case CmdClose:
				isRun = false
				t.close(task)
			}
		}
	}
	fmt.Println("TimerLock.Start out")
}

func (t *TimerLock) Stop() {
	fmt.Println("TimerLock.Stop in")
	if atomic.LoadInt64(&t.status) == 0 {
		return
	}
	atomic.StoreInt64(&t.status, 0)
	cmd := newTaskCmd(CmdClose, nil)
	t.chCmd <- cmd
	<-cmd.response
	t.wg.Wait()
	fmt.Println("TimerLock.Stop out")
}

func (t *TimerLock) close(taskCmd *Cmd) {
	t.mTask = make(map[int64]*task)
	t.heap = NewSmallHeap(0)
	taskCmd.response <- taskCmd
}

func (t *TimerLock) checkAndRunTask() {
	t.lock.Lock()
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
	t.lock.Unlock()
	for _, task := range list {
		t.run(task)
	}
}

func (t *TimerLock) run(task *task) {
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
	t.lock.Lock()
	if !task.once {
		task.runTime = time.Now().Add(task.interval)
		t.heap.Push(task)
	} else {
		delete(t.mTask, task.id)
	}
	t.lock.Unlock()
}

func (t *TimerLock) GetRunningCount() int64 {
	return atomic.LoadInt64(&t.runningCount)
}

func (t *TimerLock) GetFinishCount() int64 {
	return atomic.LoadInt64(&t.finishCount)
}

func (t *TimerLock) GetTaskCount() int {
	return len(t.mTask)
}
