package main

import (
	"io"
	"net"
)

// pipe 传输数据
func pipe(clientConn net.Conn, upstreamConn net.Conn, toCloseConns chan<- net.Conn) {
	clientAddr := clientConn.RemoteAddr()
	upstreamAddr := upstreamConn.RemoteAddr()
	infoLogger.Printf("pipe: %s <--> %s start...", clientAddr, upstreamAddr)

	done := make(chan struct{}, 2)

	go copy(upstreamConn, clientConn, done)
	go copy(clientConn, upstreamConn, done)

	<-done

	toCloseConns <- clientConn
	toCloseConns <- upstreamConn

	<-done

	infoLogger.Printf("pipe: %s <--> %s done.", clientAddr, upstreamAddr)
}

func copy(dst, src net.Conn, done chan<- struct{}) {
	srcAddr := src.RemoteAddr()
	dstAddr := dst.RemoteAddr()
	infoLogger.Printf("copy: %s --> %s start...", srcAddr, dstAddr)

	n, err := io.Copy(dst, src)
	if err != nil {
		errLogger.Printf("io.Copy failed, error: %s.", err)
	}

	infoLogger.Printf("copy: %s --> %s done, written: %d.", srcAddr, dstAddr, n)
	done <- struct{}{}
}
