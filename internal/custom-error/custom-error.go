package customerror

import "errors"

var ErrUniqueKeyConstrantViolation = errors.New("unique key violation")
var ErrNoSuchUser = errors.New("no such user")
