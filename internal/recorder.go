package internal

import (
	"bytes"
	"io"
	"net/http"
)

// ResponseRecorder captures response data for caching or idempotency storage.
type ResponseRecorder struct {
	http.ResponseWriter
	Status    int
	Body      *bytes.Buffer
	Committed bool
}

// NewResponseRecorder creates a new ResponseRecorder wrapping the given http.ResponseWriter.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		Body:           new(bytes.Buffer),
		Status:         http.StatusOK,
	}
}

// WriteHeader captures the status code and writes it to the underlying ResponseWriter.
func (r *ResponseRecorder) WriteHeader(code int) {
	if r.Committed {
		return
	}
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
	r.Committed = true
}

// Write captures the response body and writes it to the underlying ResponseWriter.
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}
	// MultiWriter streams to the client and our buffer simultaneously
	return io.MultiWriter(r.ResponseWriter, r.Body).Write(b)
}

// Flush implementation for buffered underlying writers (important for HTTP/2)
func (r *ResponseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
