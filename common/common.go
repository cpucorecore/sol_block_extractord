package common

import "fmt"

func TxCoordinate(height uint64, idx int, txHash string) string {
	return fmt.Sprintf("tx:%d#%d#%s", height, idx, txHash)
}
