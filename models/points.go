package models

import (
	"fmt"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/chi/render"
	"net/http"
)

type Point struct {
	gorm.Model
	Longitude float32 `json:"lon"`
	Latitude  float32 `json:"lat"`
	user      User
	UserID    uint
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
	Id uint `json:"id"`
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

	user, ok := req.Context().Value("user").(*User)
	fmt.Printf("%s %s", user, ok)
	if ok != true {
		render.Render(rw, req, h.ErrServer)
		return
	}

	data.Point.user = *user
	data.Point.UserID = user.ID

	if err := e.DB.Create(data.Point).Error; err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}

	//fcm.PushNotification("Point created: " + string(data.Point), "")

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, NewPointResponse(data.Point))
}

func NewPointResponse(p *Point) *PointResponse {
	resp := &PointResponse{Id: p.ID}
	return resp
}

func (rd *PointResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire

	return nil
}
