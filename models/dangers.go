package models

import (
	"context"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/paulmach/go.geo"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"net/http"
	"strconv"
	"strings"
)

type Area struct {
	ID     uint    `json:"id"`
	Typ    string  `json:"type"`
	Points []Point `json:"points"`
}

type AreaRequest struct {
	*Area
}

func (a *AreaRequest) Bind(r *http.Request) error {
	return nil
}

type AreaResponse struct {
	*Area
}

func (p *AreaResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewAreaResponse(p *Area) *AreaResponse {
	resp := &AreaResponse{Area: p}
	return resp
}

type DangerResponse struct {
	Areas []uint `json:"areas"`
}

func (DangerResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewAreaListResponse(areas []*Area) []render.Renderer {
	list := []render.Renderer{}
	for _, area := range areas {
		list = append(list, NewAreaResponse(area))
	}
	return list
}

func (e *Env) CreateArea(w http.ResponseWriter, r *http.Request) {
	data := &PolygonRequest{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, h.ErrInvalidRequest(err))
		return
	}

	pointsTest := ""
	//
	//for _, point := range data.Points {
	//	  pointsTest += strconv.FormatFloat(point.Latitude, 'f', -1, 64) + " " + strconv.FormatFloat(point.Longitude, 'f', -1, 64) + ", "
	//}

	pointsTest = strings.TrimRight(pointsTest, ", ")
	strArray := strings.Split(pointsTest, ", ")
	strArray = append(strArray, strArray[0])
	pointsTest = strings.Join(strArray, ", ")
	polyInsert := "POLYGON((" + pointsTest + "))"

	sql := "INSERT INTO dangers VALUES(default, ?, ST_PolygonFromText(?, 0))"
	if err := e.DB.Exec(sql, data.ID, polyInsert).Error; err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}


	render.Status(r, http.StatusCreated)
	render.Render(w, r, h.SucCreate)
}

func (e *Env) GetListArea(w http.ResponseWriter, r *http.Request) {
	var areas = []*Area{}

	rows, err := e.DB.Raw("SELECT id, typ, ST_AsBinary(geom) FROM dangers").Rows()
	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	for rows.Next() {
		var p *geo.PointSet
		var area Area
		rows.Scan(&area.ID, &area.Typ, &p)

		for i := 0; i < p.Length(); i++ {
			//area.Points = append(area.Points, Point{Latitude: p.GetAt(i).Lat(), Longitude: p.GetAt(i).Lng()})
		}

		areas = append(areas, &area)
	}

	if err := render.RenderList(w, r, NewAreaListResponse(areas)); err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
}

func (e *Env) GetArea(w http.ResponseWriter, r *http.Request) {
	poly := r.Context().Value("area").(*Area)

	if err := render.Render(w, r, NewAreaResponse(poly)); err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
}
func (e *Env) CheckPointArea(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	geoPoint := "point(" + lat + " " + lon + ")"

	sql := "select id, st_contains( geom, st_geomfromtext(?, 4326)) from dangers;"

	rows, err := e.DB.Raw(sql, geoPoint).Rows()
	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	var response DangerResponse
	for rows.Next() {
		var id uint
		var state bool
		rows.Scan(&id, &state)

		if state {
			response.Areas = append(response.Areas, id)
		}
	}

	if err := render.Render(w, r, response); err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
}

func (e *Env) AreaCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pointId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		rows, err := e.DB.Raw("SELECT id, typ, ST_AsBinary(geom) FROM dangers WHERE id = ?", pointId).Rows()
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		var area Area
		for rows.Next() {
			var p *geo.PointSet
			rows.Scan(&area.ID, &area.Typ, &p)

			for i := 0; i < p.Length(); i++ {
				//area.Points = append(area.Points, Point{Latitude: p.GetAt(i).Lat(), Longitude: p.GetAt(i).Lng()})
			}

		}
		ctx := context.WithValue(r.Context(), "area", &area)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
