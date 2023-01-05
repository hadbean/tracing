package logger

import (
	"bytes"
	"runtime"
)

//用来生成trace_id
var trace_id_map = make(map[string]string)

func Goid() string {
	b := originId()
	if v, e := trace_id_map[string(b)]; e {
		return v
	}
	return string(b)
}
func originId() string {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return string(b)
}
func SetGoid(traceId string) {
	id := originId()
	trace_id_map[id] = traceId
}
func Remove() {
	delete(trace_id_map, originId())
}
