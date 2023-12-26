package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	injcommon "github.com/InjectiveLabs/sdk-go/client/common"
	"github.com/urfave/cli/v2"

	"sol_block_extractord/biz"
	"sol_block_extractord/common"
	"sol_block_extractord/config"
	"sol_block_extractord/filters"
	"sol_block_extractord/finished_block_manager"
	"sol_block_extractord/log"
	"sol_block_extractord/postgres"
	"sol_block_extractord/types"
)

func main() {
	app := &cli.App{
		Name: "inj_extractor",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "inj_network_name",
				Value:       "mainnet",
				Destination: &config.Cfg.Inj.NetworkName,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "inj_network_node",
				Value:       "lb",
				Destination: &config.Cfg.Inj.NetworkNode,
				Required:    true,
			},
			&cli.Int64Flag{
				Name:        "inj_start_block",
				Value:       0,
				Destination: &config.Cfg.Inj.StartBlock,
				Required:    true,
			},
			&cli.IntFlag{
				Name:        "inj_block_workers",
				Value:       1,
				Destination: &config.Cfg.Inj.BlockWorkers,
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
			&cli.Int64Flag{
				Name:        "open_mint_height",
				Value:       0,
				Destination: &config.Cfg.Biz.OpenMintHeight,
				Required:    true,
			},
			&cli.Int64Flag{
				Name:        "open_transfer_height",
				Value:       0,
				Destination: &config.Cfg.Biz.OpenTransferHeight,
				Required:    true,
			},
			&cli.Int64Flag{
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

					injNetwork := injcommon.LoadNetwork(config.Cfg.Inj.NetworkName, config.Cfg.Inj.NetworkNode)
					GInjTmEndpoint = injNetwork.TmEndpoint

					finished_block_manager.Setup(config.Cfg.Inj.StartBlock)

					taskCh := make(chan int64, 10000)
					go DispatchTasks(config.Cfg.Inj.StartBlock, taskCh)

					blockCh := make(chan BlockInfo, 1000)
					for workerId := 0; workerId < config.Cfg.Inj.BlockWorkers; workerId++ {
						go SyncBlocks(GInjTmEndpoint, workerId, taskCh, blockCh)
					}

					operationCh := make(chan types.Operation, 1000)
					go postgres.PostOperations(operationCh)

					for b := range blockCh {
						curHeight := b.B.Block.Height
						log.Logger.Info(fmt.Sprintf("block:%d with %d txs begin", curHeight, len(b.B.Block.Txs)))

						for txIdx, tx := range b.B.Block.Txs {
							txHash := "0x" + hex.EncodeToString(tx.Hash())
							txCoordinate := common.TxCoordinate(curHeight, txIdx, txHash)
							log.Logger.Info(fmt.Sprintf("%s begin", txCoordinate))

							if b.R.TxsResults[txIdx].Code != 0 {
								log.Logger.Info(fmt.Sprintf("ignore %s for 'tx failed on blockchain'", txCoordinate))
								continue
							}

							op, err := biz.Tx2Op(tx)
							if err != nil {
								log.Logger.Info(fmt.Sprintf("Tx2Op %s failed, err: %v", txCoordinate, err))
								continue
							}

							op.BlockHeight = b.B.Block.Height
							op.BlockHeightStr = strconv.FormatInt(op.BlockHeight, 10)

							op.BlockTimeSec = b.B.Block.Header.Time.Unix()
							op.BlockTimeSecStr = strconv.FormatInt(op.BlockTimeSec, 10)

							op.TxIdx = txIdx
							op.TxHash = txHash

							pass, reason := filters.FilterOperation(op)
							if !pass {
								log.Logger.Info(fmt.Sprintf("%s filtered with reason: [%s]", txCoordinate, reason))
								continue
							}

							operationCh <- op
							log.Logger.Info(fmt.Sprintf("%s commit operation to queue", txCoordinate))
						}

						log.Logger.Info(fmt.Sprintf("block:%d all operations commit to queue", curHeight))
						finished_block_manager.Update(curHeight)
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
