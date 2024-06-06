package brc20

type Operation string

const (
	OperationDeploy   Operation = "deploy"
	OperationMint     Operation = "mint"
	OperationTransfer Operation = "transfer"
)

func (o Operation) IsValid() bool {
	switch o {
	case OperationDeploy, OperationMint, OperationTransfer:
		return true
	}
	return false
}

func (o Operation) String() string {
	return string(o)
}
