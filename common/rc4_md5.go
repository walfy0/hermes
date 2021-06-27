package common

import (
	"crypto/md5"
	"crypto/rc4"

	"github.com/sirupsen/logrus"
	"github.com/hermes/config"
)

func Rc4_md5(buf []byte) []byte {
	md5sum := md5.Sum([]byte(config.Rc4Md5Key))
	cipher, err := rc4.NewCipher([]byte(md5sum[:])) //定义一个加密器“cipher”,把我们的秘钥穿进去就OK拉！
	if err != nil {
		logrus.Warnln(err)
	}
	cipher.XORKeyStream(buf, buf)
	return buf
}
