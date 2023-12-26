package filters

import (
	"fmt"

	"sol_block_extractord/config"
	"sol_block_extractord/types"
)

var (
	validOps        = []string{"deploy", "mint", "transfer"} // todo as parameters
	invalidOpReason = fmt.Sprintf("only support ops: %v", validOps)
)

func MemoFilterP(p string) bool {
	return p == config.Cfg.Biz.Ins.P
}

func MemoFilterOp(Op string) bool {
	for _, op := range validOps {
		if Op == op {
			return true
		}
	}
	return false
}

func FilterMemo(memo types.Memo) (pass bool, reason string) {
	pass = false

	if !MemoFilterP(memo.P) {
		return false, fmt.Sprintf("wrong p: %s", memo.P)
	}

	if !MemoFilterOp(memo.Op) {
		return false, invalidOpReason
	}

	pass, reason = memo.IsValidOp()
	if !pass {
		return
	}

	pass, reason = memo.IsValidTick()
	if !pass {
		return
	}

	return
}
