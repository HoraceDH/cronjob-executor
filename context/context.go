package context

import (
	"sync"
	"sync/atomic"
)

// Created in 2025-03-23 17:08.
// @author Horace

var DispatcherStopped atomic.Bool = atomic.Bool{}
var Shutdown atomic.Bool = atomic.Bool{}
var Version = "Go-1.0.0"
var SignKey atomic.Value = atomic.Value{}
var WaitGroup sync.WaitGroup = sync.WaitGroup{}
