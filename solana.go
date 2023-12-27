package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/holiman/uint256"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"sol_block_extractord/common"
	"sol_block_extractord/finished_block_manager"
	"sol_block_extractord/log"
	"sol_block_extractord/types"
)

const (
	NativeDenom               = "SOL"
	SystemInstructionTransfer = 2
)

var (
	memoProgramId           = solana.MustPublicKeyFromBase58("MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr")
	systemTransferProgramId = solana.MustPublicKeyFromBase58("11111111111111111111111111111111")
)

func SOLDispatchTasks(startHeight uint64, taskCh chan uint64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endpoint := rpc.LocalNet_RPC
	cli := rpc.New(endpoint)

	taskCh <- startHeight
	cursor := startHeight

	var errCnt int
	var start time.Time
	var duration time.Duration
	for {
		errCnt = 0
		start = time.Now()
		latestBlockHeight, err := cli.GetBlockHeight(ctx, rpc.CommitmentFinalized)
		duration = time.Now().Sub(start)
		if err != nil {
			errCnt++
			log.Logger.Warn(fmt.Sprintf("sol GetBlockHeight failed %d times with err %s, time elapse ms %d", errCnt, err.Error(), duration.Milliseconds()))
			time.Sleep(time.Second * 3)
			continue
		}
		log.Logger.Info(fmt.Sprintf("sol GetBlockHeight success with retry count %d, time elapse ms %d", errCnt, duration.Milliseconds()))

		if latestBlockHeight <= cursor {
			log.Logger.Warn(fmt.Sprintf("sol GetBlockHeight remote height %d <= local height %d", latestBlockHeight, cursor))
			time.Sleep(time.Second * 1)
			continue
		}
		log.Logger.Info(fmt.Sprintf("sol GetBlockHeight: cursor %d, remote height %d", cursor, latestBlockHeight))

		for height := cursor + 1; height <= latestBlockHeight; height++ {
			taskCh <- height
		}

		cursor = latestBlockHeight
	}
}

func SOLSyncBlocks(workerId int, taskCh chan uint64, blockCh chan *rpc.GetBlockResult) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endpoint := rpc.LocalNet_RPC
	cli := rpc.New(endpoint)

	workerBufferCh := make(chan *rpc.GetBlockResult, 50)
	go func() {
		for b := range workerBufferCh {
			taskCoordinate := fmt.Sprintf("(workerId%d, task%d) workerBuffer length:%d", workerId, *b.BlockHeight, len(workerBufferCh))
			var waitCnt int64 = 0
			for { // wait the worker's turn to commit finished work
				fh := finished_block_manager.Get()
				if b.ParentSlot == fh {
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

		includeRewards := false

		getBlockFailedCnt := 0
		getBlockNilCnt := 0
		for {
			start = time.Now()
			b, err := cli.GetBlockWithOpts(ctx, task, &rpc.GetBlockOpts{
				Encoding:           solana.EncodingBase64,
				Commitment:         rpc.CommitmentConfirmed,
				TransactionDetails: rpc.TransactionDetailsFull,
				Rewards:            &includeRewards,
			})
			end = time.Now()
			durationMs := end.Sub(start).Milliseconds()
			if err != nil {
				getBlockFailedCnt++
				log.Logger.Warn(fmt.Sprintf("task %s do 'GetBlock' failed %d times with err: [%v], elapse ms:%v", taskCoordinate, getBlockFailedCnt, err, durationMs))
				time.Sleep(time.Second * 5)
				continue
			}
			if b == nil {
				getBlockNilCnt++
				log.Logger.Warn(fmt.Sprintf("task %s do 'GetBlock' returns nil %d times ,elapse ms:%v", taskCoordinate, getBlockNilCnt, durationMs))
				time.Sleep(time.Second * 5)
				continue
			}
			log.Logger.Info(fmt.Sprintf("task %s do 'GetBlock' succeed with failed count %d with returns nil count %d, elapse ms:%v", taskCoordinate, getBlockFailedCnt, getBlockNilCnt, durationMs))

			workerBufferCh <- b
			log.Logger.Info(fmt.Sprintf("task %s commit to worker buffer", taskCoordinate))
			break
		}
	}
}

func ParseTx(blockHeight uint64, txIdx int, txWithMeta *rpc.TransactionWithMeta) (op types.Operation, err error) {
	tx, err := txWithMeta.GetTransaction()
	if err != nil {
		// TODO FIXME
		// TODO release resources
		os.Exit(-1)
	}

	if len(tx.Signatures) != 1 {
		// TODO FIXME: until now we don't known how to process a transaction with multiple signatures
		err = errors.New(fmt.Sprintf("tx signatures len %d != 1, ignore", len(tx.Signatures)))
		return
	}

	txCoordinate := common.TxCoordinate(blockHeight, txIdx, tx.Signatures[0].String())

	log.Logger.Info(fmt.Sprintf("--%s begin", txCoordinate))
	if txWithMeta.Meta.Err != nil {
		err = errors.New(fmt.Sprintf("tx not success with err: %v, ignore", txWithMeta.Meta.Err))
		return
	}

	if len(tx.Message.Instructions) != 2 {
		// TODO FIXME: until now we don't want to process a transaction with len(instructions) != 2
		err = errors.New(fmt.Sprintf("tx instructions len %d != 2", len(tx.Message.Instructions)))
		return
	}

	// TODO FIXME: no limit for the sequence of instructions
	transferInst := tx.Message.Instructions[0]
	callMemoInst := tx.Message.Instructions[1]

	if !tx.Message.AccountKeys[transferInst.ProgramIDIndex].Equals(systemTransferProgramId) {
		err = errors.New(fmt.Sprintf("instuction transfer: wrong program id:%s, must:%s",
			tx.Message.AccountKeys[transferInst.ProgramIDIndex].String(),
			systemTransferProgramId.String(),
		))
		return
	}

	if !tx.Message.AccountKeys[callMemoInst.ProgramIDIndex].Equals(memoProgramId) {
		err = errors.New(fmt.Sprintf("instruction memo: wrong program id:%s, must:%s",
			tx.Message.AccountKeys[callMemoInst.ProgramIDIndex].String(),
			memoProgramId.String(),
		))
		return
	}

	if !tx.Message.AccountKeys[transferInst.Accounts[0]].Equals(tx.Message.AccountKeys[callMemoInst.Accounts[0]]) {
		err = errors.New(fmt.Sprintf("transfer instruction from addr:%s != memo instruction from addr:%s",
			tx.Message.AccountKeys[transferInst.Accounts[0]].String(),
			tx.Message.AccountKeys[callMemoInst.Accounts[0]].String(),
		))
		return
	}

	var inst, value uint64
	if len(transferInst.Data) == 12 {
		inst = uint64(binary.LittleEndian.Uint32(transferInst.Data[:4]))
		value = binary.LittleEndian.Uint64(transferInst.Data[4:])
	} else if len(transferInst.Data) == 9 {
		inst = uint64(transferInst.Data[0])
		value = binary.LittleEndian.Uint64(transferInst.Data[1:])
	} else {
		err = errors.New(fmt.Sprintf("wrong system transfer call data len:%d", len(transferInst.Data)))
		return
	}

	if inst != SystemInstructionTransfer {
		err = errors.New(fmt.Sprintf("wrong system transfer instruction:%d", inst))
		return
	}

	op.TxHash = tx.Signatures[0].String()
	op.From = tx.Message.AccountKeys[transferInst.Accounts[0]].String()
	op.To = tx.Message.AccountKeys[transferInst.Accounts[1]].String()
	op.Denom = NativeDenom
	op.Value = uint256.NewInt(value)
	op.MemoRaw = callMemoInst.Data.String()

	return op, nil
}
