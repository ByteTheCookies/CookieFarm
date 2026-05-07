package ckp

import (
	"net"
)

type Connection interface {
	net.Conn
	GetNetConn() net.Conn
	GetServer() *Server
	GetClientAddr() *net.TCPAddr
	GetServerAddr() *net.TCPAddr

	Reset(netConn net.Conn)
	SetServer(server *Server)
}

type TCPConn struct {
	net.Conn
	server            *Server
	ts                int64    //nolint
	_cacheLinePadding [24]byte //nolint
}

func (conn *TCPConn) GetClientAddr() *net.TCPAddr {
	return conn.RemoteAddr().(*net.TCPAddr)
}

func (conn *TCPConn) GetServerAddr() *net.TCPAddr {
	return conn.LocalAddr().(*net.TCPAddr)
}

func (conn *TCPConn) GetNetConn() net.Conn {
	return conn.Conn
}

func (conn *TCPConn) GetNetTCPConn() (c *net.TCPConn) {
	c, _ = conn.Conn.(*net.TCPConn)
	return c
}

func (conn *TCPConn) SetServer(s *Server) {
	conn.server = s
}

func (conn *TCPConn) GetServer() *Server {
	return conn.server
}

func (conn *TCPConn) Reset(netConn net.Conn) {
	conn.Conn = netConn
}
