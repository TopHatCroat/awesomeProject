package points

import (
	"net/http"
	"github.com/pressly/chi/render"
	h "github.com/TopHatCroat/awesomeProject/helpers"
)

type Point struct {
	Id int `json:"id"`
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

var points = []*Point{
	{Id:1, Latitude: 45, Longitude: -15},
	{Id:2, Latitude: 46, Longitude: -14},
	{Id:3, Latitude: 47, Longitude: -13},
	{Id:4, Latitude: 48, Longitude: -12},
	{Id:5, Latitude: 49, Longitude: -11},
}

func List(rw http.ResponseWriter, req *http.Request) {
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

func Create(rw http.ResponseWriter, req *http.Request) {
	data := &PointRequest{}

	if err := render.Bind(req, data); err != nil {
		render.Render(rw, req, h.ErrInvalidRequest(err))
		return
	}

	article := data.Point

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, NewPointResponse(article))
}

func NewPointResponse(p *Point) *PointResponse {
	resp := &PointResponse{Point: p}
	return resp
}

func (rd *PointResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire

	return nil
}
