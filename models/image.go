package models

import (
	"errors"
	"fmt"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/pressly/chi/render"
	"github.com/twinj/uuid"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Image struct {
	h.Model
	Path   string `json:"path"`
	User   User   `json:"-"`
	UserID uint   `json:"user_id"`
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

	user, ok := r.Context().Value("user").(*User)
	fmt.Printf("Current user: %s %s\n", user, ok)
	if ok != true {
		render.Render(w, r, h.ErrServer)
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

		//dst, err := os.Create(filepath.Join(filesDir, files[i].Filename))
		fileName := uuid.NewV4().String() + filepath.Ext(files[i].Filename)
		dst, err := os.Create(filepath.Join(filesDir, fileName))
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

		img := Image{UserID: user.ID, Path: "http://" + SERVER_IP + ":3000/imgs/" + fileName}
		if err := e.DB.Create(&img).Error; err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		render.Status(r, http.StatusCreated)
		render.Render(w, r, h.SucCreate)
		return
	}

	render.Status(r, http.StatusBadRequest)
	render.Render(w, r, h.ErrInvalidRequest(errors.New("Something went wrong")))
}

func (e *Env) ListImages(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*User)
	if ok != true {
		render.Render(w, r, h.ErrServer)
		return
	}

	var images = []*Image{}
	if err := e.DB.Where("user_id = ?", user.ID).Find(&images).Error; err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	if err := render.RenderList(w, r, NewImageListReponse(images)); err != nil {
		render.Render(w, r, h.ErrServer)
		return
	}
}

func NewImageListReponse(images []*Image) []render.Renderer {
	list := []render.Renderer{}
	for _, image := range images {
		list = append(list, NewImageResponse(image))
	}
	return list
}

type ImageResponse struct {
	*Image
}

func (ImageResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewImageResponse(p *Image) *ImageResponse {
	resp := &ImageResponse{Image: p}
	return resp
}
