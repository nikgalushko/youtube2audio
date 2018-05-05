package private

import (
	"log"
	"net/http"
	"time"

	"github.com/jetuuuu/youtube2audio/app/rest/interfaces"

	"github.com/go-chi/render"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"

	"github.com/jetuuuu/youtube2audio/app/config"
	"github.com/jetuuuu/youtube2audio/app/rest/errors"
	"github.com/jetuuuu/youtube2audio/app/storage"
)

type Server struct {
	token     *jwtauth.JWTAuth
	cfgReader config.ConfigReader
	s         *storage.Storage
}

func New(c config.ConfigReader, store *storage.Storage) *Server {
	s := Server{
		token:     jwtauth.New("HS256", []byte("private_secret"), nil),
		cfgReader: c,
		s:         store,
	}
	return &s
}

func (s *Server) Run() error {
	log.Printf("[private] Run")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.RealIP)
	router.Use(middleware.Throttle(10), middleware.Timeout(30*time.Second))
	router.Use(s.checkIPMiddleware)

	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/converter", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(s.token))
				r.Use(jwtauth.Authenticator)

				r.Post("/job", s.changeJobStatus)
			})

			r.Group(func(r chi.Router) {
				r.Get("/register", s.register)
			})
		})
	})

	err := http.ListenAndServe(":8001", router)
	log.Fatal(err)
	return err
}

func (s Server) register(w http.ResponseWriter, r *http.Request) {
	_, token, err := s.token.Encode(jwtauth.Claims{"exp": time.Now().Add(24 * time.Hour).Unix()})
	if err != nil {
		render.Render(w, r, &errors.Renderer{Status: http.StatusBadRequest, Error: err})
	}

	c := storage.Converter{
		Adress:       r.RemoteAddr,
		RegisterTime: time.Now(),
		Token:        token,
	}
	s.s.Save("converters", r.RemoteAddr, &c)

	render.JSON(w, r, interfaces.JSON{"token": token})
}

func (s Server) changeJobStatus(w http.ResponseWriter, r *http.Request) {
	var err error
	request := struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
	}{}

	defer r.Body.Close()

	if err = render.DecodeJSON(r.Body, &request); err == nil {
		var job storage.Job
		if err = s.s.Load("jobs", request.JobID, &job); err == nil {
			job.Status = request.Status
			err = s.s.Save("jobs", request.JobID, &job)
		}
	}

	if err == nil {
		render.Status(r, http.StatusOK)
	} else {
		render.Render(w, r, &errors.Renderer{Status: http.StatusBadRequest, Error: err})
	}
}

func (s Server) checkIPMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cfg := s.cfgReader.Read()
		if !cfg.Converters.Contains(r.RemoteAddr) {
			render.Render(w, r, &errors.Renderer{Status: http.StatusForbidden})
			return
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
