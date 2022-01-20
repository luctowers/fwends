package handler

func NewErrorHandler(status int, err error) Handler {
	return &errorHandler{status, err}
}

type errorHandler struct {
	status int
	err    error
}

func (h *errorHandler) Handle(i Input) (int, error) {
	return h.status, h.err
}
