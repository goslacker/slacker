package tool

import (
	"fmt"
	"net"
	"strings"
)

func SelfIP(target string) (ip string, err error) {
	// 使用udp发起网络连接, 这样不需要关注连接是否可通, 随便填一个即可
	conn, err := net.Dial("udp", target)
	if err != nil {
		err = fmt.Errorf("get SelfIP to target<%s> failed: %w", target, err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
