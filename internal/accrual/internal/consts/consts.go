package consts

import "fmt"

var (
	ErrNotRegistered = fmt.Errorf("not registered")
	ErrToManyRequest = fmt.Errorf("too many requests")
	ErrInternal      = fmt.Errorf("internal")
)
