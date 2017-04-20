package main

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	api "github.com/laincloud/lainlet/api/v2"
	"github.com/laincloud/lainlet/client"
)

const retryInterval = 5 * time.Second

var (
	errInvalidMessage = errors.New("invalid message")
	upstreamsLock     sync.RWMutex
)

// watcher 负责从 lainlet 得到 upstreams 信息
type watcher struct {
	address       string
	appName       string
	procType      string
	procName      string
	timeout       time.Duration
	heartbeat     int
	upstreams     []string
	upstreamIndex int
	ctx           context.Context
	cancel        context.CancelFunc
}

func newWatcher(addr, appName, procType, procName string, timeout time.Duration, heartbeat int) *watcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &watcher{
		address:       addr,
		appName:       appName,
		procType:      procType,
		procName:      procName,
		timeout:       timeout,
		heartbeat:     heartbeat,
		upstreams:     nil,
		upstreamIndex: -1,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (w *watcher) Start(quit <-chan bool) {
	c := client.New(w.address)
	proxyData := new(api.ProxyData)
	containerName := buildContainerName(w.appName, w.procType, w.procName)
	uri := buildURI(proxyData, w.appName, w.heartbeat)

	for {
		events, err := c.Watch(uri, w.ctx)
		if err != nil {
			errLogger.Printf("client.Watch failed, error: %s.", err)
		}

		for event := range events {
			infoLogger.Printf("Receive an event, id: %d, event: %s", event.Id, event.Event)
			if event.Id != 0 {
				if err := proxyData.Decode(event.Data); err != nil {
					errLogger.Printf("proxyData.Decode failed, error: %s.", err)
				}

				infoLogger.Printf("data: %+v.", proxyData.Data)
				for k, v := range proxyData.Data {
					if k == containerName {
						w.updateUpstreams(v)
						break
					}
				}
			}
		}

		select {
		case <-time.After(retryInterval):
			infoLogger.Printf("Will watch again...")
		case <-quit:
			return
		}
	}
}

func (w *watcher) Close() {
	w.cancel()
}

func (w *watcher) Upstream() string {
	upstreamsLock.Lock()
	w.upstreamIndex++
	w.upstreamIndex %= len(w.upstreams)
	upstream := w.upstreams[w.upstreamIndex]
	upstreamsLock.Unlock()

	return upstream
}

func (w *watcher) updateUpstreams(procInfo api.ProcInfo) {
	upstreams := make([]string, len(procInfo.Containers))

	for i, v := range procInfo.Containers {
		var buf bytes.Buffer
		buf.WriteString(v.ContainerIp)
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(v.ContainerPort))
		upstreams[i] = buf.String()
	}

	upstreamsLock.Lock()
	w.upstreams = upstreams
	w.upstreamIndex = -1
	upstreamsLock.Unlock()
}

func buildContainerName(appName, procType, procName string) string {
	var buf bytes.Buffer
	buf.WriteString(appName)
	buf.WriteString(".")
	buf.WriteString(procType)
	buf.WriteString(".")
	buf.WriteString(procName)

	return buf.String()
}

func buildURI(proxyData *api.ProxyData, appName string, heartbeat int) string {
	var buf bytes.Buffer
	buf.WriteString(proxyData.URI())
	buf.WriteString("?appname=")
	buf.WriteString(appName)
	buf.WriteString("&heartbeat=")
	buf.WriteString(strconv.Itoa(heartbeat))

	return buf.String()
}
