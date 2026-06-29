package perf

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/stevelan/cigarmender/internal/log"
)

type Capture[T any] struct {
	value T
}

func (c *Capture[T]) CaptureCPU(operation func() (T, error)) (T, error) {
	f, err := os.Create("cpu.pb.gz")
	if err != nil {
		return c.value, err
	}
	defer log.CloseAndLog("closing CPU capture", f.Close)
	if err := pprof.StartCPUProfile(f); err != nil {
		return c.value, err
	}
	defer pprof.StopCPUProfile()
	i, err := operation()
	return i, err
}

func CaptureHeap() error {
	runtime.GC() // run GC first to capture only the most current objects
	f, err := os.Create("heap.pb.gz")
	if err != nil {
		return err
	}
	defer log.CloseAndLog("closing heap capture", f.Close)
	return pprof.Lookup("heap").WriteTo(f, 0)
}

func CaptureAllocs() error {
	f, err := os.Create("allocs.pb.gz")
	if err != nil {
		return err
	}
	defer log.CloseAndLog("Closing alloc caputre", f.Close)
	return pprof.Lookup("allocs").WriteTo(f, 0)
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Verbose("Task completed", "task", name, "elapsed", elapsed)
}
