package config

import (
	"fmt"
	"testing"

	"github.com/test-go/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	err := LoadFromFile("../config.demo.yaml")
	require.Nil(t, err, "LoadFromFile err")
	fmt.Println(GConfig.ToString())
}
