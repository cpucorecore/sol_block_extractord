package finished_block_manager

import (
	"sync/atomic"
)

var finishedHeight uint64

func Setup(startBlock uint64) {
	finishedHeight = 0
	if startBlock > 0 {
		Update(startBlock - 1)
	}
}

func Get() uint64 {
	return atomic.LoadUint64(&finishedHeight)
}

func Update(height uint64) {
	old := atomic.LoadUint64(&finishedHeight)
	atomic.CompareAndSwapUint64(&finishedHeight, old, height)
}
