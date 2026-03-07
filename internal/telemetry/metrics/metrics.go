package metrics

//go:generate $GOPATH/bin/mockgen -destination=prometheus_mock.go -package=metrics github.com/prometheus/client_golang/prometheus Registerer

type OpType string

const (
	OpQuery  OpType = "query"
	OpExec   OpType = "exec"
	OpTx     OpType = "tx"
	OpPool   OpType = "pool"
	OpBcrypt OpType = "bcrypt"
)

func (o OpType) String() string {
	return string(o)
}
