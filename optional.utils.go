package Grizzly

/*
Optional is a utility function that returns a pointer to the given value.
Due to the nature of Go, Nullable mean that the value has to be a pointer.
So this function is used to convert a value of various type to a pointer.
*/
func Optional[T any](s T) *T { return &s }
