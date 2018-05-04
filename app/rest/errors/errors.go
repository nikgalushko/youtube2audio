package errors

import (
	"net/http"

	"github.com/go-chi/render"
)

var (
	NotFound       = &Renderer{Status: http.StatusNotFound}
	InvalidRequest = &Renderer{Status: http.StatusBadRequest}
)

type Renderer struct {
	Status int   `json:"status"`
	Error  error `json:"error,omitempty"`
}

func (er *Renderer) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, er.Status)
	return nil
}
