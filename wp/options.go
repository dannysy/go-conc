package wp

import (
	"fmt"
	"runtime"
	"time"
)

var DefaultOptions = Options{
	Size:       uint32(runtime.GOMAXPROCS(0)),
	Idle:       1,
	TaskChSize: uint32(runtime.GOMAXPROCS(0)) * 10, // TODO think about it
	RecoveryFn: func() {
		if msg := recover(); msg != nil {
			fmt.Println(msg)
		}
	},
	OutFn:         func(_ interface{}) {},
	WatcherPeriod: time.Second,
}

// Options contains worker pool behavior properties
type Options struct {
	Size          uint32
	TaskChSize    uint32
	Idle          uint32
	RecoveryFn    func()
	OutFn         func(interface{})
	WatcherPeriod time.Duration
}
