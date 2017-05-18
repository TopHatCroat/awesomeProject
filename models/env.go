package models

import (
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/jinzhu/gorm"
	"github.com/pressly/chi/render"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"errors"
)

type Env struct {
	DB         *gorm.DB
	PolygonAdd chan Polygon
}

func NewEnviroment(db *gorm.DB) *Env {
	env := &Env{DB: db, PolygonAdd: make(chan Polygon)}

	return env
}

func (e *Env) ImageUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 * 1024 * 1024) //5 MB
	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	m := r.MultipartForm

	files := m.File["img"]

	if len(files) == 0 {
		render.Render(w, r, h.ErrRender(errors.New("No file with name 'img'")))
		return
	}

	for i, _ := range files {
		//for each fileheader, get a handle to the actual file
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}
		//create destination file making sure the path is writeable.
		workDir, _ := os.Getwd()
		filesDir := filepath.Join(workDir, "imgs")

		dst, err := os.Create(filepath.Join(filesDir, files[i].Filename))
		defer dst.Close()
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}
		//copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

	}

	//user, ok := r.Context().Value("user").(*User)
	//fmt.Printf("Current user: %s %s\n", user, ok)
	//if ok != true {
	//	render.Render(w, r, h.ErrServer)
	//	return
	//}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, h.SucCreate)
}


