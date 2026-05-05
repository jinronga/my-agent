package observability

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const TraceHeader = "X-Trace-ID"

func TraceIDFromRequest(r *http.Request) string {
	if traceID := r.Header.Get(TraceHeader); traceID != "" {
		return traceID
	}
	return NewTraceID()
}

func NewTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "trace_unavailable"
	}
	return hex.EncodeToString(b[:])
}
