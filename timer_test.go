package timer

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTimer(t *testing.T) {
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
}

func TestNewTimerLock(t *testing.T) {
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
}

func TestTimerCompare(t *testing.T) {
	tm := NewTimer()
	go tm.Start()
	const count = 20000
	startTime := time.Now()
	for i := 0; i < count; i++ {
		tm.Add(func() {
			fmt.Println(i)
		}, 0, false)
	}
	tm.Stop()
	endTime := time.Now()
	useTime := endTime.Sub(startTime)
	t.Logf("timer useTime=%v", useTime)

	tm = NewTimerLock()
	go tm.Start()
	startTime = time.Now()
	for i := 0; i < count; i++ {
		tm.Add(func() {
			fmt.Println(i)
		}, 0, false)
	}
	tm.Stop()
	endTime = time.Now()
	useTime = endTime.Sub(startTime)
	t.Logf("timerLock useTime=%v", useTime)
}
