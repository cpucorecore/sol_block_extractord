package config

type Inscription struct {
	P    string
	Tick string
}

type Business struct {
	DeployHeight       int64
	OpenMintHeight     int64
	OpenTransferHeight int64

	MemoLenMin int
	MemoLenMax int

	Ins Inscription
}
