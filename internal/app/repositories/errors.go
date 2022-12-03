package repositories

import "fmt"

var ErrUserAlreadyExists = fmt.Errorf("user with given login already exists")
var ErrUserDoesNotExist = fmt.Errorf("user does not exist")
var ErrOrderAlreadyExists = fmt.Errorf("order with this number already exists")
var ErrCanNotWithdrawBalance = fmt.Errorf("can not withdraw balance")
