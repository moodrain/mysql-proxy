package main

import (
	"mysql-proxy/lib"
	"net"
	"runtime"
	"sync"
)

func main() {

	mysqlUrl := ":3306"
	listen := ":8888"
	connCount := runtime.NumCPU()

	server, _ := net.Listen("tcp", listen)
	connList := make([]lib.ProxyConn, connCount)
	for i := 0; i < connCount; i++ {
		proxy := lib.ProxyConn{}
		connList[i] = proxy
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	for i := 0; i < connCount; i++ {

		go func(proxy lib.ProxyConn) {

			proxy.NewClientConn(server)
			proxy.NewMysqlConn(mysqlUrl)
			err := proxy.Handshake()
			if err != nil {
				proxy.Close()
			}

			go proxy.PipeMysql2Client()
			go proxy.PipeClient2Mysql()

			for {
				if !proxy.IsClientClose() {
					continue
				}
				proxy.NewClientConn(server)
				err := proxy.FakeHandshake()
				if err != nil {
					proxy.CloseClient()
				}
				go proxy.PipeClient2Mysql()
			}

		}(connList[i])

	}

	wg.Wait()
}
