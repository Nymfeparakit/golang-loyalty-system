package repositories

import "fmt"

var ErrUserAlreadyExists = fmt.Errorf("user with given login already exists")
var ErrUserDoesNotExist = fmt.Errorf("user does not exist")
