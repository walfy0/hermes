package service

import (
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hermes/common"
	"github.com/sirupsen/logrus"
)

type SsrSocketInterface interface {
	Listen() error
	Handle(conn net.Conn)
}

type SsrSocket struct {
	addr       string
	remoteAddr string
	user       string
	password   string
	conn       net.Listener
	flag	   chan bool
}

func (s *SsrSocket) Listen() error        { return nil }
func (s *SsrSocket) Handle(conn net.Conn) {}

func (s *SsrSocket) DoSignal() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	flag := make(chan bool, 1)

	go func() {
		logrus.Info("start listen signal!")
		for c := range sigs {
			switch c {
			case syscall.SIGINT, syscall.SIGTERM:
				// ctrl+c / kill pid is restart!
				// restart after closing the connection for 10 second
				logrus.Info("about to restart!")
				flag <- true
				logrus.Info("restarting!")
				time.Sleep(time.Second * 10)
				s.Listen()
				logrus.Info("restart success!")
			}
		}
	}()
}

func (s *SsrSocket) DialRemote() (net.Conn, error) {
	conn, err := net.Dial("tcp", s.remoteAddr)
	logrus.Info("dial remote")
	if err != nil {
		logrus.Infof("link to remote service err: %v", err)
		return nil, err
	}
	return conn, err
}

func (s *SsrSocket) Recode(dst io.Writer, src io.Reader, isEncode bool) error {
	size := 32 * 1024
	buf := make([]byte, size)
	for {
		nr, err := src.Read(buf[0:size])
		if nr > 0 {
			if isEncode {
				s.Encode(buf[0:nr])
			} else {
				s.Decode(buf[0:nr])
			}
			dst.Write(buf[0:nr])
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

func (s *SsrSocket) Encode(buf []byte) {
	common.Rc4_md5(buf)
}

func (s *SsrSocket) Write(client net.Conn, buf []byte) (int, error) {
	common.Rc4_md5(buf)
	return client.Write(buf)
}

func (s *SsrSocket) Decode(buf []byte) {
	common.Rc4_md5(buf)
}

func (s *SsrSocket) Read(client io.Reader, buf []byte) (int, error) {
	n, err := client.Read(buf)
	if err != io.EOF && err != nil {
		logrus.Warnf("read decode err:", err)
		return n, err
	}
	s.Decode(buf[0:n])
	return n, nil
}

// read all buf !!!!!!!!!!!!!!!!
