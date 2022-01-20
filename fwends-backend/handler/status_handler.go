package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type StatusHandler struct {
	handler Handler
	debug   bool
}

func NewStatusHandler(h Handler, debug bool) StatusHandler {
	return StatusHandler{handler: h, debug: debug}
}

func (h StatusHandler) Handle(i Input) (int, error) {
	// wrap the response writer to capture the status
	wrap := responseWriterCapture{Response: i.Response, Logger: i.Logger}
	i.Response = &wrap
	status, err := h.handler.Handle(i)

	// if a status was written, check for discrepency and log it
	if wrap.HeaderWritten {
		if status != wrap.StatusWritten {
			i.Logger.With(
				zap.Int("status", status),
				zap.Int("statusWritten", wrap.StatusWritten),
			).Error("status returned from handler does not match written status")
		}
	}

	// if connection is closed, return
	select {
	case <-i.Request.Context().Done():
		return wrap.StatusWritten, err
	default:
	}

	// return a default response based on status if none was sent
	if !wrap.HeaderWritten {
		if h.debug {
			defaultHandlerResponse(i.Response, status, err)
		} else {
			// potentially sensitive errors shouldn't be sent to client in production
			defaultHandlerResponse(i.Response, status, nil)
		}
	}

	return wrap.StatusWritten, err
}

// returns a response such as
//
// {
//   "status": 500,
//   "message": "Internal Server Error",
//   "error": "this bit is omittted when http_debug is not true"
// }
func defaultHandlerResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resbody := defaultResponseBody{
		Status:  status,
		Message: http.StatusText(status),
		Error:   err,
	}
	json.NewEncoder(w).Encode(resbody)
}

type responseWriterCapture struct {
	Response      http.ResponseWriter
	HeaderWritten bool
	StatusWritten int
	Logger        *zap.Logger
}

func (w responseWriterCapture) Header() http.Header {
	return w.Response.Header()
}

func (w *responseWriterCapture) Write(b []byte) (int, error) {
	if !w.HeaderWritten {
		w.StatusWritten = http.StatusOK
	}
	w.HeaderWritten = true
	return w.Response.Write(b)
}

func (w *responseWriterCapture) WriteHeader(status int) {
	if !w.HeaderWritten {
		w.StatusWritten = status
	} else {
		w.Logger.With(
			zap.Int("status", status),
			zap.Int("statusWritten", w.StatusWritten),
		).Error("Unable to write http status, it has already been written")
	}
	w.HeaderWritten = true
	w.Response.WriteHeader(status)
}

type defaultResponseBody struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   error  `json:"error,omitempty"`
}
