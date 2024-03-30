package etcd

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

//func getETCDKey(serverID, serverType string) string {
//	return fmt.Sprintf("servers/%s/%s", serverType, serverID)
//}

func getServerID(key string, sep string) (string, error) {
	tmpArr := strings.Split(key, sep)
	if len(tmpArr) == 0 {
		return "", fmt.Errorf("invalid key or sep")
	}
	lastIndex := len(tmpArr) - 1
	return tmpArr[lastIndex], nil
}

func splitHostPort(addr string) (host string, port int, err error) {
	if h, p, e := net.SplitHostPort(addr); e != nil {
		if addr != "nonhost" {
			err = e
		}
		host = "nonhost"
		port = -1
	} else {
		host = h
		port, err = strconv.Atoi(p)
	}
	return
}
