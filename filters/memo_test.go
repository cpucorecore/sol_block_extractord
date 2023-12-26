package filters

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"sol_block_extractord/config"
	"sol_block_extractord/types"
)

type TestCaseMemo struct {
	pass  bool
	input string
	desc  string
}

var testcases = [...]TestCaseMemo{
	{true, `data:,{"p":"test-20","op":"deploy","tick":"dcba","max":"100","lim":"50"}`, ""},
	{true, `data:,{"p":"test-20","op":"deploy","tick":"dcba","max":"100","lim":"50"}`, ""},
	{true, `data:,{"p":"test-20","op":"deploy","tick":"dcba","max":"100","lim":"50"}`, ""},
	{false, `data:,{"p":"test-20","op":"deploy","tick":"dcba","max":"100","lim":"500"}`, ""},
	{true, `data:,{"p":"test-20","op":"mint","tick":"dcba","amt":"100"}`, ""},
	{true, `data:,{"p":"test-20","op":"transfer","tick":"dcba","amt":"100"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba","max":"100","lim":"50"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba","max":"100"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba","lim":"100"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba","max":"abc"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba"}`, ""},
	{true, `data:,{"p":"test-20","op":"transfer","tick":"dcba","amt":"100","lim":"50"}`, ""},
	{true, `data:,{"p":"test-20","op":"transfer","tick":"dcba","amt":"100","max":"50"}`, ""},
	{true, `data:,{"p":"test-20","op":"deploy","tick":"dcba","amt":"100","max":"50","lim":"50"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba", "amt":"abc"}`, ""},
	{true, `data:,{"p":"test-20","op":"mint","tick":"dcba", "amt":"111"}`, ""},
	{false, `data:,{"p":"test-20","op":"mint","tick":"dcba", "amt":"abc"}`, ""},
	{true, `data:,{"p":"test-20","op":"mint","tick":"dcba", "amt":"111"}`, ""},
}

func TestFilterMemo(t *testing.T) {
	config.Cfg.Biz.Ins.P = "test-20"
	config.Cfg.Biz.Ins.Tick = tick

	for i, tc := range testcases {
		memo, err := types.ParseMemo(base64.StdEncoding.EncodeToString([]byte(tc.input)))
		require.Equal(t, true, err == nil, fmt.Sprintf("case%d", i))
		pass, reason := FilterMemo(memo)
		require.Equal(t, tc.pass, pass, fmt.Sprintf("case%d, reason:%s", i, reason))
	}
}
