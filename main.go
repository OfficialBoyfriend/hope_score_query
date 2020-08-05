package main

import (
	"errors"
	"log"
	"score_query_server/utils"
	"sync"
	"time"

	"score_query_server/modules"
	"score_query_server/pkg"
)

var lock *sync.RWMutex

func init() {
	// 初始化读写锁
	lock = new(sync.RWMutex)
	// 日志格式设置
	log.SetFlags(log.Llongfile | log.LstdFlags)
	// 初始化数据库
	modules.Init()
}

func main() {

	highTask := make(chan func() error, 1)
	baseTask := make(chan func() error, 15)
	errNotice := make(chan error)
	parseErrCount := 0
	maxParseErrNum := 5
	parseSleep := time.Second * 15
	t1 := time.NewTicker(time.Second * 30)
	t2 := time.NewTicker(time.Minute * 15)
	t3 := time.NewTicker(parseSleep)

	// 错误处理
	go func() {
		for {
			// 等待消息
			err := <- errNotice
			if err == nil {
				continue
			}

			// 判断是否为解析错误
			if errors.As(err, new(pkg.ParseError)) {
				// 增加错误计数值
				parseErrCount++
				// 判断是否超过限制
				// 超过则设置解析间隔为3分钟
				if parseErrCount >= maxParseErrNum {
					parseSleep = time.Minute * 3
					log.Printf("解析错误次数超过限制（%v次），解析间隔将延长", maxParseErrNum)
				}
				continue
			}

			// 判断是否为解析成功
			if errors.As(err, new(pkg.ParseSuccess)) {
				// 错误计数置零
				parseErrCount = 0
				// 设置解析间隔为15秒
				parseSleep = time.Second * 15
				log.Println("解析间隔恢复15秒")
				continue
			}

			// 一般错误，直接打印至终端
			log.Println(err)
		}
	}()

	// 生成成绩
	go func() {
	for {
		<-t1.C
		log.Printf("生成成绩")
		lock.Lock()
		highTask <- pkg.GeneratedScoreJson
		lock.Unlock()
		t1.Reset(time.Second * 30)
	}
	}()

	// 生成成绩(前九)
	go func() {
		for {
			<- t2.C
			log.Printf("生成成绩(前九)")
			baseTask <- pkg.GeneratedScoreTopJson
			t2.Reset(time.Minute * 15)
		}
	}()

	// 成绩采集
	go func() {
		for {
			<- t3.C
			log.Printf("抓取成绩")
			ids, err := modules.NewId().GetFind(map[string]interface{}{"is_valid": false})
			if err != nil {
				errNotice <- err
				return
			}
			for _, v := range ids {
				// 发送解析任务
				id := v
				baseTask <- func() error {
					log.Println(utils.GetGoId(), "- 抓取成绩:", id.Group)
					// 读写锁
					lock.RLock()
					defer lock.RUnlock()
					// 抓取数据
					if err := pkg.ParseAndWriteScore(id.Group); err != nil {
						return pkg.ParseError{Msg: "抓取数据失败",Err: err}
					}
					return nil
				}
			}
			t3.Reset(parseSleep)
		}
	}()

	// 启动任务
	utils.RunTask(20, time.Second, highTask, baseTask, errNotice)
}
