package config

type Inscription struct {
	P    string
	Tick string
}

type Business struct {
	DeployHeight       uint64
	OpenMintHeight     uint64
	OpenTransferHeight uint64

	MemoLenMin int
	MemoLenMax int

	Ins Inscription
}
