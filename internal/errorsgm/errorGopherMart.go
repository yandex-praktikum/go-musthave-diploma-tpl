package errorsgm

import (
	"fmt"
)

var (
	ErrLoadedEarlierThisUser    = fmt.Errorf("order was loaded earlier by this user")
	ErrLoadedEarlierAnotherUser = fmt.Errorf("order was loaded earlier by another user")
	ErrDontHavePoints           = fmt.Errorf("not enough points")
	ErrAccrualGetError          = fmt.Errorf("error get in accular")
)
