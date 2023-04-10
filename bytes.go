package util

func BytesJoin(arr ...[]byte) []byte {
	var n int
	for i := 0; i < len(arr); i++ {
		n += len(arr[i])
	}
	if n <= 0 {
		return nil
	}
	builder := GetBufferPool()
	defer PutBufferPool(builder)
	builder.Grow(n)
	for _, s := range arr {
		builder.Write(s)
	}
	return builder.Bytes()
}
