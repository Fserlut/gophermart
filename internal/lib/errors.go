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

type ErrOrderAlreadyCreated struct{}

func (e *ErrOrderAlreadyCreated) Error() string {
	return "Order already created"
}

type ErrOrderAlreadyCreatedByOtherUser struct{}

func (e *ErrOrderAlreadyCreatedByOtherUser) Error() string {
	return "Order already created by other user"
}

type NotFoundUserIDInContext struct{}

func (e *NotFoundUserIDInContext) Error() string {
	return "Unauthorized"
}

type OrderNotFound struct{}

func (e *OrderNotFound) Error() string {
	return "order not found"
}

type TooManyRequestsError struct{}

func (e *TooManyRequestsError) Error() string {
	return "rate limit exceeded"
}
