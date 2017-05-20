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
	"github.com/laincloud/proxyd/log"
)

const retryInterval = 5 * time.Second

var (
	errNoUpstream = errors.New("no upstream")
	upstreamsLock sync.RWMutex
)

// Watcher 负责从 lainlet 得到 upstreams 信息
type Watcher struct {
	address       string
	appName       string
	procType      string
	procName      string
	heartbeat     int
	upstreams     []string
	upstreamIndex int
	ctx           context.Context
	cancel        context.CancelFunc
	proxyData     *api.ProxyData
	containerName string
}

func newWatcher(addr, appName, procType, procName string, heartbeat int) *Watcher {
	upstreams := make([]string, 0)
	ctx, cancel := context.WithCancel(context.Background())
	proxyData := new(api.ProxyData)
	containerName := buildContainerName(appName, procType, procName)

	return &Watcher{
		address:       addr,
		appName:       appName,
		procType:      procType,
		procName:      procName,
		heartbeat:     heartbeat,
		upstreams:     upstreams,
		upstreamIndex: -1,
		ctx:           ctx,
		cancel:        cancel,
		proxyData:     proxyData,
		containerName: containerName,
	}
}

// Run 监听 lainlet 以更新 upstreams
func (w *Watcher) Run(done <-chan struct{}) {
	log.Infof("Watcher.Run()..., watcherAddress: %s, serviceAppName: %s, serviceProcType: %s, serviceProcName: %s.",
		w.address, w.appName, w.procType, w.procName)

	c := client.New(w.address)
	uri := buildURI(w.proxyData, w.appName, w.heartbeat)

	for {
		events, err := c.Watch(uri, w.ctx)
		if err != nil {
			log.Errorf("client.Watch() failed, error: %s.", err)
		}

		log.Infof("Connected to lainlet.")

		for event := range events {
			w.handleEvent(event)
		}

		select {
		case <-done:
			log.Infof("Watcher.Run() done.")
			return
		case <-time.After(retryInterval):
			log.Infof("Will watch lainlet again...")
		}
	}
}

// Close 关闭与 lainlet 的连接
func (w *Watcher) Close() {
	w.cancel()
}

func (w *Watcher) handleEvent(event *client.Response) {
	log.Infof("Receive an event, id: %d, event: %s", event.Id, event.Event)
	if event.Id != 0 {
		if err := w.proxyData.Decode(event.Data); err != nil {
			log.Errorf("proxyData.Decode failed, error: %s.", err)
		}

		log.Infof("data: %+v.", w.proxyData.Data)
		for k, v := range w.proxyData.Data {
			if k == w.containerName {
				w.updateUpstreams(v)
				break
			}
		}
	}
}

// Upstream 用 round-robin 算法返回一个 upstream
func (w *Watcher) Upstream() (string, error) {
	upstreamsLock.Lock()
	defer upstreamsLock.Unlock()

	if len(w.upstreams) == 0 {
		return "", errNoUpstream
	}

	w.upstreamIndex++
	w.upstreamIndex %= len(w.upstreams)
	upstream := w.upstreams[w.upstreamIndex]
	return upstream, nil
}

func (w *Watcher) updateUpstreams(procInfo api.ProcInfo) {
	upstreams := make([]string, len(procInfo.Containers))

	for i, v := range procInfo.Containers {
		var buf bytes.Buffer
		buf.WriteString(v.ContainerIp)
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(v.ContainerPort))
		upstreams[i] = buf.String()
		log.Infof("updateUpstreams..., index: %d, upstream: %s.", i, buf.String())
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
	buf.WriteString("/v2")
	buf.WriteString(proxyData.URI())
	buf.WriteString("?appname=")
	buf.WriteString(appName)
	buf.WriteString("&heartbeat=")
	buf.WriteString(strconv.Itoa(heartbeat))

	return buf.String()
}
