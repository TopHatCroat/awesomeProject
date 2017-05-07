package models

import (
	h "github.com/TopHatCroat/awesomeProject/helpers"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/chi/render"
	"net/http"
)

type Point struct {
	Id        int     `json:"id"`
	Longitude float32 `json:"log"`
	Latitude  float32 `json:"lat"`
}

type PointRequest struct {
	*Point
}

func (p *PointRequest) Bind(r *http.Request) error {
	// just a post-process after a decode..
	return nil
}

type PointResponse struct {
	*Point
}

func (e *Env) List(rw http.ResponseWriter, req *http.Request) {
	var points = []*Point{}
	e.DB.Find(&points)

	if err := render.RenderList(rw, req, NewPointListResponse(points)); err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}
}

func NewPointListResponse(points []*Point) []render.Renderer {
	list := []render.Renderer{}
	for _, point := range points {
		list = append(list, NewPointResponse(point))
	}
	return list
}

func (e *Env) Create(rw http.ResponseWriter, req *http.Request) {
	data := &PointRequest{}

	if err := render.Bind(req, data); err != nil {
		render.Render(rw, req, h.ErrInvalidRequest(err))
		return
	}

	e.DB.Create(data.Point)

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, NewPointResponse(data.Point))
}

func NewPointResponse(p *Point) *PointResponse {
	resp := &PointResponse{Point: p}
	return resp
}

func (rd *PointResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire

	return nil
}
