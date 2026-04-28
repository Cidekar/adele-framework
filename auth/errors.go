package auth

import "errors"

var InvalidPasswordOrUserError = errors.New("the username or password is incorrect")

var RememberTokenHashError = errors.New("error hashing remember token")

var RememberTokenStoreError = errors.New("error adding remember token to storage")

var RememberTokenDeleteError = errors.New("error removing remember token form storage")
