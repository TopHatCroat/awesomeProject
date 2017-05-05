package main

import (
	"fmt"
	"context"
	"errors"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/TopHatCroat/awesomeProject/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
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

	db *gorm.DB
)

const htmlIndex = `<html><body>
Login in with <a href="/login">Google</a>
</body></html>
`

func handleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Printf("oauth state, '%s'\n", "wtf")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	fmt.Printf("oauth state, '%s'\n", state)

	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	fmt.Printf("oauth code, '%s'\n", code)

	token, err := oauthConf.Exchange(context.TODO(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	//
	//resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
	//	url.QueryEscape(token.AccessToken))
	fmt.Printf("token, '%s'\n", token)

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	if err != nil {
		fmt.Printf("Get: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer response.Body.Close()

	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("ReadAll: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Printf("parseResponseBody: %s\n", string(result))

	//http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func main() {
	db, _ = gorm.Open("postgres", "host=127.0.0.1 port=5432 user=postgres dbname=postgres sslmode=disable password=postgres123")
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&models.Point{})

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I am root"))
	})

	router.Get("/error", func(rw http.ResponseWriter, req *http.Request) {
		render.Render(rw, req, &h.ErrResponse{
			Err:            errors.New("Example error"),
			HTTPStatusCode: 200,
			StatusText:     "Success",
			ErrorText:      "Example error",
		})
	})

	router.Route("/points", func(router chi.Router) {
		router.With(DBConn).Get("/", models.List)
		router.With(DBConn).Post("/", models.Create)
	})

	router.Get("/login", handleGoogleLogin)
	router.Get("/oauth2", handleGoogleCallback)

	http.ListenAndServe(":3000", router)
}

func DBConn(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		handler.ServeHTTP(rw, r.WithContext(ctx))
	})
}
