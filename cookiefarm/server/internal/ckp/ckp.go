package ckp

import (
	"context"
	"errors"
	"fmt"
	"logger"
	"net"
	"pool"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	soReusePort = 0x0F
	tcpFastOpen = 0x17
)

type Server struct {
	listenAddr          *net.TCPAddr
	listener            *net.TCPListener
	shutdown            atomic.Bool
	shutdownDeadline    time.Time
	requestHandler      RequestHandlerFunc
	connectionCreator   ConnectionCreatorFunc
	activeConnections   int32
	acceptedConnections int32
	listenConfig        *ListenConfig
	connWaitGroup       sync.WaitGroup
	connStructPool      sync.Pool
	wp                  *pool.WorkerPool[*net.TCPConn]
	allowThreadLocking  bool
	ballast             []byte
	conns               *Connections
}

type ListenConfig struct {
	lc                     net.ListenConfig
	SocketReusePort        bool
	SocketFastOpen         bool
	SocketFastOpenQueueLen int
	SocketDeferAccept      bool
}

type RequestHandlerFunc func(conn Connection)

type ConnectionCreatorFunc func() Connection

type controlFunc func(network, address string, c syscall.RawConn) error

func applyListenSocketOptions(lc *ListenConfig) controlFunc {
	return func(network, address string, c syscall.RawConn) error {
		var err error

		c.Control(func(fd uintptr) {
			if lc.SocketReusePort {
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, soReusePort, 1)
			}
			if lc.SocketFastOpen {
				qlen := lc.SocketFastOpenQueueLen
				if qlen <= 0 {
					qlen = 256
				}
				err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, tcpFastOpen, qlen)
			}
			if lc.SocketDeferAccept {
				err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_DEFER_ACCEPT, 1)
			}
		})
		return err
	}
}

var defaultListenConfig *ListenConfig = &ListenConfig{
	SocketReusePort: true,
}

func NewServer(listenAddr string) (*Server, error) {
	la, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		listenAddr:   la,
		listenConfig: defaultListenConfig,
	}

	s.connStructPool = sync.Pool{
		New: func() any {
			conn := s.connectionCreator()
			conn.SetServer(s)
			return conn
		},
	}
	s.connectionCreator = func() Connection { return &TCPConn{} }

	s.SetBallast(20)

	return s, nil
}

func (s *Server) SetListenConfig(config *ListenConfig) {
	s.listenConfig = config
}

func (s *Server) GetListenConfig() *ListenConfig {
	return s.listenConfig
}

func (s *Server) Listen(ctx context.Context) error {
	logger.Log.Info().Str("addr", s.listenAddr.String()).Msg("Starting CKP server...")
	network := "tcp4"
	if isIPv6Addr(s.listenAddr) {
		network = "tcp6"
	}

	s.listenConfig.lc.Control = applyListenSocketOptions(s.listenConfig)
	l, err := s.listenConfig.lc.Listen(ctx, network, s.listenAddr.String())
	if err != nil {
		return err
	}

	tcpl, ok := l.(*net.TCPListener)
	if !ok {
		return errors.New("listener must be of type net.TCPListener")
	}

	s.listener = tcpl
	return nil
}

func (s *Server) GetActiveConnections() int32 {
	return s.activeConnections
}

func (s *Server) GetAcceptedConnections() int32 {
	return s.acceptedConnections
}

func (s *Server) GetListenAddr() *net.TCPAddr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr().(*net.TCPAddr)
}

func (s *Server) Shutdown(d time.Duration) (err error) {
	s.shutdownDeadline = time.Time{}
	if d > 0 {
		s.shutdownDeadline = time.Now().Add(d)
	}
	s.shutdown.Store(true)
	err = s.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Halt() (err error) {
	return s.Shutdown(-1 * time.Second)
}

func (s *Server) Serve() error {
	maxProcs := runtime.GOMAXPROCS(0)

	s.wp = pool.NewWorkerPool(s.serveConn)
	s.wp.SetNumShards(maxProcs * 2)
	s.wp.SetIdleWorkerLifetime(5 * time.Second)
	s.wp.Start()
	defer s.wp.Stop()

	for range maxProcs {
		go s.acceptLoop()
	}

	if s.shutdownDeadline.IsZero() {
		s.connWaitGroup.Wait()
	} else {
		diff := time.Until(s.shutdownDeadline)

		if diff > 0 {
			time.Sleep(diff)
		}
	}

	return nil
}

func (s *Server) SetRequestHandler(requestHandler RequestHandlerFunc) {
	s.requestHandler = requestHandler
}

func (s *Server) SetConnectionCreator(f ConnectionCreatorFunc) {
	s.connectionCreator = f
}

func (s *Server) SetAllowThreadLocking(allow bool) {
	s.allowThreadLocking = allow
}

func (s *Server) SetBallast(sizeInMiB int) {
	s.ballast = make([]byte, sizeInMiB*1024*1024)
}

func (s *Server) acceptLoop() error {
	for {
		if s.shutdown.Load() {
			_ = s.listener.Close()
			break
		}

		tcpConn, err := s.listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if (!opErr.Temporary() || !opErr.Timeout()) && s.shutdown.Load() {
					break
				}
			}

			s.listener.Close()
			return err
		}

		newConn := s.connectionCreator()
		if newConn == nil {
			tcpConn.Close()
			logger.Log.Warn().Str("remote", tcpConn.RemoteAddr().String()).Msg("Connection creator returned nil; closing connection")
			continue
		}
		if tcpC, ok := newConn.(*TCPConn); ok {
			tcpC.SetServer(s)
			tcpC.Reset(tcpConn)
			s.conns.Add(tcpC)
		} else {
			logger.Log.Warn().Str("remote", tcpConn.RemoteAddr().String()).Msg("Connection creator did not return *TCPConn; skipping add to conns")
		}

		atomic.AddInt32(&s.acceptedConnections, 1)

		s.wp.AddTask(tcpConn)
	}

	return nil
}

func (s *Server) serveConn(netConn *net.TCPConn) {
	s.connWaitGroup.Add(1)
	defer s.connWaitGroup.Done()
	conn := s.connStructPool.Get().(*TCPConn)

	atomic.AddInt32(&s.activeConnections, 1)

	conn.Reset(netConn)
	s.requestHandler(conn)
	atomic.AddInt32(&s.activeConnections, -1)

	s.conns.Remove(conn)
	s.connStructPool.Put(conn)
}

func isIPv6Addr(addr *net.TCPAddr) bool {
	return addr.IP.To4() == nil && len(addr.IP) == net.IPv6len
}

func StartServer(port uint16) (*Connections, error) {
	s, err := NewServer(fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s.SetListenConfig(&ListenConfig{
		SocketReusePort:   true,
		SocketFastOpen:    false,
		SocketDeferAccept: false,
	})

	s.SetRequestHandler(handler)
	s.SetAllowThreadLocking(true)

	s.conns = &Connections{}

	err = s.Listen(context.Background())
	if err != nil {
		return nil, err
	}

	go s.Serve()

	return s.conns, nil
}
