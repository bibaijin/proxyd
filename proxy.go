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
			}
		}
	}()
	return conns
}

func handleConns(conns <-chan net.Conn, watcher *Watcher) {
	toCloseConns := make(chan net.Conn, 2*MaxConnectionNum)
	go func() {
		defer close(toCloseConns)
		var wg sync.WaitGroup
		for conn := range conns {
			wg.Add(1)

			go func(conn net.Conn) {
				defer wg.Done()
				cs := _handleConn(conn, watcher)
				for c := range cs {
					toCloseConns <- c
				}
			}(conn)
		}

		wg.Wait()
		log.Infof("handleConns() done.")
	}()

	for conn := range toCloseConns {
		if err := conn.Close(); err != nil {
			log.Errorf("conn.Close() failed, error: %s.", err)
		}
	}
}

func _handleConn(conn net.Conn, watcher *Watcher) <-chan net.Conn {
	toCloseConns := make(chan net.Conn, 2)
	go func() {
		defer close(toCloseConns)
		upstream, err := watcher.Upstream()
		if err != nil {
			log.Errorf("watcher.Upstream() failed, error: %s.", err)
			toCloseConns <- conn
			return
		}

		log.Infof("watcher.Upstream(), upstream: %s.", upstream)

		upstreamConn, err := net.Dial("tcp", upstream)
		if err != nil {
			log.Errorf("net.Dial() failed, error: %s.", err)
			toCloseConns <- conn
			return
		}

		cs := pipe(conn, upstreamConn)
		for c := range cs {
			toCloseConns <- c
		}
	}()
	return toCloseConns
}
