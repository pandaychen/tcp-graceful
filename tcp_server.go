package graceful

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
)

type HandleTcpConns func(net.Conn)

const (
	DEFAULT_SOCKET_FILE = "/tmp/tcp-graceful.sock"
)

type TcpServer struct {
	addr        string
	pid, port   int
	listenSock  *net.TCPListener
	connManager *TcpConnManager
}

// NewTcpServer ...
func NewTcpServer(addr string, port int) (*TcpServer, error) {
	var (
		err    error
		server *TcpServer
	)

	server = &TcpServer{
		addr:        addr,
		port:        port,
		connManager: NewTcpConnManager(),
	}

	raddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr error: %v", err)
	}
	server.listenSock, err = net.ListenTCP("tcp", raddr)
	if err != nil {
		return nil, fmt.Errorf("ListenTCP error: %v", err)
	}

	return server, nil
}

// NewTcpServerFromENV ...
func NewTcpServerFromENV(old_fd uintptr) (*TcpServer, error) {
	var (
		err      error
		server   *TcpServer
		file     *os.File
		ok       bool
		listener net.Listener
	)

	server = &TcpServer{
		connManager: NewTcpConnManager(),
	}

	file = os.NewFile(old_fd, DEFAULT_SOCKET_FILE)
	if file == nil {
		return nil, errors.New("bad file fd")
	}

	if listener, err = net.FileListener(file); err != nil {
		return nil, err
	}

	server.listenSock, ok = listener.(*net.TCPListener)
	if !ok {
		return nil, errors.New("bad tcp socket")
	}

	return server, nil
}

// handleAccept ...
func (s *TcpServer) HandleAccept(handleTcpConn HandleTcpConns) {
	for {
		conn, err := s.listenSock.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				//match s.Stop()
				return
			}
		}
		go func() {
			s.connManager.Add(1)
			handleTcpConn(conn)
			s.connManager.Done()
		}()
	}
}

// getListenerFD ...
func (s *TcpServer) GetListenerFD() (uintptr, error) {
	file, err := s.listenSock.File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}

//stopAccept ...
func (s *TcpServer) StopAccept() {
	s.listenSock.SetDeadline(time.Now())
}

func (s *TcpServer) WaitAllConnectionsQuit() {
	s.connManager.Wait()
}
