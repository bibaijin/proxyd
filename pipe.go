package main

import (
	"io"
	"net"

	"github.com/laincloud/proxyd/log"
)

// pipe 传输数据
func pipe(clientConn net.Conn, upstreamConn net.Conn) {
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

	if err := clientConn.Close(); err != nil {
		log.Errorf("clientConn.Close() failed, error: %s.", err)
	}

	if err := upstreamConn.Close(); err != nil {
		log.Errorf("upstreamConn.Close() failed, error: %s.", err)
	}

	<-done

	log.Infof("pipe: %s <--> %s done.", clientAddr, upstreamAddr)
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
