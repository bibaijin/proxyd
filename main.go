package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	// LogFlag 控制日志的前缀
	LogFlag = log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	// MaxConnectionNum 表示最大连接数
	MaxConnectionNum = 60000
)

var (
	errLogger  = log.New(os.Stderr, "ERROR ", LogFlag)
	infoLogger = log.New(os.Stdout, "INFO ", LogFlag)

	port            = flag.Int("port", 8080, "port to listen")
	serviceProcType = flag.String("serviceproctype", "worker", "proc type of the service")
	serviceName     = flag.String("servicename", "", "name of the service")
	watchAddr       = flag.String("watchaddr", "lainlet.lain:9001", "the address to watch for upstreams")
	watchHeartbeat  = flag.Int("watchheartbeat", 5, "watch heartbeat interval")

	test             = flag.Bool("test", false, "test mode")
	upstreams        = flag.String("upstreams", "", "test upstreams")               // 只在 test 模式下使用
	cpuprofile       = flag.String("cpuprofile", "", "write cpu profile `file`")    // 只在 test 模式下使用
	memprofile       = flag.String("memprofile", "", "write memory profile `file`") // 只在 test 模式下使用
	blockProfile     = flag.String("blockprofile", "", "block profile")             // 只在 test 模式下使用
	blockProfileRate = flag.Int("blockprofilerate", 0, "block profile rate")        // 只在 test 模式下使用
)

func main() {
	flag.Parse()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	stopWatch := make(chan struct{}, 1)
	stopProduceConns := make(chan struct{}, 1)

	watchDone := make(chan struct{}, 1)
	produceConnsDone := make(chan struct{}, 1)
	consumeConnsDone := make(chan struct{}, 1)
	consumeToCloseConnsDone := make(chan struct{}, 1)

	watcher := Setup(stopWatch, watchDone)

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("net.Listen failed, error: %s.", err)
	}

	infoLogger.Printf("net.Listen()..., port: %d.", *port)

	conns := make(chan net.Conn, MaxConnectionNum)
	toCloseConns := make(chan net.Conn, 2*MaxConnectionNum)

	defer func() {
		// 停止监控 lainlet
		watcher.Close(stopWatch)
		<-watchDone

		// 停止监听端口
		if err = ln.Close(); err != nil {
			errLogger.Printf("ln.Close() failed, error: %s.", err)
		}
		stopProduceConns <- struct{}{}
		<-produceConnsDone
		<-consumeConnsDone
		<-consumeToCloseConnsDone

		Teardown()
		infoLogger.Print("Shutdown gracefully.")
	}()

	go ProduceConns(ln, conns, toCloseConns, stopProduceConns, produceConnsDone)
	go ConsumeConns(watcher, conns, toCloseConns, consumeConnsDone)
	go ConsumeToCloseConns(toCloseConns, consumeToCloseConnsDone)

	signal := <-quit
	infoLogger.Printf("Receive a signal: %d, and accept() will shutdown gracefully...", signal)
}
