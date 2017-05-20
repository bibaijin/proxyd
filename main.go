package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/laincloud/proxyd/log"
)

var (
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

func init() {
	flag.Parse()
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	done := make(chan struct{})
	defer close(done)

	var watcher *Watcher
	if *test {
		setup()
		defer teardown()
		us := strings.Split(*upstreams, ",")
		watcher = newWatcher(*watchAddr, "", *serviceProcType, *serviceName, *watchHeartbeat)
		watcher.upstreams = us
	} else {
		serviceAppName := os.Getenv("LAIN_APPNAME")
		if serviceAppName == "" {
			log.Fatalf("No ${LAIN_APPNAME} environment variable found.")
		}

		if *serviceName == "" {
			log.Fatalf("No service name found.")
		}
		watcher = newWatcher(*watchAddr, serviceAppName, *serviceProcType, *serviceName, *watchHeartbeat)
		go watcher.Run(done)
		defer watcher.Close()
	}

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("net.Listen() failed, error: %s.", err)
	}
	defer func() {
		if err = ln.Close(); err != nil {
			log.Errorf("ln.Close() failed, error: %s.", err)
		}
	}()
	log.Infof("net.Listen()..., port: %d.", *port)

	conns := accept(done, ln)
	go handleConns(conns, watcher)

	<-quit
	log.Infof("Shutting down...")
}
