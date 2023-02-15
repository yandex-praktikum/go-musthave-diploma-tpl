package errorsgm

import (
	"fmt"
)

type ErrorGopherMart struct {
	Code string
	Err  error
}

func (err ErrorGopherMart) Error() string {
	return err.Err.Error()
}

const (
	//http
	GetError = "H01"

	//Server status
	StatusOk                  = "S01"
	StatusTooManyRequests     = "S02"
	StatusInternalServerError = "S03"

	//Marshal
	MarshalError   = "M01"
	UnmarshalError = "M02"

	//io
	ReadAllError = "I01"

	//strconv
	ParseIntError = "C01"

	//WTF??
	RespStatusCodeNotMatch = "WTF"
)

var (
	ErrLoadedEarlierThisUser    = fmt.Errorf("order was loaded earlier by this user")
	ErrLoadedEarlierAnotherUser = fmt.Errorf("order was loaded earlier by another user")
	ErrDontHavePoints           = fmt.Errorf("not enough points")
	ErrAccrualGetError          = fmt.Errorf("error get in accular")
)
