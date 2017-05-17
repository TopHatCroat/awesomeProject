package models

import (
	"context"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"net/http"
	"strconv"
)

type Polygon struct {
	h.Model
	Typ    string  `json:"type"`
	Points []Point `gorm:"many2many:polygon_points" json:"points"`
}

type PolygonRequest struct {
	*Polygon
}

func (PolygonRequest) Bind(r *http.Request) error {
	return nil
}

type PolygonResponse struct {
	*Polygon
}

func (p *PolygonResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewPolygonReponse(p *Polygon) *PolygonResponse {
	resp := &PolygonResponse{Polygon: p}
	return resp
}

func (e *Env) CreatePolygon(w http.ResponseWriter, r *http.Request) {
	data := &PolygonRequest{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, h.ErrInvalidRequest(err))
		return
	}

	if err := e.DB.Create(data.Polygon).Error; err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	e.PolygonAdd <- *data.Polygon

	render.Status(r, http.StatusCreated)
	render.Render(w, r, h.SucCreate)
}

func (e *Env) GetPolygon(w http.ResponseWriter, r *http.Request) {
	poly := r.Context().Value("poly").(*Polygon)

	if err := render.Render(w, r, NewPolygonReponse(poly)); err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
}

func (e *Env) PolygonCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pointId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		poly := Polygon{}
		if err := e.DB.First(&poly, pointId).Error; err != nil {
			render.Render(w, r, h.ErrNotFound)
			return
		}

		points := []Point{}
		if err := e.DB.Model(&poly).Related(&points, "Points").Error; err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		poly.Points = points

		ctx := context.WithValue(r.Context(), "poly", &poly)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
