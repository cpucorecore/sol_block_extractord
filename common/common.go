package common

import "fmt"

func TxCoordinate(height int64, idx int, txHash string) string {
	return fmt.Sprintf("tx:%d#%d#%s", height, idx, txHash)
}
