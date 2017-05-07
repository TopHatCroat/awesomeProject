package main

import (
	"context"
	"errors"
	"fmt"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/TopHatCroat/awesomeProject/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
	"net/http"
)

var (
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

func main() {
	var err error
	db, err = gorm.Open("postgres", "host=127.0.0.1 port=5432 user=postgres dbname=postgres sslmode=disable password=postgres123")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.AutoMigrate(&models.Point{}, &models.User{}, &models.Session{})

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

	router.Route("/users", func(r chi.Router) {
		r.With(DBConn).Post("/", models.CreateUser)
		r.With(DBConn).Get("/", models.ListUsers)
	})

	router.With(DBConn).Post("/login", models.LoginUser)

	router.Get("/googleLogin", models.GoogleLogin)
	router.Get("/oauth2", models.GoogleCallback)
	router.With(DBConn).Post("/glogin", models.GOAuthLogin)

	http.ListenAndServe(":3000", router)
}

func DBConn(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		handler.ServeHTTP(rw, r.WithContext(ctx))
	})
}
