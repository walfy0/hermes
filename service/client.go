package service

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"

	"github.com/sirupsen/logrus"
)

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func goClient() {
	defaultConfig := Config{}
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("read config.json err")
	}
	err = json.Unmarshal(data, &defaultConfig)
	if err != nil {
		panic("json parse err")
	}
	flag.StringVar(&defaultConfig.User, "u", "root", "")
	flag.StringVar(&defaultConfig.Password, "u", "", "")
	flag.StringVar(&defaultConfig.Host, "u", "localhost", "")
	flag.IntVar(&defaultConfig.Port, "u", 1080, "")
}

//ss-local
type SsrClient struct {
	*SsrSocket
}

func NewSsrClient() SsrSocketInterface {
	return &SsrClient{
		&SsrSocket{
			addr:       ":1081",
			remoteAddr: "127.0.0.1:1082",
			// remoteAddr: "139.196.100.116:1081",
		},
	}
}

func (c *SsrClient) Listen() error {
	var err error
	c.conn, err = net.Listen("tcp", c.addr)
	if err != nil {
		logrus.Errorf("err: ", err)
		return err
	}
	defer c.conn.Close()
	logrus.Info("start ss-local")
	for {
		client, err := c.conn.Accept()
		if err != nil {
			client.Close()
			return err
		}
		logrus.Info("go ac")
		go c.Handle(client)
	}
}

func (c *SsrClient) Handle(conn net.Conn) {
	remote, err := c.DialRemote()
	if err != nil {
		logrus.Warnf("ssrClient handle err: %v", err)
		return
	}
	defer remote.Close()
	logrus.Info("start remote")
	go func() {
		//ss-local <- |encode| ss-service
		err := c.Recode(conn, remote, false)
		if err != nil {
			logrus.Warnf("ss client recode err:", err)
			remote.Close()
			conn.Close()
		}
	}()
	//ss-local |encode| -> ss-service
	err = c.Recode(remote, conn, true)
	if err != nil {
		logrus.Warnf("ss client recode err:", err)
	}
}
