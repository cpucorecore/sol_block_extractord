package filters

import (
	"fmt"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"sol_block_extractord/config"
	"sol_block_extractord/types"
)

type TestCaseOp struct {
	pass bool
	op   types.Operation
	desc string
}

const (
	deployHeight       = 100
	openMintHeight     = 200
	openTransferHeight = 300
	inscriptionP       = "test-20"
	tick               = "DCBA"
	to                 = "to"
	denom              = "denom"
)

func TestFilterOperation(t *testing.T) {
	config.Cfg.Biz.OpenMintHeight = openMintHeight
	config.Cfg.Biz.Ins.P = inscriptionP

	validMemo := types.Memo{
		P:    inscriptionP,
		Op:   "mint",
		Tick: tick,
		Amt:  "100",
		AmtN: 100,
	}
	var tcs = [...]TestCaseOp{
		{true, types.Operation{BlockHeight: openMintHeight, To: to, Denom: denom, Value: uint256.NewInt(10), M: validMemo}, "pass"},
		{true, types.Operation{BlockHeight: openMintHeight + 10, To: to, Denom: denom, Value: uint256.NewInt(10), M: validMemo}, "pass"},
		{false, types.Operation{BlockHeight: openMintHeight - 1, To: to, Denom: denom, Value: uint256.NewInt(10), M: validMemo}, "not open"},
	}

	for i, tc := range tcs {
		pass, reason := FilterOperation(tc.op)
		if pass != tc.pass {
			t.Fatalf("case %d failed, reason: %s", i, reason)
		} else {
			t.Logf("case%d:desc:%s, reason:%s", i, tc.desc, reason)
		}
	}
}

func TestDeploy(t *testing.T) {
	config.Cfg.Biz = config.Business{
		Ins:                config.Inscription{P: inscriptionP, Tick: tick},
		MemoLenMin:         10,
		DeployHeight:       0,
		OpenMintHeight:     100,
		OpenTransferHeight: 200,
	}

	validMemo := types.Memo{
		P:    inscriptionP,
		Op:   "deploy",
		Tick: tick,
		Max:  "1000",
		MaxN: 1000,
		Lim:  "100",
		LimN: 100,
	}

	deployOp := types.Operation{BlockHeight: openMintHeight, To: to, Denom: denom, Value: uint256.NewInt(10), M: validMemo}

	pass, reason := FilterOperation(deployOp)
	t.Log(reason)
	require.Equal(t, true, pass)

	config.Cfg.Biz.DeployHeight = deployHeight
	pass, reason = FilterOperation(deployOp)
	t.Log(reason)
	require.Equal(t, false, pass)
}

func TestMint(t *testing.T) {
	config.Cfg.Biz = config.Business{
		Ins:                config.Inscription{P: inscriptionP, Tick: tick},
		MemoLenMin:         10,
		DeployHeight:       0,
		OpenMintHeight:     openMintHeight,
		OpenTransferHeight: openTransferHeight,
	}

	mintMemo := types.Memo{
		P:    inscriptionP,
		Op:   "mint",
		Tick: tick,
		Amt:  "100",
		AmtN: 100,
	}

	tcs := [...]TestCaseOp{
		{false, types.Operation{BlockHeight: openMintHeight - 1, To: to, Denom: denom, Value: uint256.NewInt(10), M: mintMemo}, "not open"},
		{true, types.Operation{BlockHeight: openMintHeight, To: to, Denom: denom, Value: uint256.NewInt(10), M: mintMemo}, "ok"},
	}

	for i, tc := range tcs {
		pass, reason := FilterOperation(tc.op)
		if pass == false {
			t.Logf("reason:%s", reason)
		}
		require.Equal(t, tc.pass, pass, fmt.Sprintf("case %d failed desc: %s, reason: %s", i, tc.desc, reason))
	}
}

func TestTransfer(t *testing.T) {
	config.Cfg.Biz = config.Business{
		Ins:                config.Inscription{P: inscriptionP, Tick: tick},
		MemoLenMin:         10,
		DeployHeight:       0,
		OpenMintHeight:     openMintHeight,
		OpenTransferHeight: openTransferHeight,
	}

	transferMemo := types.Memo{
		P:    inscriptionP,
		Op:   "transfer",
		Tick: tick,
		Amt:  "100",
		AmtN: 100,
	}

	tcs := [...]TestCaseOp{
		{false, types.Operation{BlockHeight: openTransferHeight - 1, To: to, Denom: denom, Value: uint256.NewInt(10), M: transferMemo}, "not open"},
		{true, types.Operation{BlockHeight: openTransferHeight, To: to, Denom: denom, Value: uint256.NewInt(10), M: transferMemo}, "ok"},
	}

	for i, tc := range tcs {
		pass, reason := FilterOperation(tc.op)
		if pass == false {
			t.Logf("reason:%s", reason)
		}
		require.Equal(t, tc.pass, pass, fmt.Sprintf("case %d failed desc: %s, reason: %s", i, tc.desc, reason))
	}
}
