package platform

func recoverJSValue[T any](operation func() T) (value T) {
	defer func() {
		if recover() != nil {
			var zero T
			value = zero
		}
	}()
	return operation()
}
