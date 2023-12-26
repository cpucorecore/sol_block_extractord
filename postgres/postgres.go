package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"sol_block_extractord/common"
	"sol_block_extractord/config"
	"sol_block_extractord/filters"
	"sol_block_extractord/log"
	"sol_block_extractord/types"
)

type Api interface {
	PostOperation(op types.Operation) error
}

type Cli struct {
	db   *sql.DB
	stmt *sql.Stmt
}

func NewCli() (cli Cli, err error) {
	dataSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Cfg.Pg.Host,
		config.Cfg.Pg.Port,
		config.Cfg.Pg.User,
		config.Cfg.Pg.Password,
		config.Cfg.Pg.DbName)

	log.Logger.Debug("dataSource", zap.String("ds", dataSource))
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.Logger.Error("connect postgres server failed", zap.String("err", err.Error()))
		return
	}

	stmt, err := db.Prepare(
		"INSERT INTO \"Operation\"(\"from\", \"to\", txhash, \"rawData\", \"blockHeight\", p, op, tick, amt, lim, max, \"createdAt\",\"updatedAt\", value, timestamp, \"txIndex\") VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),now(),$12,$13,$14) RETURNING id")
	if err != nil {
		log.Logger.Error("postgres prepare failed", zap.String("err", err.Error()))
		return
	}

	return Cli{db: db, stmt: stmt}, nil
}

const maxRetry = 3

func (cli *Cli) PostOperation(op types.Operation) (err error) {
	retry := 0
	var start, end time.Time
	for {
		start = time.Now()
		_, err = cli.stmt.Exec(op.From, op.To, op.TxHash, op.MemoRaw, op.BlockHeightStr, op.M.P, op.M.Op, op.M.Tick, op.M.Amt, op.M.Lim, op.M.Max, op.Value.String(), op.BlockTimeSecStr, op.TxIdx)
		end = time.Now()
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { //unique constrain
				log.Logger.Info("duplicated txId", zap.String("txId", op.TxHash))
				return nil
			} else {
				retry = retry + 1
				if retry > maxRetry {
					log.Logger.Warn("reach max retry", zap.String("operation", fmt.Sprintf("%v", op)), zap.String("err", err.Error()))
					break
				}
				time.Sleep(time.Second * 2)
			}
		} else {
			log.Logger.Info(fmt.Sprintf("exec sql elapse %v", end.Sub(start).Milliseconds()))
			break
		}
	}

	return err
}

func PostOperations(operationCh chan types.Operation) {
	deployed := config.Cfg.Biz.DeployHeight != 0
	cli, err := NewCli()
	if err != nil {
		log.Logger.Fatal(fmt.Sprintf("pg NewCli err:%v", err))
	}

	for operation := range operationCh {
		txCoordinate := common.TxCoordinate(operation.BlockHeight, operation.TxIdx, operation.TxHash)
		log.Logger.Info(fmt.Sprintf("operation begin: %s", operation.ToString()))

		pass, reason := filters.FilterOperation(operation)
		if !pass {
			log.Logger.Error(fmt.Sprintf("%s filtered with reason: [%s]", txCoordinate, reason))
			continue
		}

		err = cli.PostOperation(operation)
		if err != nil {
			cli.Shutdown()
			log.Logger.Fatal(fmt.Sprintf("!! %s do [operation ==> pg] failed with err:%s!!. begin shutdown", txCoordinate, err.Error()))
		}

		if !deployed && operation.M.Op == types.OpDeploy {
			config.Cfg.Biz.DeployHeight = operation.BlockHeight
			deployed = true
		}

		log.Logger.Info(fmt.Sprintf("*****%s succeed****", txCoordinate))
	}
}

func (cli *Cli) Shutdown() {
	cli.stmt.Close()
	cli.db.Close()
}
