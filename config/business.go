package config

type Inscription struct {
	P    string
	Tick string
}

type Business struct {
	DeployHeight       uint64
	OpenMintHeight     uint64
	OpenTransferHeight uint64

	FreeMint    bool
	ToAddrLimit string

	MemoLenMin int
	MemoLenMax int

	Ins Inscription
}
