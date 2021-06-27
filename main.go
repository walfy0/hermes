package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hermes/service"
	"github.com/sirupsen/logrus"
)

func init() {
	// write1 := os.Stdout
	now := time.Now()
	fileName := fmt.Sprintf("./output/log/%d-%02d-%02d: %02d", now.Year(), now.Month(), now.Day(), now.Hour())
	write2, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic("init error")
	}
	logrus.SetOutput(io.MultiWriter(write2))
	logrus.SetReportCaller(true)
}

func main() {
	service.StartDaemon()
	s := service.NewSsrClient()
	s.Listen()
}
