package lib

import (
	"fmt"
	"net"
	"strconv"
)

type ProxyConn struct {
	MysqlConn  *net.TCPConn
	ClientConn *net.TCPConn

	InitHandshakePacket   Packet
	FinishHandshakePacket Packet

	ID          int
	clientClose bool
}

func (p *ProxyConn) NewMysqlConn(url string) {
	mysqlConn, _ := net.Dial("tcp", url)
	p.MysqlConn = mysqlConn.(*net.TCPConn)
	p.MysqlConn.SetNoDelay(true)
	p.MysqlConn.SetKeepAlive(true)
}

func (p *ProxyConn) NewClientConn(server net.Listener) {
	clientConn, _ := server.Accept()
	p.ClientConn = clientConn.(*net.TCPConn)
	p.ClientConn.SetNoDelay(true)
	p.clientClose = false
}

func (p ProxyConn) IsClientClose() bool {
	return p.clientClose
}

func (p *ProxyConn) CloseClient() {
	p.clientClose = true
	p.ClientConn.Close()
	fmt.Println("Connection " + strconv.Itoa(p.ID) + " 's client close")
}

func (p *ProxyConn) Close() {
	p.clientClose = true
	fmt.Println("Connection " + strconv.Itoa(p.ID) + " close")
	p.ClientConn.Close()
	p.MysqlConn.Close()
}

func (p ProxyConn) ReadMysql() (Packet, error) {
	packet, err := ReadPacket(p.MysqlConn)
	return packet, err
}

func (p ProxyConn) ReadClient() (Packet, error) {
	packet, err := ReadPacket(p.ClientConn)
	return packet, err
}

func (p ProxyConn) SendMysql(packet Packet) error {
	_, err := p.MysqlConn.Write(packet.Raw())
	return err
}

func (p ProxyConn) SendClient(packet Packet) error {
	_, err := p.ClientConn.Write(packet.Raw())
	return err
}

func (p *ProxyConn) Handshake() error {
	packet, err := p.ReadMysql()
	if err != nil {
		return err
	}
	if err = p.SendClient(packet); err != nil {
		return err
	}
	p.InitHandshakePacket = packet
	packet, err = p.ReadClient()
	if err != nil {
		return err
	}
	if err = p.SendMysql(packet); err != nil {
		return err
	}
	packet, err = p.ReadMysql()
	if err != nil {
		return err
	}
	if err = p.SendClient(packet); err != nil {
		return err
	}
	p.FinishHandshakePacket = packet
	return nil
}

func (p ProxyConn) FakeHandshake() error {
	if err := p.SendClient(p.InitHandshakePacket); err != nil {
		return err
	}
	if _, err := p.ReadClient(); err != nil {
		return err
	}
	if err := p.SendClient(p.FinishHandshakePacket); err != nil {
		return err
	}
	return nil
}

func (p *ProxyConn) PipeClient2Mysql() {
	for {
		packet, err := p.ReadClient()
		if err != nil {
			p.CloseClient()
			break
		}
		if len(packet.Data()) == 1 && packet.Data()[0] == 1 {
			// ignore client close command
		} else {
			p.SendMysql(packet)
		}
	}
}

func (p *ProxyConn) PipeMysql2Client() {
	for {
		packet, err := p.ReadMysql()
		if err != nil {
			p.Close()
			break
		}
		p.SendClient(packet)
	}
}
