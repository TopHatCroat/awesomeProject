package helpers

import (
	"github.com/pressly/chi/render"
	"net/http"
)

type ErrResponse struct {
	Err            error `json:"-"`    // low-level runtime error
	HTTPStatusCode int   `json:"code"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "You fucked up",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
var ErrServer = &ErrResponse{HTTPStatusCode: 500, StatusText: "I fucked up"}
var ErrAuth = &ErrResponse{HTTPStatusCode: 401, StatusText: "Unauthorized"}
