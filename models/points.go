package models

import (
	"context"
	"fmt"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"net/http"
	"strconv"
)

type Point struct {
	gorm.Model
	Longitude float32 `json:"lon"`
	Latitude  float32 `json:"lat"`
	user      User
	UserID    uint `json:"userId"`
}

type PointRequest struct {
	*Point
}

func (PointRequest) Bind(r *http.Request) error {
	return nil
}

type PointResponse struct {
	*Point
}

func (rd *PointResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

func (e *Env) List(rw http.ResponseWriter, req *http.Request) {
	var points = []*Point{}
	e.DB.Find(&points)

	if err := render.RenderList(rw, req, NewPointListResponse(points)); err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}
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
	data.Point.ID = 0

	if err := e.DB.Create(data.Point).Error; err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}

	if user.Fcm != "" {
		user.PushPointNotification(*data.Point)
	}

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, NewPointResponse(data.Point))
}

func (e *Env) GetPoint(w http.ResponseWriter, r *http.Request) {
	point := r.Context().Value("point").(*Point)

	if err := render.Render(w, r, NewPointResponse(point)); err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
}

func (e *Env) UpdatePoint(w http.ResponseWriter, r *http.Request) {
	point := r.Context().Value("point").(*Point)

	p := &PointRequest{Point: point}
	if err := render.Bind(r, p); err != nil {
		render.Render(w, r, h.ErrInvalidRequest(err))
		return
	}

	if err := e.DB.Model(point).Update(map[string]interface{}{"longitude": p.Longitude, "latitude": p.Latitude}).Error; err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	if err := render.Render(w, r, NewPointResponse(point)); err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
}

func (e *Env) DeletePoint(w http.ResponseWriter, r *http.Request) {
	point := r.Context().Value("point").(*Point)

	if err := e.DB.Delete(point).Error; err != nil {
		render.Render(w, r, h.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, h.SucDelete)
}

func (e *Env) PointCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pointId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		point := Point{}
		if err := e.DB.First(&point, pointId).Error; err != nil {
			render.Render(w, r, h.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "point", &point)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewPointListResponse(points []*Point) []render.Renderer {
	list := []render.Renderer{}
	for _, point := range points {
		list = append(list, NewPointResponse(point))
	}
	return list
}

func NewPointResponse(p *Point) *PointResponse {
	resp := &PointResponse{Point: p}
	return resp
}
