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
