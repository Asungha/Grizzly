package exception

type IException interface {
	Error() string
}

type EmptyFunctionError struct{}

func (e *EmptyFunctionError) Error() string {
	return "Empty function error"
}

func isGrizzlyError(err error) bool {
	_, ok := err.(IException)
	return ok
}

func Catch[T IException](err error) error {
	if isGrizzlyError(err) {
		return err
	}
	return nil
}
