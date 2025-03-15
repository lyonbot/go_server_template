package misc

type H map[string]any

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func IsZero[T comparable](value T) bool {
	var zero T
	return zero == value
}

func IsNotZero[T comparable](value T) bool {
	return !IsZero(value)
}
