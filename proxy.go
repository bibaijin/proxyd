package main

import (
	"net"
	"sync"

	"github.com/laincloud/proxyd/log"
)

const (
	// MaxConnectionNum 表示最大连接数
	MaxConnectionNum = 60000
)

func accept(done <-chan struct{}, ln net.Listener) <-chan net.Conn {
	conns := make(chan net.Conn, MaxConnectionNum)
	go func() {
		defer close(conns)
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Errorf("ln.Accept() failed, error: %s.", err)
			} else {
				conns <- conn
			}

			select {
			case <-done:
				log.Infof("accept() done.")
				return
			default:
				log.Infof("Ready for another connection...")
			}
		}
	}()
	return conns
}

func handleConns(conns <-chan net.Conn, watcher *Watcher) {
	var wg sync.WaitGroup
	for conn := range conns {
		wg.Add(1)

		go func(conn net.Conn) {
			defer wg.Done()
			_handleConn(conn, watcher)
		}(conn)
	}

	wg.Wait()
	log.Infof("handleConns() done.")
}

func _handleConn(conn net.Conn, watcher *Watcher) {
	if err := conn.Close(); err != nil {
		log.Errorf("conn.Close() failed, error: %s.", err)
	}

	upstream, err := watcher.Upstream()
	if err != nil {
		log.Errorf("watcher.Upstream() failed, error: %s.", err)
		return
	}

	log.Infof("watcher.Upstream(), upstream: %s.", upstream)

	upstreamConn, err := net.Dial("tcp", upstream)
	if err != nil {
		log.Errorf("net.Dial() failed, error: %s.", err)
		return
	}

	pipe(conn, upstreamConn)
}
