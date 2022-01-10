package util

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// A decorated version of httprouter.Handle that takes a custom logger and returns a status and optional error.
type DecoratedHandle func(http.ResponseWriter, *http.Request, httprouter.Params, *log.Entry) (int, error)

// Wraps a decorated handle to a regular handle.
// Custom logging logic and other error checking is injected here too.
func WrapDecoratedHandle(fn DecoratedHandle) httprouter.Handle {

	type responseBody struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Error   error  `json:"error,omitempty"`
	}

	httpDebugEnabled := viper.GetBool("http_debug")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// generate an identifier to uniquely identify the request
		id := rand.Int63()

		// create a logger with useful contextuals
		logger := log.WithFields(log.Fields{
			"request_id": id,
			"method":     r.Method,
			"url":        r.URL,
		}).WithField("request_id", id)
		logger.Debug("Request received")

		// wrap the response writer to later check whether header and status were written
		wrap := responseWriterWrapper{
			Writer: w,
			Logger: logger,
		}

		// call the decorated handler
		status, err := fn(&wrap, r, ps, logger)
		logger.WithField("status", status).Debug("Request processed")

		// log any returned error, with log level corresponding to status
		if err != nil {
			logger := logger.WithError(err)
			switch {
			case status >= 400 && status <= 499:
				logger.Warn("Error returned by http handler")
			case status >= 500 && status <= 599:
				logger.Error("Error returned by http handler")
			default:
				logger.Info("Error returned by http handler")
			}
		}

		if wrap.HeaderWritten { // status and header already written

			// log any status discrepency
			if status != wrap.StatusWritten {
				logger.WithFields(log.Fields{
					"status":        status,
					"statusWritten": wrap.StatusWritten,
				}).Error("Status returned from handler does not match written status")
			}

		} else { // status, header and body are yet to be written

			// return a response such as
			// {
			//   "status": 500,
			//   "message": "Internal Server Error",
			//   "error": "this bit is omittted when http_debug is not true"
			// }
			w.WriteHeader(status)
			resbody := responseBody{
				Status:  status,
				Message: http.StatusText(status),
			}
			if httpDebugEnabled {
				resbody.Error = err
			}
			json.NewEncoder(w).Encode(resbody)

		}

	}
}

type responseWriterWrapper struct {
	Writer        http.ResponseWriter
	HeaderWritten bool
	StatusWritten int
	Logger        *log.Entry
}

func (w *responseWriterWrapper) Header() http.Header {
	return w.Writer.Header()
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if !w.HeaderWritten {
		w.StatusWritten = http.StatusOK
	}
	w.HeaderWritten = true
	return w.Writer.Write(b)
}

func (w *responseWriterWrapper) WriteHeader(status int) {
	if !w.HeaderWritten {
		w.StatusWritten = status
	} else {
		w.Logger.WithFields(log.Fields{
			"status":        status,
			"statusWritten": w.StatusWritten,
		}).Error("Unable to write http status, it has already been written")
	}
	w.HeaderWritten = true
	w.Writer.WriteHeader(status)
}
