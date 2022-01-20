package handler

import (
	"math/rand"

	"go.uber.org/zap"
)

type LoggingHandler struct {
	handler Handler
}

func NewLoggingHandler(h Handler) LoggingHandler {
	return LoggingHandler{handler: h}
}

func (h LoggingHandler) Handle(i Input) (int, error) {
	// add useful fields to logger
	i.Logger = i.Logger.With(
		zap.Int64("request_id", rand.Int63()), // random identifier
		zap.String("method", i.Request.Method),
		zap.String("url", i.Request.URL.String()),
	)

	// handle the request
	i.Logger.Debug("request received")
	status, err := h.handler.Handle(i)
	logger := i.Logger.With(zap.Int("status", status))
	logger.Debug("request processed")

	// log any returned error, with log level corresponding to status
	if err != nil {
		logger := i.Logger.With(zap.Error(err))
		switch {
		case status >= 400 && status <= 499:
			logger.Warn("error returned by http handler")
		case status >= 500 && status <= 599:
			logger.Error("error returned by http handler")
		default:
			logger.Info("error returned by http handler")
		}
	}

	// check if connection has closed prematurely
	select {
	case <-i.Request.Context().Done():
		logger.Warn("request cancelled or interrupted")
	default:
	}

	return status, err
}
