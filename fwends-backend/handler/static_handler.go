package handler

func NewStaticHandler(content []byte, contentType string, status int) Handler {
	return &staticHandler{content, contentType, status}
}

type staticHandler struct {
	content     []byte
	contentType string
	status      int
}

func (h *staticHandler) Handle(i Input) (int, error) {
	i.Response.Header().Set("Content-Type", h.contentType)
	i.Response.Write(h.content)
	return h.status, nil
}
