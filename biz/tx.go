package biz

import (
	"errors"
	"fmt"
	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/holiman/uint256"

	"sol_block_extractord/types"
)

var msgTypes = map[string]interface{}{
	"/cosmos.bank.v1beta1.MsgSend":      banktypes.MsgSend{},
	"/cosmos.bank.v1beta1.MsgMultiSend": banktypes.MsgMultiSend{},
}

const MaxMemoLen = 512

func Tx2Op(txBytes []byte) (op types.Operation, err error) {
	txRaw := cosmostx.TxRaw{}
	err = txRaw.Unmarshal(txBytes)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal TxRaw err: %v", err))
		return
	}

	txBody := cosmostx.TxBody{}
	err = txBody.Unmarshal(txRaw.BodyBytes)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal txBody err: %v", err))
		return
	}

	if len(txBody.Memo) > MaxMemoLen { //TODO check "abc\n\n\n"
		err = errors.New("memo too long")
		return
	}

	memo, err := types.ParseMemo(txBody.Memo)
	if err != nil {
		err = errors.New(fmt.Sprintf("parse memo err: %v", err))
		return
	}

	op.M = memo
	op.MemoRaw = txBody.Memo

	switch msgTypes[txBody.Messages[0].TypeUrl].(type) {
	case banktypes.MsgSend:
		var result banktypes.MsgSend
		result.XXX_Unmarshal(txBody.Messages[0].Value)

		op.From = result.FromAddress
		op.To = result.ToAddress
		op.Denom = result.Amount[0].Denom
		v, _ := uint256.FromBig(result.Amount[0].Amount.BigInt())
		op.Value = v
	default:
		return op, errors.New("not support MsgMultiSend and other msg")
	}

	return
}
