package main

import (
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/laincloud/proxyd/log"
)

func setup() {
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
			log.Fatalf("upstreams is required.")
		}
	}
}

func teardown() {
	if *test {
		pprof.StopCPUProfile()
		_stopMemProfile()
		_stopBlockProfile()
	}
}

func _stopMemProfile() {
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Errorf("os.Create() failed, error: %s.", err)
			return
		}

		runtime.GC()

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Errorf("pprof.WriteHeapProfile() failed, error: %s.", err)
		}

		if err := f.Close(); err != nil {
			log.Errorf("f.Close() failed, error: %s.", err)
		}
	}
}

func _stopBlockProfile() {
	if *blockProfile != "" && *blockProfileRate > 0 {
		f, err := os.Create(*blockProfile)
		if err != nil {
			log.Errorf("os.Create() failed, error: %s.", err)
			return
		}

		if err := pprof.Lookup("block").WriteTo(f, 1); err != nil {
			log.Errorf("*Profile.WriteTo() failed, error: %s.", err)
		}

		if err := f.Close(); err != nil {
			log.Errorf("f.Close() failed, error: %s.", err)
		}
	}
}
