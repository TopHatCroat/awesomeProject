package main

import (
	"errors"
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

func main() {
	var err error
	db, err = gorm.Open("postgres", "host=127.0.0.1 port=5432 user=postgres dbname=postgres sslmode=disable password=postgres123")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.AutoMigrate(&models.Point{}, &models.User{}, &models.Session{})

	e := models.Env{DB: db}

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
		router.Get("/", e.List)
		router.With(e.Authenticate).Post("/", e.Create)

		router.Route("/:id", func(r chi.Router) {
			r.Use(e.PointCtx)
			r.Get("/", e.GetPoint)
			r.With(e.Authenticate).Put("/", e.UpdatePoint)
			r.With(e.Authenticate).Delete("/", e.DeletePoint)
		})
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/", e.CreateUser)
		r.Get("/", e.ListUsers)
	})

	router.Post("/login", e.LoginUser)

	router.Get("/googleLogin", e.GoogleLogin)
	router.Get("/oauth2", e.GoogleCallback)
	router.Post("/glogin", e.GOAuthLogin)

	http.ListenAndServe(":3000", router)
}

//func DBConn(handler http.Handler) http.Handler {
//	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
//		ctx := context.WithValue(r.Context(), "db", db)
//		handler.ServeHTTP(rw, r.WithContext(ctx))
//	})
//}
