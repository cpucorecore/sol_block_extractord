package types

import (
	"encoding/json"
	"github.com/holiman/uint256"
	"strconv"
)

type Operation struct { // TODO rename to Transaction
	BlockHeight  uint64
	BlockTimeSec int64
	TxIdx        int
	TxHash       string
	From         string
	To           string
	Denom        string
	Value        *uint256.Int
	MemoRaw      string
	M            Memo

	BlockHeightStr  string
	BlockTimeSecStr string
}

func (op *Operation) ToString() string {
	bs, _ := json.Marshal(*op)
	return string(bs)
}

func (op *Operation) SetupBlockInfo(blockHeight uint64, blockTimeSec int64, txIdx int) {
	op.BlockHeight = blockHeight
	op.BlockTimeSec = blockTimeSec
	op.TxIdx = txIdx
	op.BlockHeightStr = strconv.FormatUint(op.BlockHeight, 10)
	op.BlockTimeSecStr = strconv.FormatInt(op.BlockTimeSec, 10)
}
