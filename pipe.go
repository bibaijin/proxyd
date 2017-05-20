package main

import (
	"io"
	"net"

	"github.com/laincloud/proxyd/log"
)

// pipe 传输数据
func pipe(clientConn net.Conn, upstreamConn net.Conn) <-chan net.Conn {
	toCloseConns := make(chan net.Conn, 2)
	go func() {
		defer close(toCloseConns)
		clientAddr := clientConn.RemoteAddr()
		upstreamAddr := upstreamConn.RemoteAddr()
		log.Infof("pipe: %s <--> %s start...", clientAddr, upstreamAddr)

		done := make(chan struct{}, 2)

		go func() {
			_copy(upstreamConn, clientConn)
			done <- struct{}{}
		}()

		go func() {
			_copy(clientConn, upstreamConn)
			done <- struct{}{}
		}()

		<-done

		toCloseConns <- clientConn
		toCloseConns <- upstreamConn

		<-done

		log.Infof("pipe: %s <--> %s done.", clientAddr, upstreamAddr)
	}()
	return toCloseConns
}

func _copy(dst, src net.Conn) {
	srcAddr := src.RemoteAddr()
	dstAddr := dst.RemoteAddr()
	log.Infof("copy: %s --> %s start...", srcAddr, dstAddr)

	n, err := io.Copy(dst, src)
	if err != nil {
		log.Errorf("io.Copy() failed, error: %s.", err)
	}

	log.Infof("copy: %s --> %s done, written: %d.", srcAddr, dstAddr, n)
}
