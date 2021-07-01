package tcp

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/network/conn"
	"net"
	"runtime"
	"time"
)

func Dial(addr string, ctx *lokas.Context) (*conn.Conn, error) {
	c, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn := conn.NewTcpConn(c, ctx, nil)
	conn.ServeIO()
	return conn, nil
}

func DialEx(addr string, ctx *lokas.Context, retryWait time.Duration) chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		retryChan := make(chan bool, 1)
		retryChan <- true
		needWait := false
		for {
			select {
			case <-doneChan:
				return
			case <-retryChan:
				if needWait && retryWait > 0 {
					select {
					case <-time.NewTimer(retryWait).C:
					case <-doneChan:
						return
					}
				} else {
					needWait = true
				}
				c, err := net.DialTimeout("tcp", addr, 5*time.Second)
				if err != nil {
					log.Debugf("connect error: %s", err.Error())
					retryChan <- true
					continue
				}
				conn := conn.NewTcpConn(c, ctx, nil)
				conn.ServeIO()
				conn.Wait() // this will block the for loop, the application layer need to close the conn to let it go along
				runtime.Gosched()
				retryChan <- true
			}
		}
	}()
	return doneChan
}
