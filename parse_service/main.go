package main

import (
	"log"

	micro "github.com/micro/go-micro/v2"

	proto "score_query_server/parse_service/proto"
)

func init() {
	// 日志格式设置
	log.SetFlags(log.Llongfile | log.LstdFlags)
}

func main() {
	// 创建新的服务，这里可以传入其它选项。
	service := micro.NewService(
		micro.Name("greeter"),
	)

	// 初始化方法会解析命令行标识
	service.Init()

	// 注册处理器
	proto.RegisterParseHandler(service.Server(), new(Greeter))

	// 运行服务
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}