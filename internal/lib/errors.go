package lib

type ErrUserExists struct{}

func (e *ErrUserExists) Error() string {
	return "User already exists"
}

type ErrUserNotFound struct{}

func (e *ErrUserNotFound) Error() string {
	return "User not found"
}

type ErrWrongPasswordOrLogin struct{}

func (e *ErrWrongPasswordOrLogin) Error() string {
	return "Wrong password or login"
}
