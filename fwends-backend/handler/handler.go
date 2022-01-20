package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type Input struct {
	Request  *http.Request
	Response http.ResponseWriter
	Params   httprouter.Params
	Logger   *zap.Logger
}

type Handler interface {
	Handle(Input) (int, error)
}

func ToHTTPRouterHandle(h Handler, l *zap.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.Handle(Input{
			Request:  r,
			Response: w,
			Params:   ps,
			Logger:   l,
		})
	}
}
