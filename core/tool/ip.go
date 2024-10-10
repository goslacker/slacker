package tool

import (
	"fmt"
	"net"
	"strings"
)

func IPByTarget(targetAddr string) (ip string, err error) {
	// 使用udp发起网络连接, 这样不需要关注连接是否可通, 随便填一个即可
	conn, err := net.Dial("udp", targetAddr)
	if err != nil {
		err = fmt.Errorf("get SelfIP to target<%s> failed: %w", targetAddr, err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func IP() (ip string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		err = fmt.Errorf("get interfaces failed: %w", err)
		return
	}

	for _, iface := range interfaces {
		addrs, e := iface.Addrs()
		if e != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				return
			}
		}
	}

	err = fmt.Errorf("no valid address found")
	return
}
