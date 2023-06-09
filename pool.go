package goutil

import (
	"bytes"
	"net/url"
	"strings"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func GetBufferPool() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func PutBufferPool(pool *bytes.Buffer) {
	pool.Reset()
	bufferPool.Put(pool)
}

var stringsBuilderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

func GetStringsBuilder() *strings.Builder {
	builder := stringsBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	return builder
}

func PutStringsBuilder(builder *strings.Builder) {
	builder.Reset()
	stringsBuilderPool.Put(builder)
}

var urlValuesPool = sync.Pool{
	New: func() any {
		return make(url.Values)
	},
}

func GetUrlValues() url.Values {
	vals := urlValuesPool.Get().(url.Values)
	for k := range vals {
		vals.Del(k)
	}
	return vals
}

func PutUrlValues(vals url.Values) {
	urlValuesPool.Put(vals)
}
