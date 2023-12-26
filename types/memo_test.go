package types

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"sol_block_extractord/config"
)

type TestCase struct {
	pass  bool
	input string
	desc  string
}

var parseMemoCases = [...]TestCase{
	{true, `data:,{"p":"test-20","op":"deploy","tick":"ttta","max":"100","lim":"50"}`, ""},
	{false, `ddata:,{"p":"test-20","op":"deploy","tick":"ttta","max":"100","lim":"50"}`, ""},
	{false, `datad:,{"p":"test-20","op":"deploy","tick":"ttta","max":"100","lim":"50"}`, ""},
	{false, `:,{"p":"test-20","op":"deploy","tick":"ttta","max":"100","lim":"50"}`, ""},
	{false, `{"p":"test-20","op":"deploy","tick":"ttta","max":"100","lim":"50"}`, ""},
	{false, `data:,{"p":"test-20","op":"deploy","tick":"ttta","max":"100","lim":"50"}d`, ""},
	{false, `data:,{"p":"test-20","op":"deploy","tick":"ttta",max":"100","lim":"50"}d`, ""},
	{false, `data:,{"p":"test-20",,"op":"deploy","tick":"ttta","max":"100","lim":"50"}`, ""},
	{false, `data:,{"p":"test-20","op":"deploy","tickd":"ttta","max":"100","lim":"50"}`, ""},
	{false, `data:,{"p":"test-20"}`, ""},
	{false, `data:,{"lim":"50"}`, ""},
}

func TestParseMemo(t *testing.T) {
	config.Cfg.Biz.Ins.P = "test-20"

	for i, tc := range parseMemoCases {
		_, err := ParseMemo(base64.StdEncoding.EncodeToString([]byte(tc.input)))
		require.Equal(t, tc.pass, err == nil, fmt.Sprintf("case%d failed, err:%v", i, err))
	}
}
