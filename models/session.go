package models

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/pressly/chi/render"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "1046962736770-0ss7chk20buubrhpmp6i3hlpj6c3fi6g.apps.googleusercontent.com",
		ClientSecret: "oez3UDoAEWAfBkS6r5CD4Rmm",
		RedirectURL:  "http://localhost:3000/oauth2",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
	oauthStateString = "thisshouldberandom"
)

type Session struct {
	h.Model
	Token      string
	LastUsedAt int64
	User       User
	UserID     uint
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TokenAuth
}

type TokenAuth struct {
	Token string `json:"token"`
}

func (u *LoginRequest) Bind(r *http.Request) error {
	return nil
}

type LoginResponse struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}

func (lr *LoginResponse) Render(rw http.ResponseWriter, req *http.Request) error {
	return nil
}

func (e *Env) LoginUser(rw http.ResponseWriter, req *http.Request) {
	data := &LoginRequest{}

	if err := render.Bind(req, data); err != nil {
		render.Render(rw, req, h.ErrInvalidRequest(err))
		return
	}

	hash := sha512.New()
	hash.Sum([]byte(data.Password))
	passDigest := hash.Sum(nil)

	var user User
	e.DB.Where(&User{Email: data.Email, PassDigest: passDigest}).First(&user)

	if user.Email == "" {
		render.Render(rw, req, h.ErrNotFound)
		return
	}

	session := &Session{
		Token:      generateToken(64),
		LastUsedAt: time.Now().UnixNano(),
		UserID:     user.ID,
	}

	if err := e.DB.Create(&session).Error; err != nil {
		render.Render(rw, req, h.ErrRender(err))
		return
	}

	render.Status(req, http.StatusCreated)
	render.Render(rw, req, &LoginResponse{Token: session.Token})
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
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

func (e *Env) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			render.Render(rw, r, h.ErrAuth)
			return
		}

		ses := Session{}

		if ok := e.DB.Where(&Session{Token: auth}).Preload("User").First(&ses).RecordNotFound(); ok == true {
			render.Render(rw, r, h.ErrAuth)
			return
		}

		ctx := context.WithValue(r.Context(), "user", &ses.User)

		handler.ServeHTTP(rw, r.WithContext(ctx))
	})
}

//func DBConn(handler http.Handler) http.Handler {
//	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
//		ctx := context.WithValue(r.Context(), "db", db)
//		handler.ServeHTTP(rw, r.WithContext(ctx))
//	})
//}

func (e *Env) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (e *Env) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	fmt.Printf("oauth state, '%s'\n", state)

	if state != oauthStateString {
		render.Render(w, r, h.ErrRender(errors.New("Invalid oauth state")))
		return
	}

	code := r.FormValue("code")
	fmt.Printf("oauth code, '%s'\n", code)

	token, err := oauthConf.Exchange(context.TODO(), code)
	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	render.Render(w, r, &LoginResponse{Token: token.AccessToken})
}

func (e *Env) GOAuthLogin(w http.ResponseWriter, r *http.Request) {
	data := &LoginRequest{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, h.ErrInvalidRequest(err))
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + data.Token)

	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
	defer response.Body.Close()

	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}
	log.Printf("parseResponseBody: %s\n", string(result))

	gdata := make(map[string]interface{})
	err = json.Unmarshal(result, &gdata)
	if err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	user := User{}
	e.DB.Where(&User{Email: string(gdata["email"].(string))}).First(&user)

	if user.Email == "" {
		user.Email = string(gdata["email"].(string))
		if err := e.DB.Create(&user).Error; err != nil {
			render.Render(w, r, h.ErrRender(err))
			return
		}
	}

	session := &Session{
		Token:      data.Token,
		LastUsedAt: time.Now().UnixNano(),
		UserID:     user.ID,
	}

	if err := e.DB.Create(&session).Error; err != nil {
		render.Render(w, r, h.ErrRender(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, &LoginResponse{Token: session.Token})
}
