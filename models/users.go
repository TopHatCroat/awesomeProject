package models

import (
	"crypto/sha512"
	"errors"
	"github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/jinzhu/gorm"
	"github.com/pressly/chi/render"
	"net/http"
	"regexp"
)

type User struct {
	gorm.Model
	Email      string `json:"email"`
	PassDigest []byte `json:"-"`
}

type NewUserRequest struct {
	*User
	Password string `json:"password"`
}

type UserRespose struct {
	*User
}

func (u *NewUserRequest) Bind(r *http.Request) error {
	nameRx, err := regexp.Compile("^\\S+@\\S+\\.\\S+$")
	if err != nil {
		return err
	}

	if nameRx.MatchString(u.Email) != true {
		return errors.New("Email not valid")
	}

	passRx, err := regexp.Compile(".{3,}")
	if err != nil {
		return err
	}

	if passRx.MatchString(u.Password) != true {
		return errors.New("Password not valid")
	}

	return nil
}

func (e *Env) CreateUser(rw http.ResponseWriter, req *http.Request) {
	data := &NewUserRequest{}

	if err := render.Bind(req, data); err != nil {
		render.Render(rw, req, helpers.ErrInvalidRequest(err))
		return
	}

	user := data.User

	h := sha512.New()
	h.Sum([]byte(data.Password))
	user.PassDigest = h.Sum(nil)

	if err := e.DB.Create(&user).Error; err != nil {
		render.Render(rw, req, helpers.ErrRender(err))
		return
	}

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, helpers.CreateSuccess)
}

func (e *Env) ListUsers(rw http.ResponseWriter, req *http.Request) {
	var users = []*User{}
	e.DB.Find(&users)

	if err := render.RenderList(rw, req, NewUserListReponse(users)); err != nil {
		render.Render(rw, req, helpers.ErrServer)
		return
	}
}

func NewUserListReponse(users []*User) []render.Renderer {
	list := []render.Renderer{}
	for _, user := range users {
		list = append(list, &UserRespose{user})
	}
	return list
}

func (rd *UserRespose) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}
