package domain

import (
	"strconv"
)

type UserID int64

func (u UserID) String() string {
	return strconv.FormatInt(int64(u), 10)
}
