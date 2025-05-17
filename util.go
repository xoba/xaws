package xaws

// ptr returns a pointer to the provided value.
func ptr[T any](t T) *T {
	return &t
}
