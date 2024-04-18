package goutil

func IF[T any](b bool, v1 T, v2 T) T {
	if b {
		return v1
	}
	return v2
}
