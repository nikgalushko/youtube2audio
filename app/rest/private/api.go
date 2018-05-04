package private

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"

	"github.com/jetuuuu/youtube2audio/app/config"
	"github.com/jetuuuu/youtube2audio/app/storage"
)

type JSON map[string]interface{}

type Server struct {
	token     *jwtauth.JWTAuth
	cfgReader config.ConfigReader
	cfg       *atomic.Value
	s         *storage.Storage
}

func New(c config.ConfigReader, store *storage.Storage) *Server {
	v := atomic.Value{}
	s := Server{
		token:     jwtauth.New("HS256", []byte("private_secret"), nil),
		cfgReader: c,
		cfg:       &v,
		s:         store,
	}
	return &s
}

func (s *Server) Run() error {
	log.Printf("[private] Run")
	cfg, err := s.cfgReader.Read()
	if err != nil {
		log.Fatalf("running server error %s", err.Error())
		return err
	}
	s.cfg.Store(cfg)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.RealIP)
	router.Use(middleware.Throttle(10), middleware.Timeout(30*time.Second))
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/converter", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(s.token))
				r.Use(jwtauth.Authenticator)
			})

			r.Group(func(r chi.Router) {
				r.Post("/register", s.register)
			})
		})
	})

	go func() {
		ticker := time.Tick(3 * time.Second)
		for range ticker {
			newCfg, err := s.cfgReader.Read()
			if err == nil {
				s.cfg.Store(newCfg)
			}
			log.Printf("Reread config %v", newCfg)
		}
	}()

	err = http.ListenAndServe(":8001", router)
	log.Fatal(err)
	return err
}

func (s Server) register(w http.ResponseWriter, r *http.Request) {

}
