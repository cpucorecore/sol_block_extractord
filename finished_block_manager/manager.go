package finished_block_manager

import (
	"sync/atomic"
)

var finishedHeight int64

func Setup(startBlock int64) {
	finishedHeight = 0
	if startBlock > 0 {
		Update(startBlock - 1)
	}
}

func Get() int64 {
	return atomic.LoadInt64(&finishedHeight)
}

func Update(height int64) {
	old := atomic.LoadInt64(&finishedHeight)
	atomic.CompareAndSwapInt64(&finishedHeight, old, height)
}
