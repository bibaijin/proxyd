package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

// LogFlag 控制日志的前缀
const LogFlag = log.LstdFlags | log.Lmicroseconds | log.Lshortfile

var (
	errLogger   = log.New(os.Stderr, "ERROR ", LogFlag)
	infoLogger  = log.New(os.Stdout, "INFO ", LogFlag)
	debugLogger = log.New(os.Stdout, "DEBUG ", LogFlag)

	port             = flag.Int("port", 8080, "port to listen")
	watcherAddr      = flag.String("watcheraddr", "lainlet.lain:9001/v2", "watcher address to get upstreams")
	watcherHeartbeat = flag.Int("watcherheartbeat", 5, "heartbeat interval for watcher")
	watcherTimeout   = flag.Int("watchertimeout", 30, "timeout for watcher")
	procType         = flag.String("proctype", "worker", "proc type of the service")
	procName         = flag.String("procname", "", "proc name of the service")
)

func main() {
	flag.Parse()

	appName := os.Getenv("LAIN_APPNAME")
	timeout := time.Duration(*watcherTimeout) * time.Second
	watcher := newWatcher(*watcherAddr, appName, *procType, *procName, timeout, *watcherHeartbeat)
	quit := make(chan bool, 1)
	go watcher.Start(quit)
	defer func() {
		watcher.Close()
		quit <- true
	}()

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("net.Listen failed, error: %s.", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			errLogger.Printf("ln.Accept failed, error: %s.", err)
			continue
		}

		go handleConnection(conn, watcher.Upstream())
	}
}
