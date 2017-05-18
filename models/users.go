package models

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/TopHatCroat/awesomeProject/fcm"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/mattermost/gcm"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
	"net/http"
	"regexp"
	"strconv"
)

type User struct {
	h.Model
	Email      string `json:"email"`
	PassDigest []byte `json:"-"`
	Fcm        string `json:"fcm"`
}

type NewUserRequest struct {
	*User
	Password string `json:"password"`
}

type UserRespose struct {
	*User
}

func (u *NewUserRequest) Bind(r *http.Request) error {
	if u.User == nil {
		return errors.New("Invalid request")
	}

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
		render.Render(rw, req, h.ErrInvalidRequest(err))
		return
	}

	user := data.User

	hash := sha512.New()
	hash.Sum([]byte(data.Password))
	user.PassDigest = hash.Sum(nil)

	if err := e.DB.Create(&user).Error; err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, h.SucCreate)
}

func (e *Env) ListUsers(rw http.ResponseWriter, req *http.Request) {
	var users = []*User{}
	e.DB.Find(&users)

	if err := render.RenderList(rw, req, NewUserListReponse(users)); err != nil {
		render.Render(rw, req, h.ErrServer)
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

func (e *Env) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pointId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}

		user := User{}
		if err := e.DB.First(&user, pointId).Error; err != nil {
			render.Render(w, r, h.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "targetUser", &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type FCMRequest struct {
	Token string `json:"fcm_token"`
}

func (f *FCMRequest) Bind(r *http.Request) error {

	return nil
}

func (e *Env) RegisterFCM(rw http.ResponseWriter, req *http.Request) {
	data := &FCMRequest{}

	if err := render.Bind(req, data); err != nil {
		render.Render(rw, req, h.ErrInvalidRequest(err))
		return
	}

	user, ok := req.Context().Value("user").(*User)
	if ok != true {
		render.Render(rw, req, h.ErrServer)
		return
	}

	target, ok := req.Context().Value("targetUser").(*User)
	if ok != true {
		render.Render(rw, req, h.ErrServer)
		return
	}

	if user.ID != target.ID {
		render.Render(rw, req, h.ErrAuth)
		return
	}

	user.Fcm = data.Token
	if err := e.DB.Save(user).Error; err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}

	render.Status(req, http.StatusOK)
	render.Render(rw, req, h.SucCreate)
}

func (u *User) PushPointNotification(point Point) {
	data := map[string]interface{}{"msg": "Point created", "latitude": point.Geo.Lat(), "longitude": point.Geo.Lng()}
	regIDs := []string{u.Fcm}
	gmsg := gcm.NewMessage(data, regIDs...)

	// Create a Sender to send the message.
	sender := &gcm.Sender{ApiKey: fcm.APIKEY}

	// Send the message and receive the response after at most two retries.
	response, err := sender.Send(gmsg, 2)
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return
	}

	fmt.Printf("Success: %d, Failure: %d", response.Results, response.Failure)
}
