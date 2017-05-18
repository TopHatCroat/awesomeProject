package main

import (
	"errors"
	"flag"
	"fmt"
	c "github.com/TopHatCroat/awesomeProject/control"
	h "github.com/TopHatCroat/awesomeProject/helpers"
	"github.com/TopHatCroat/awesomeProject/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pressly/chi"
	"github.com/pressly/chi/docgen"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
	"net/http"
	"net/http/httputil"
)

var (
	db        *gorm.DB
	genRoutes = flag.Bool("routes", false, "Generate router documentation")
)

func init() {
	render.Respond = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		render.DefaultResponder(w, r, v)
	}
}

func main() {
	flag.Parse()
	var err error
	db, err = gorm.Open("postgres", "host=127.0.0.1 port=5432 user=postgres dbname=postgres sslmode=disable password=postgres123")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.LogMode(true)
	db.AutoMigrate(&models.Point{}, &models.User{}, &models.Session{}, &models.Polygon{})

	e := models.NewEnviroment(db)

	go c.InitDanger(e)
	go c.DangerProcessing(e)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(OptionsAllowed)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		request, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s \n", request)
		w.Write([]byte("I am root"))
	})

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		request, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s \n", request)
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
		router.Get("/", e.ListPoints)
		//router.Options("/", Dummy)
		router.With(e.Authenticate).Post("/", e.CreatePoint)

		router.Route("/:id", func(r chi.Router) {
			r.Use(e.PointCtx)
			r.Get("/", e.GetPoint)
			r.With(e.Authenticate).Put("/", e.UpdatePoint)
			r.With(e.Authenticate).Delete("/", e.DeletePoint)
		})
	})

	router.With(e.Authenticate).Route("/polygons", func(r chi.Router) {
		r.Post("/", e.CreatePolygon)
		r.Get("/", e.GetPolygonList)
		r.Get("/check", e.CheckPointInPoly)
		r.Route("/:id", func(r2 chi.Router) {
			r2.Use(e.PolygonCtx)
			r2.Get("/", e.GetPolygon)
		})
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/", e.CreateUser)
		r.Get("/", e.ListUsers)

		r.With(e.Authenticate).Route("/:id", func(r2 chi.Router) {
			r2.Use(e.UserCtx)
			r2.Post("/fcm", e.RegisterFCM)
		})
	})

	router.Post("/login", e.LoginUser)

	router.Get("/googleLogin", e.GoogleLogin)
	router.Get("/oauth2", e.GoogleCallback)
	router.Post("/glogin", e.GOAuthLogin)

	if *genRoutes {
		// fmt.Println(docgen.JSONRoutesDoc(r))
		fmt.Println(docgen.MarkdownRoutesDoc(router, docgen.MarkdownOpts{
			ProjectPath: "github.com/TopHatCroat/awesomeProject",
			Intro:       "Awesome generated docs.",
		}))
		return
	}

	http.ListenAndServe(":3000", router)
}
func OptionsAllowed(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Allow", "GET,HEAD,POST,OPTIONS,PUT,DELETE")
			w.Header().Set("Content-Type", "httpd/unix-directory")
			return
		}

		handler.ServeHTTP(w, r)
	})
}

//func DBConn(handler http.Handler) http.Handler {
//	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
//		ctx := context.WithValue(r.Context(), "db", db)
//		handler.ServeHTTP(rw, r.WithContext(ctx))
//	})
//}
