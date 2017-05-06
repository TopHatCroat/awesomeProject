package models

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/jinzhu/gorm"
	"github.com/pressly/chi/render"
	"net/http"
	"time"
)

type Session struct {
	gorm.Model
	Token      string
	LastUsedAt int64
	User       User
	UserId     uint
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *LoginRequest) Bind(r *http.Request) error {
	return nil
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (lr *LoginResponse) Render(rw http.ResponseWriter, req *http.Request) error {
	return nil
}

func LoginUser(rw http.ResponseWriter, req *http.Request) {
	data := &LoginRequest{}

	if err := render.Bind(req, data); err != nil {
		render.Render(rw, req, helpers.ErrInvalidRequest(err))
		return
	}

	db, ok := req.Context().Value("db").(*gorm.DB)
	if ok != true {
		render.Render(rw, req, helpers.ErrServer)
		return
	}

	h := sha512.New()
	h.Sum([]byte(data.Password))
	passDigest := h.Sum(nil)

	var user User
	db.Where(&User{Email: data.Email, PassDigest: passDigest}).First(&user)

	if user.Email == "" {
		render.Render(rw, req, helpers.ErrNotFound)
		return
	}

	session := &Session{
		Token:      generateToken(64),
		LastUsedAt: time.Now().UnixNano(),
		UserId:     user.ID,
	}

	db.Create(&session)

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, &LoginResponse{Token: session.Token})
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateToken(s int) string {
	b, err := generateRandomBytes(s)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
