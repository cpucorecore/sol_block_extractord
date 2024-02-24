package config

type Business struct {
	DeployHeight       uint64
	OpenMintHeight     uint64
	OpenTransferHeight uint64

	FreeMint    bool
	ToAddrLimit string

	MemoLenMin int
	MemoLenMax int

	Inscription Inscription
}
