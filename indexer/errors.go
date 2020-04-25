package indexer

type LoginError struct {
	error
}

func (e *LoginError) Error() string {
	return e.error.Error()
}
