package graceful

import "sync"

type TcpConnManager struct {
	sync.WaitGroup
	counter int64
}

func NewTcpConnManager() *TcpConnManager {
	m := &TcpConnManager{}
	return m
}

func (m *TcpConnManager) Add(num int64) {
	m.counter += num
	m.WaitGroup.Add(int(num))
}

func (m *TcpConnManager) Done() {
	m.counter--
	m.WaitGroup.Done()
}

func (m *TcpConnManager) NoAliveConn() bool {
	return m.counter == 0
}
