package fisco

type Config struct {
	NodeAddr       string
	GroupID        int
	ChainID        int64
	ContractAddr   string
	PrivateKey     string
	TimeoutSeconds int64
	Enabled        bool
}
