package helpers

import (
	"github.com/pressly/chi/render"
	"net/http"
)

type SuccessResponse struct {
	HTTPStatusCode int    `json:"code"` // http response status code
	StatusText     string `json:"status"`
}

func (sr *SuccessResponse) Render(rw http.ResponseWriter, req *http.Request) error {
	render.Status(req, sr.HTTPStatusCode)
	return nil
}

var SucCreate = &SuccessResponse{HTTPStatusCode: 200, StatusText: "Resource created."}
var SucUpdate = &SuccessResponse{HTTPStatusCode: 200, StatusText: "Resource updated."}
var SucDelete = &SuccessResponse{HTTPStatusCode: 200, StatusText: "Resource deleted."}
