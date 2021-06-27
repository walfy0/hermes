package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hermes/service"
	"github.com/sirupsen/logrus"
)

func Init() {
	now := time.Now()
	fileName := fmt.Sprintf("./output/log/%d-%02d-%02d_%02d", now.Year(), now.Month(), now.Day(), now.Hour())
	write2, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(io.MultiWriter(write2))
	logrus.SetReportCaller(true)
}

func main() {
	Init()
	// service.StartDaemon()
	s := service.NewSsrService()
	s.Listen()
}
