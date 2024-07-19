package tool

func Reference[T any](v T) *T {
	return &v
}
