package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/facebookgo/pidfile"
	graceful "github.com/pandaychen/tcp-graceful"
)

const (
	LOCAL_ENV_KEY_STRING = "GRACEFUL_TCPSERVER"
)

func SetPidfile() {
	// create pid file
	pidfile.SetPidfilePath("/var/run/tmp.pid")

	err := pidfile.Write()
	if err != nil {
		panic("Could not write pid file")
	}
}

func main() {
	var (
		s   *graceful.TcpServer
		err error
	)

	if os.Getenv(LOCAL_ENV_KEY_STRING) == "true" {
		//read from env
		s, err = graceful.NewTcpServerFromENV(3)
	} else {
		s, err = graceful.NewTcpServer("127.0.0.1", 1234)
	}
	if err != nil {
		fmt.Println("fail to init server:", err)
		return
	}

	var tcpConnHandler = func(conn net.Conn) {
		tick := time.NewTicker(time.Second)
		buffer := make([]byte, 64)
		for {
			select {
			case <-tick.C:
				_, err := conn.Write([]byte("world"))
				if err != nil {
					conn.Close()
					return
				}

				n, err := conn.Read(buffer)
				if err != nil {
					conn.Close()
					return
				}

				fmt.Printf("read from client:%s", n, string(buffer[:n]))
			}
		}
	}

	SetPidfile()

	go s.HandleAccept(tcpConnHandler)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)
	for sig := range signals {
		if sig == syscall.SIGHUP {
			s.StopAccept()
			listenerFD, err := s.GetListenerFD()
			if err != nil {
				fmt.Println("Fail to get socket file descriptor:", err)
				os.Exit(1)
			}
			os.Setenv(LOCAL_ENV_KEY_STRING, "true")
			execSpec := &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFD},
			}
			_, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				os.Exit(1)
			}
			s.WaitAllConnectionsQuit()
			os.Exit(0)
		}
	}
}
