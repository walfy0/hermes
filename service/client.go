package service

import (
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
)

//ss-local
type SsrClient struct {
	*SsrSocket
}

func NewSsrClient() SsrSocketInterface {
	remoteAddr := ConfigJson.Host + ":" + strconv.Itoa(ConfigJson.Port)
	return &SsrClient{
		&SsrSocket{
			addr:       ":1081",
			remoteAddr: remoteAddr,
		},
	}
}

func (c *SsrClient) Listen() error {
	c.DoSignal()
	var err error
	c.conn, err = net.Listen("tcp", c.addr)
	if err != nil {
		logrus.Errorf("err: ", err)
		return err
	}
	defer c.conn.Close()
	logrus.Info("start ss-local")
	for {
		select {
		case isClose := <-c.flag:
			if isClose {
				logrus.Info("client close listening")
				return c.conn.Close()
			}
		default:
			client, err := c.conn.Accept()
			if err != nil {
				client.Close()
				return err
			}
			logrus.Info("go ac")
			go c.Handle(client)
		}
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
