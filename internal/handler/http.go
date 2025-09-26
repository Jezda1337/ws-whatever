package handler

import (
	"html/template"
	"net/http"
)

type HTTPHandler struct {
	tmpl *template.Template
}

func NewHTTPHandler(tmpl *template.Template) *HTTPHandler {
	return &HTTPHandler{tmpl: tmpl}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Execute(w, nil)
}
