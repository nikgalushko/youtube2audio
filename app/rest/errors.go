package rest

import (
	"net/http"

	"github.com/go-chi/render"
)

var (
	errorNotFound       = &errorRenderer{Status: http.StatusNotFound}
	errorInvalidRequest = &errorRenderer{Status: http.StatusBadRequest}
)

type errorRenderer struct {
	Status int   `json:"status"`
	Error  error `json:"error,omitempty"`
}

func (er *errorRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, er.Status)
	return nil
}
