package types

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/buger/jsonparser"

	"sol_block_extractord/config"
)

const (
	ColumnNameP    = "p"
	ColumnNameOp   = "op"
	ColumnNameTick = "tick"
	ColumnNameAmt  = "amt"
	ColumnNameLim  = "lim"
	ColumnNameMax  = "max"
)

const (
	OpDeploy   = "deploy"
	OpMint     = "mint"
	OpTransfer = "transfer"
)

const ReasonWrongTick = "wrong tick"

var validOps = []string{OpDeploy, OpMint, OpTransfer} // todo as parameters

const MemoPrefix = "data:,"

var MemoPrefixLen = len(MemoPrefix)

func FilterPrefix(prefix string) bool {
	return prefix == MemoPrefix
}

type Memo struct {
	P    string
	Op   string
	Tick string
	Amt  string
	Lim  string
	Max  string

	AmtN int64
	LimN int64
	MaxN int64
}

func (m *Memo) IsMintOp() bool {
	return m.Op == OpMint
}

func (m *Memo) ShouldParseTxTransferValue() bool {
	return m.IsMintOp() && !config.Cfg.Biz.FreeMint // || m.IsBuyOp()
}

func (m *Memo) IsValidOp() (bool, string) {
	switch m.Op {
	case OpDeploy:
		return m.MaxN > 0 && m.LimN > 0 && m.MaxN >= m.LimN, "wrong max or lim in deploy op"
	case OpMint:
		return m.AmtN > 0, "wrong amt in mint"
	case OpTransfer:
		return m.AmtN > 0, "wrong amt in transfer"
	default:
		return false, "pass"
	}
}

func (m *Memo) AdjustOp() {
	switch m.Op {
	case OpDeploy:
		m.Amt = ""
		m.AmtN = 0
	case OpMint, OpTransfer:
		m.Max = ""
		m.MaxN = 0
		m.Lim = ""
		m.LimN = 0
	default:
	}
}

func (m *Memo) IsValidTick() (pass bool, reason string) {
	pass = m.Tick == config.Cfg.Biz.Ins.Tick
	if !pass {
		reason = ReasonWrongTick
	}
	return
}

func ParseMemo(base58Memo string) (memo Memo, err error) {
	memoBase58Decoded := string(base58.Decode(base58Memo))
	memoBase64Decoded, err := base64.StdEncoding.DecodeString(memoBase58Decoded)
	if err != nil {
		err = errors.New(fmt.Sprintf("decode memo in base64 err: %v", err))
		return
	}

	if len(memoBase64Decoded) <= MemoPrefixLen {
		err = errors.New(fmt.Sprintf("memo too short"))
		return
	}

	if string(memoBase64Decoded[:MemoPrefixLen]) != MemoPrefix {
		err = errors.New(fmt.Sprintf("memo prefix not 'data:,'"))
		return
	}

	memoJson := memoBase64Decoded[MemoPrefixLen:]
	if !json.Valid(memoJson) {
		err = errors.New("invalid json format")
		return
	}

	p, pE := jsonparser.GetString(memoJson, ColumnNameP)
	if pE != nil {
		err = errors.New(fmt.Sprintf("memo parse p err: %v", pE))
		return
	}
	memo.P = p

	op, opE := jsonparser.GetString(memoJson, ColumnNameOp)
	if opE != nil {
		err = errors.New(fmt.Sprintf("memo parse op err: %v", opE))
		return
	}
	memo.Op = op

	tick, tickE := jsonparser.GetString(memoJson, ColumnNameTick)
	if tickE != nil {
		err = errors.New(fmt.Sprintf("memo parse tick err: %v", tickE))
		return
	}
	memo.Tick = strings.ToUpper(tick)

	amt, amtE := jsonparser.GetString(memoJson, ColumnNameAmt)
	if amtE == nil {
		memo.Amt = amt

		amtN, s2iE := strconv.ParseInt(amt, 10, 64)
		if s2iE == nil {
			memo.AmtN = amtN
		}
	}

	lim, limE := jsonparser.GetString(memoJson, ColumnNameLim)
	if limE == nil {
		memo.Lim = lim

		limN, s2iE := strconv.ParseInt(lim, 10, 64)
		if s2iE == nil {
			memo.LimN = limN
		}
	}

	max, maxE := jsonparser.GetString(memoJson, ColumnNameMax)
	if maxE == nil {
		memo.Max = max

		maxN, s2iE := strconv.ParseInt(max, 10, 64)
		if s2iE == nil {
			memo.MaxN = maxN
		}
	}

	memo.AdjustOp()

	return
}
