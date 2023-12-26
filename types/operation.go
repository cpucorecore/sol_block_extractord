package types

import (
	"encoding/json"
	"github.com/holiman/uint256"
)

type Operation struct {
	BlockHeight  int64
	BlockTimeSec int64
	TxIdx        int
	TxHash       string
	From         string
	To           string
	Denom        string
	Value        *uint256.Int
	M            Memo
	MemoRaw      string

	BlockHeightStr  string
	BlockTimeSecStr string
}

func (op *Operation) ToString() string {
	bs, _ := json.Marshal(*op)
	return string(bs)
}
