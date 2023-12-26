package main

import (
	"context"
	"fmt"
	"time"

	"github.com/InjectiveLabs/sdk-go/client/tm"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"

	"sol_block_extractord/finished_block_manager"
	"sol_block_extractord/log"
)

type BlockInfo struct {
	B *rpctypes.ResultBlock
	R *rpctypes.ResultBlockResults
}

func DispatchTasks(startHeight int64, taskCh chan int64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//cli := tm.NewRPCClient(GInjTmEndpoint)
	cli := tm.NewRPCClient("http://8.217.163.175:26657")

	taskCh <- startHeight
	cursor := startHeight

	var errCnt int
	var start time.Time
	var duration time.Duration
	for {
		errCnt = 0
		start = time.Now()
		latestBlockHeight, err := cli.GetLatestBlockHeight(ctx)
		duration = time.Now().Sub(start)
		if err != nil {
			errCnt++
			log.Logger.Warn(fmt.Sprintf("inj GetLatestBlockHeight failed %d times with err %s, time elapse ms %d", errCnt, err.Error(), duration.Milliseconds()))
			time.Sleep(time.Second * 3)
			continue
		}
		log.Logger.Info(fmt.Sprintf("inj GetLatestBlockHeight success with retry count %d, time elapse ms %d", errCnt, duration.Milliseconds()))

		if latestBlockHeight <= cursor {
			log.Logger.Warn(fmt.Sprintf("inj GetLatestBlockHeight remote height %d < local height %d", latestBlockHeight, cursor))
			time.Sleep(time.Second * 1)
			continue
		}
		log.Logger.Info(fmt.Sprintf("inj GetLatestBlockHeight: cursor %d, remote height %d", cursor, latestBlockHeight))

		for height := cursor + 1; height <= latestBlockHeight; height++ {
			taskCh <- height
		}

		cursor = latestBlockHeight
	}
}

func SyncBlocks(endpoint string, workerId int, taskCh chan int64, blockCh chan BlockInfo) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//injCli := tm.NewRPCClient(endpoint)
	injCli := tm.NewRPCClient("http://8.217.163.175:26657")

	workerBufferCh := make(chan BlockInfo, 50)
	go func() {
		for b := range workerBufferCh {
			taskCoordinate := fmt.Sprintf("(workerId%d, task%d) workerBuffer length:%d", workerId, b.B.Block.Height, len(workerBufferCh))
			var waitCnt int64 = 0
			for { // wait the worker's turn to commit finished work
				fh := finished_block_manager.Get()
				if b.B.Block.Height == fh+1 {
					log.Logger.Info(fmt.Sprintf("task %s committed with wait time ms:%v", taskCoordinate, waitCnt*time.Millisecond.Milliseconds()))
					break
				}
				time.Sleep(time.Millisecond * 5)
				waitCnt++
				if waitCnt%1000 == 0 {
					log.Logger.Info(fmt.Sprintf("task %s wait with time ms:%v", taskCoordinate, waitCnt*time.Millisecond.Milliseconds()))
				}
			}
			blockCh <- b
		}
	}()

	var start, end time.Time
	for task := range taskCh {
		taskCoordinate := fmt.Sprintf("(workerId%d, task%d)", workerId, task)
		log.Logger.Info(fmt.Sprintf("task %s begin", taskCoordinate))

		getBlockFailedCnt := 0
		getBlockResultsFailedCnt := 0
		for {
			start = time.Now()
			b, err := injCli.GetBlock(ctx, task)
			end = time.Now()
			durationMs := end.Sub(start).Milliseconds()
			if err != nil {
				getBlockFailedCnt++
				log.Logger.Warn(fmt.Sprintf("task %s do 'GetBlock' failed %d times with err: [%v], elapse ms:%v", taskCoordinate, getBlockFailedCnt, err, durationMs))
				time.Sleep(time.Second * 5)
				continue
			}
			log.Logger.Info(fmt.Sprintf("task %s do 'GetBlock' succeed with failed count %d, elapse ms:%v", taskCoordinate, getBlockFailedCnt, durationMs))

			start = time.Now()
			r, err := injCli.GetBlockResults(ctx, task)
			end = time.Now()
			if err != nil {
				getBlockResultsFailedCnt++
				log.Logger.Warn(fmt.Sprintf("task %s do 'GetBlockResults' failed %d times with err: [%v], elapse ms:%v", taskCoordinate, getBlockResultsFailedCnt, err, durationMs))
				time.Sleep(time.Second * 5)
				continue
			}
			log.Logger.Info(fmt.Sprintf("task %s do 'GetBlockResults' succeed with failed count %d, elapse ms:%v", taskCoordinate, getBlockResultsFailedCnt, durationMs))

			workerBufferCh <- BlockInfo{b, r}
			log.Logger.Info(fmt.Sprintf("task %s commit to worker buffer", taskCoordinate))
			break
		}
	}
}
