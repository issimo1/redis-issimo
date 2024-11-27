package tcpserver

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/issimo1/redis-issimo/utils/logger"
)

type TCPConfig struct {
	Addr string
}

// TCPServer ..
type TCPServer struct {
	listener      net.Listener
	waitDone      sync.WaitGroup
	clientCounter int64
	conf          TCPConfig
	closeTcp      int32
	quit          chan os.Signal

	//待补充
	redisHandler string
}

func NewTCPServer(conf TCPConfig, handler string) *TCPServer {
	return &TCPServer{
		conf:          conf,
		closeTcp:      0,
		clientCounter: 0,
		quit:          make(chan os.Signal, 1),
		redisHandler:  handler,
	}
}

func (t *TCPServer) Start() error {
	listen, err := net.Listen("tcp", t.conf.Addr)
	if err != nil {
		return err
	}
	t.listener = listen
	logger.Infof("bind %s listening...", t.conf.Addr)
	go t.accept()

	signal.Notify(t.quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-t.quit
	return nil
}

func (t *TCPServer) accept() error {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				logger.Infof("accept occurs temporary error: %v, retry in 5ms", err)
				time.Sleep(5 * time.Millisecond)
				continue
			}
			logger.Warn(err.Error())
			atomic.CompareAndSwapInt32(&t.closeTcp, 0, 1)
			t.quit <- syscall.SIGTERM
			break
		}
		go t.handleConn(conn)
	}
	return nil
}

func (t *TCPServer) handleConn(conn net.Conn) {
	if atomic.LoadInt32(&t.closeTcp) == 1 {
		conn.Close()
		return
	}
	logger.Debugf("accept new conn %s", conn.RemoteAddr().String())
	t.waitDone.Add(1)
	atomic.AddInt64(&t.clientCounter, 1)
	defer func() {
		t.waitDone.Done()
		atomic.AddInt64(&t.clientCounter, -1)
	}()

	t.redisHandler = "11"
}

func (t *TCPServer) Close() {
	logger.Info("graceful shutdown issimo1 redis server")
	atomic.CompareAndSwapInt32(&t.closeTcp, 0, 1)
	t.listener.Close()
	t.redisHandler = ""
	t.waitDone.Wait()
}
