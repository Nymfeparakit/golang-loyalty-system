package services

import "fmt"

var ErrInvalidCredentials = fmt.Errorf("invalid login or password")
var ErrInvalidAccessToken = fmt.Errorf("access token is invalid")
var ErrUserDoesNotExist = fmt.Errorf("user does not exist")
var ErrOrderExistsForOtherUser = fmt.Errorf("order already exists for other user")
var ErrUserAlreadyExists = fmt.Errorf("user with given login already exists")
var ErrTooManyRequests = fmt.Errorf("too many requests")
