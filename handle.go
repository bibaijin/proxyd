package main

import (
	"io"
	"net"
)

func handleConnection(clientConn net.Conn, upstream string) {
	defer clientConn.Close()

	upstreamConn, err := net.Dial("tcp", upstream)
	if err != nil {
		errLogger.Printf("net.Dial failed, error: %s.", err)
	}
	defer upstreamConn.Close()

	isClientConnClosed := make(chan bool, 1)
	isUpstreamConnClosed := make(chan bool, 1)
	go copy(upstreamConn, clientConn, isClientConnClosed)
	go copy(clientConn, upstreamConn, isUpstreamConnClosed)

	select {
	case <-isClientConnClosed:
		upstreamConn.Close()
	case <-isUpstreamConnClosed:
		clientConn.Close()
	}
}

func copy(dst, src net.Conn, isSrcClosed chan<- bool) {
	io.Copy(dst, src)
	isSrcClosed <- true
	infoLogger.Printf("copy: %s --> %s done.", src.RemoteAddr(), dst.RemoteAddr())
}
