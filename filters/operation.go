package filters

import (
	"sol_block_extractord/config"
	"sol_block_extractord/types"
)

const (
	ReasonMintNotOpen     = "mint not open"
	ReasonTransferNotOpen = "transfer not open"
	ReasonAlreadyDeployed = "already deployed"
)

func FilterOperation(op types.Operation) (bool, string) {
	pass, reason := FilterMemo(op.M)
	if !pass {
		return false, reason
	}

	if op.M.Op == types.OpDeploy {
		if config.Cfg.Biz.DeployHeight != 0 {
			return false, ReasonAlreadyDeployed
		}
	} else if op.M.Op == types.OpMint {
		if op.BlockHeight < config.Cfg.Biz.OpenMintHeight {
			return false, ReasonMintNotOpen
		}
	} else if op.M.Op == types.OpTransfer {
		if op.BlockHeight < config.Cfg.Biz.OpenTransferHeight {
			return false, ReasonTransferNotOpen
		}
	} else {
		return false, "op not supported"
	}

	return true, ""
}
