package custom_errors

type AlreadyExistsUserError struct {
}

func (error *AlreadyExistsUserError) Error() string {
	return "Such a user exists already"
}
