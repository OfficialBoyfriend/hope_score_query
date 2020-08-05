package utils

import "time"

// RunTask 运行任务
func RunTask(maxTaskNum uint, sleepTime time.Duration, highChan, baseChan <- chan func() error, errChan chan <- error) {

	// 任务计数
	taskNum := uint(0)

	run := func(function func() error) {
		taskNum++
		errChan <- function()
		taskNum--
	}

	for {
		// 协程数量控制
		if taskNum > maxTaskNum {
			time.Sleep(time.Millisecond * 100)
		}

		select {
		case function := <-highChan: // 高优先级
			// 执行任务
			go run(function)
		default: 					 // 正常优先级
			// 检测是否有任务
			function, ok := <-baseChan
			if !ok {
				break
			}
			// 执行任务
			go run(function)
		}

		// 等待时间
		if sleepTime != 0 {
			time.Sleep(sleepTime)
		}
	}
}
