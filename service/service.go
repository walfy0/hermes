package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/hermes/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SsrService struct {
	*SsrSocket
}

func NewSsrService() SsrSocketInterface {
	return &SsrService{
		&SsrSocket{
			addr:     ":" + strconv.Itoa(ConfigJson.Port),
			user:     ConfigJson.User,
			password: ConfigJson.Password,
		},
	}
}

func (s *SsrService) Listen() error {
	s.DoSignal()
	var err error
	s.conn, err = net.Listen("tcp", s.addr)
	if err != nil {
		logrus.Errorf("listen err: ", err)
		return err
	}
	defer s.conn.Close()
	logrus.Info("start ss-service")
	for {
		select {
		case isClose := <-s.flag:
			if isClose {
				logrus.Info("service close listening")
				return s.conn.Close()
			}
		default:
			client, err := s.conn.Accept()
			if err != nil {
				client.Close()
				return err
			}
			go s.Handle(client)
		}
	}
}

func (s *SsrService) Handle(conn net.Conn) {
	if err := s.Socks5Auth(conn); err != nil {
		logrus.Warnf("SsrService Handle auth err: ", err)
		conn.Close()
		return
	}
	service, err := s.Socks5Connect(conn)
	if err != nil {
		logrus.Warnf("SsrService Handle connect err: ", err)
		conn.Close()
		return
	}
	logrus.Info("start socks5")
	go func() {
		//ss-local <- |encode| ss-service
		defer service.Close()
		defer conn.Close()
		err = s.Recode(conn, service, false)
		if err != nil {
			logrus.Warnf("ss service recode err:", err)
		}
	}()
	go func() {
		//ss-local |encode| -> ss-service
		defer service.Close()
		defer conn.Close()
		err = s.Recode(service, conn, true)
		if err != nil {
			logrus.Warnf("ss service recode err:", err)
		}
	}()
}

func (s *SsrService) Socks5Auth(client net.Conn) error {
	buf := make([]byte, 1024)
	n, err := s.Read(client, buf)
	if n < 2 || err != nil {
		logrus.Warnf("auth err: ", err)
		return errors.New("client auth err")
	}
	ver, methodCount := int(buf[0]), int(buf[1])
	if ver != 5 || n != methodCount+2 {
		logrus.Info(ver, n)
		logrus.Warnf("auth invalid version")
		return errors.New("invalid version")
	}
	//no auth service
	// n, err = client.Write([]byte{0x05, 0x00})
	// if n != 2 || err != nil {
	// 	return errors.New("write error")
	// }
	//auth with password
	n, err = s.Write(client, []byte{0x05, 0x02})
	if n != 2 || err != nil {
		logrus.Warnf("auth client write err: ", err)
		return errors.New("client write err")
	}
	n, err = s.Read(client, buf)
	if n <= 2 || err != nil {
		logrus.Warnf("auth methods err:", err)
		return errors.New("methods err")
	}
	authVersion, userLen := buf[0], int(buf[1])
	user := string(buf[2 : userLen+2])
	passwordLen := int(buf[userLen+2])
	password := string(buf[userLen+3 : userLen+passwordLen+3])
	logrus.Infof("user: %v, password: %v", user, password)
	if string(user) == s.user && string(password) == s.password {
		_, err = s.Write(client, []byte{authVersion, 0x00})
		if err != nil {
			logrus.Warnf("auth methods err: %v", err)
			return errors.New("methods err")
		}
	} else {
		_, _ = s.Write(client, []byte{authVersion, 0x01})
		return errors.New("password err")
	}
	return nil
}

func (s *SsrService) Socks5Connect(client net.Conn) (net.Conn, error) {
	buf := make([]byte, 1024)
	n, err := s.Read(client, buf)
	//ver cmd rsv atyp dst.addr dst.port
	// 1   1  0x00  1    var        2
	if n < 6 || err != nil {
		logrus.Info("connect read err: ", err, n)
		return nil, errors.New("connect read err")
	}
	ver, command, _, addType := int(buf[0]), int(buf[1]), 0, int(buf[3])
	if ver != 5 || command != 1 {
		return nil, errors.New("invalid version")
	}
	addr := ""
	port := uint16(0)
	switch addType {
	case 1:
		//ipv4
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		port = binary.BigEndian.Uint16(buf[8:10])
	case 3:
		//domain
		len := int(buf[4])
		addr = string(buf[4 : len+4])
		port = binary.BigEndian.Uint16(buf[len+4 : len+6])
	case 4:
		//ipv6
		addr = fmt.Sprintf("%d.%d.%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7], buf[8], buf[9])
		port = binary.BigEndian.Uint16(buf[10:12])
	}
	//service net->sshClient
	dest, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		logrus.Info("connect write err", err)
		return nil, err
	}
	logrus.Println(addr, port)
	logrus.Println(dest.LocalAddr().String())
	_, err = s.Write(client, []byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return nil, errors.New("write error" + err.Error())
	}
	return dest, nil
}

var sshClient *ssh.Client

func (s *SsrService) Gao() {
	b, err := ioutil.ReadFile("/home/walfy/.ssh/id_rsa")
	if err != nil {
		logrus.Panic(err)
		panic(err)
	}
	pKey, err := ssh.ParsePrivateKey(b)
	if err != nil {
		logrus.Panic(err)
	}
	sshClient, err = ssh.Dial("tcp", config.Vps, &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(pKey)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Infof("ssh success!")
	defer sshClient.Close()
}
