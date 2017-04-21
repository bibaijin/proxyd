package main

import (
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
)

// Setup 设置程序的初始状态
func Setup(quit <-chan struct{}, done chan<- struct{}) *Watcher {
	if *test {
		if *cpuprofile != "" {
			f, err := os.Create(*cpuprofile)
			if err != nil {
				log.Fatalf("os.Create() failed, error: %s.", err)
			}

			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatalf("pprof.StartCPUProfile() failed, error: %s.", err)
			}
		}

		if *blockProfile != "" && *blockProfileRate > 0 {
			runtime.SetBlockProfileRate(*blockProfileRate)
		}

		if *upstreams == "" {
			log.Fatal("upstreams is required.")
		}

		us := strings.Split(*upstreams, ",")
		watcher := newWatcher(*watchAddr, "", *serviceProcType, *serviceProcName, *watchHeartbeat)
		watcher.upstreams = us

		infoLogger.Print("Watcher.Run() done.")
		done <- struct{}{}

		return watcher
	}

	serviceName := os.Getenv("LAIN_APPNAME")
	if serviceName == "" {
		log.Fatal("No service name found.")
	}

	if *serviceProcName == "" {
		*serviceProcName = serviceName
	}

	watcher := newWatcher(*watchAddr, serviceName, *serviceProcType, *serviceProcName, *watchHeartbeat)
	go watcher.Run(quit, done)
	return watcher
}

// ProduceConns 监听客户端请求
func ProduceConns(ln net.Listener, conns chan<- net.Conn, toCloseConns chan<- net.Conn, quit <-chan struct{}, done chan<- struct{}) {
	running := true
	for running {
		conn, err := ln.Accept()
		if err != nil {
			errLogger.Printf("ln.Accept() failed, error: %s.", err)
		} else {
			conns <- conn
		}

		select {
		case <-quit:
			running = false
		default:
		}
	}

	close(conns)
	infoLogger.Print("ProduceConns() done.")
	done <- struct{}{}
}

// ConsumeConns 处理客户端请求
func ConsumeConns(watcher *Watcher, conns <-chan net.Conn, toCloseConns chan<- net.Conn, done chan<- struct{}) {
	var wg sync.WaitGroup
	for conn := range conns {
		wg.Add(1)

		go func(conn net.Conn) {
			defer wg.Done()
			produceToCloseConns(conn, watcher, toCloseConns)
		}(conn)
	}
	wg.Wait()

	close(toCloseConns)
	infoLogger.Print("ConsumeConns() done.")
	done <- struct{}{}
}

// produceToCloseConns 先处理客户端请求，并视情况连接后端，然后将处理完的连接传入 toCloseConns
func produceToCloseConns(conn net.Conn, watcher *Watcher, toCloseConns chan<- net.Conn) {
	upstream, err := watcher.Upstream()
	if err != nil {
		infoLogger.Printf("watcher.Upstream() failed, error: %s.", err)
		toCloseConns <- conn
		return
	}

	infoLogger.Printf("watcher.Upstream(), upstream: %s.", upstream)

	upstreamConn, err := net.Dial("tcp", upstream)
	if err != nil {
		errLogger.Printf("net.Dial() failed, error: %s.", err)
		toCloseConns <- conn
		return
	}

	pipe(conn, upstreamConn, toCloseConns)
}

// ConsumeToCloseConns 关闭废弃的连接
func ConsumeToCloseConns(toCloseConns <-chan net.Conn, done chan<- struct{}) {
	for conn := range toCloseConns {
		if err := conn.Close(); err != nil {
			errLogger.Printf("conn.Close() failed, error: %s.", err)
		}
	}

	infoLogger.Print("ConsumeToCloseConns() done.")
	done <- struct{}{}
}

// Teardown 做退出前的清理工作
func Teardown() {
	if *test {
		pprof.StopCPUProfile()
		stopMemProfile()
		stopBlockProfile()
	}
}

func stopMemProfile() {
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			errLogger.Printf("os.Create() failed, error: %s.", err)
			return
		}

		runtime.GC()

		if err := pprof.WriteHeapProfile(f); err != nil {
			errLogger.Printf("pprof.WriteHeapProfile() failed, error: %s.", err)
		}

		if err := f.Close(); err != nil {
			errLogger.Printf("f.Close() failed, error: %s.", err)
		}
	}
}

func stopBlockProfile() {
	if *blockProfile != "" && *blockProfileRate > 0 {
		f, err := os.Create(*blockProfile)
		if err != nil {
			errLogger.Printf("os.Create() failed, error: %s.", err)
			return
		}

		if err := pprof.Lookup("block").WriteTo(f, 1); err != nil {
			errLogger.Printf("*Profile.WriteTo() failed, error: %s.", err)
		}

		if err := f.Close(); err != nil {
			errLogger.Printf("f.Close() failed, error: %s.", err)
		}
	}
}
