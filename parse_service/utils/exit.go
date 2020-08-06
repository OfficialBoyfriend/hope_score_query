package utils

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var functions []func() error

func init() {
	c := make(chan os.Signal)
	// 监听信号
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<- c
		for _, function := range functions {
			if err := function(); err != nil {
				log.Println(err)
			}
		}
		os.Exit(0)
	}()
}

func RunOnExit(function func() error) {
	functions = append(functions, function)
}