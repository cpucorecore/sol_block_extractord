package main

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/urfave/cli/v2"

	"sol_block_extractord/config"
	"sol_block_extractord/filters"
	"sol_block_extractord/finished_block_manager"
	"sol_block_extractord/log"
	"sol_block_extractord/postgres"
	"sol_block_extractord/types"
)

func main() {
	app := &cli.App{
		Name: "sol_extractor",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:        "start_slot",
				Value:       0,
				Destination: &config.Cfg.StartSlot,
				Required:    true,
			},
			&cli.IntFlag{
				Name:        "block_workers",
				Value:       1,
				Destination: &config.Cfg.BlockWorkers,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "pg_host",
				Value:       "127.0.0.1",
				Destination: &config.Cfg.Pg.Host,
				Required:    true,
			},
			&cli.IntFlag{
				Name:        "pg_port",
				Value:       5432,
				Destination: &config.Cfg.Pg.Port,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "pg_user",
				Value:       "postgres",
				Destination: &config.Cfg.Pg.User,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "pg_password", // TODO get password by os env
				Value:       "",
				Destination: &config.Cfg.Pg.Password,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "pg_dbname",
				Value:       "",
				Destination: &config.Cfg.Pg.DbName,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "p",
				Value:       "test-20",
				Destination: &config.Cfg.Biz.Ins.P,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "tick",
				Value:       "dcba",
				Destination: &config.Cfg.Biz.Ins.Tick,
				Required:    true,
			},
			&cli.IntFlag{
				Name:        "memo_len_min_limit",
				Value:       15,
				Destination: &config.Cfg.Biz.MemoLenMin,
			},
			&cli.Uint64Flag{
				Name:        "open_mint_height",
				Value:       0,
				Destination: &config.Cfg.Biz.OpenMintHeight,
				Required:    true,
			},
			&cli.Uint64Flag{
				Name:        "open_transfer_height",
				Value:       0,
				Destination: &config.Cfg.Biz.OpenTransferHeight,
				Required:    true,
			},
			&cli.Uint64Flag{
				Name:        "deploy_height",
				Value:       0,
				Destination: &config.Cfg.Biz.DeployHeight,
				Required:    true,
			},
		},

		Commands: []*cli.Command{
			{
				Name: "start",
				Action: func(context *cli.Context) error {
					log.Logger.Info(fmt.Sprintf("cfg:%s", config.Cfg.ToString()))

					finished_block_manager.Setup(config.Cfg.StartSlot)

					taskCh := make(chan uint64, 10000)
					go SOLDispatchTasks(config.Cfg.StartSlot, taskCh)

					blockCh := make(chan *rpc.GetBlockResult, 1000)
					for workerId := 0; workerId < config.Cfg.BlockWorkers; workerId++ {
						go SOLSyncBlocks(workerId, taskCh, blockCh)
					}

					operationCh := make(chan types.Operation, 1000)
					go postgres.PostOperations(operationCh)

					for b := range blockCh {
						curSlot := b.ParentSlot + 1
						log.Logger.Info(fmt.Sprintf("slot:%d with %d txs begin", curSlot, len(b.Transactions)))

						for txIdx, txWithMeta := range b.Transactions {
							op, err := ParseTx(*b.BlockHeight, txIdx, &txWithMeta, types.ParseMemo)
							if err != nil {
								log.Logger.Info(fmt.Sprintf("ParseTx err: %s", err.Error()))
								continue
							}

							op.SetupBlockInfo(*b.BlockHeight, int64(*b.BlockTime), txIdx)
							memoBase58Decoded := string(base58.Decode(op.MemoRaw))
							op.M, err = types.ParseMemo(memoBase58Decoded)
							if err != nil {
								log.Logger.Info(fmt.Sprintf("ParseMemo memo[%s] err: %s", memoBase58Decoded, err.Error()))
								continue
							}

							pass, reason := filters.FilterOperation(op)
							if !pass {
								log.Logger.Info(fmt.Sprintf("filtered with reason: [%s]", reason))
								continue
							}

							operationCh <- op
							log.Logger.Info(fmt.Sprintf("commit operation to queue"))
						}

						log.Logger.Info(fmt.Sprintf("block:%d all operations commit to queue", curSlot))
						finished_block_manager.Update(curSlot)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Logger.Fatal(err.Error())
	}
}
